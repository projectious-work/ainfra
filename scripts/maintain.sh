#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "${script_dir}/.." && pwd)"
# shellcheck source=scripts/release-lib.sh
source "${script_dir}/release-lib.sh"

usage() {
  cat <<'EOF'
Usage: scripts/maintain.sh COMMAND [ARGUMENTS]

  test                         Run the complete local validation suite
  package VERSION TARGET BIN   Create one deterministic archive and checksum
  release VERSION              Build, verify, tag, and publish Linux artifacts
  release-host VERSION         Build and upload both native macOS artifacts

Publishing requires AINFRA_RELEASE_CONFIRM=vVERSION.
EOF
}

require_clean_release_source() {
  local version="$1" expected_branch="$2"
  [[ -z "$(git -C "${root}" status --porcelain)" ]] \
    || release_die "release requires a clean worktree"
  [[ "$(git -C "${root}" branch --show-current)" == "${expected_branch}" ]] \
    || release_die "release must run on ${expected_branch}"
  git -C "${root}" fetch origin \
    "refs/heads/${expected_branch}:refs/remotes/origin/${expected_branch}"
  [[ "$(git -C "${root}" rev-parse HEAD)" == \
    "$(git -C "${root}" rev-parse "origin/${expected_branch}")" ]] \
    || release_die "checkout must equal origin/${expected_branch}"
  [[ "${AINFRA_RELEASE_CONFIRM:-}" == "v${version}" ]] \
    || release_die "set AINFRA_RELEASE_CONFIRM=v${version} to publish"
}

set_version() {
  local version="$1" current tmp
  current="$(sed -nE 's/^version = "([^"]+)"$/\1/p' "${root}/Cargo.toml" \
    | head -1)"
  [[ -n "${current}" ]] || release_die "cannot read Cargo.toml version"
  [[ "${current}" == "${version}" ]] && return
  tmp="$(mktemp)"
  sed "0,/^version = \"${current}\"$/s//version = \"${version}\"/" \
    "${root}/Cargo.toml" > "${tmp}"
  mv "${tmp}" "${root}/Cargo.toml"
  (cd "${root}" && cargo metadata --format-version 1 --quiet >/dev/null)
  git -C "${root}" add Cargo.toml Cargo.lock
  git -C "${root}" commit -m "chore: release v${version}"
}

build_target() {
  local version="$1" target="$2"
  (
    cd "${root}"
    export CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER=x86_64-linux-gnu-gcc
    export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
    cargo build --locked --release --target "${target}"
  )
  release_package "${root}" "${version}" "${target}" \
    "${root}/target/${target}/release/ainfra"
}

verify_native_archive() {
  local version="$1" target="$2"
  local archive="${root}/dist/ainfra-v${version}-${target}.tar.gz"
  local checksum="${archive}.sha256" install_root
  [[ "$(release_sha256 "${archive}")" == "$(sed -n '1p' "${checksum}")" ]] \
    || release_die "checksum verification failed for ${archive}"
  install_root="$(mktemp -d)"
  tar -xzf "${archive}" -C "${install_root}"
  "${install_root}/ainfra" --version | grep -Fqx "ainfra ${version}"
  "${install_root}/ainfra" --help >/dev/null
  rm -rf "${install_root}"
}

release_linux() {
  local version="$1" target
  release_validate_version "${version}"
  require_clean_release_source "${version}" v0.x-release
  set_version "${version}"
  (cd "${root}" && "${script_dir}/validate-all")
  for target in aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu; do
    build_target "${version}" "${target}"
  done
  case "$(uname -m)" in
    x86_64|amd64) target=x86_64-unknown-linux-gnu ;;
    aarch64|arm64) target=aarch64-unknown-linux-gnu ;;
    *) release_die "unsupported native Linux architecture" ;;
  esac
  verify_native_archive "${version}" "${target}"
  git -C "${root}" tag -a "v${version}" -m "Release v${version}"
  git -C "${root}" push origin v0.x-release "v${version}"
  gh release create "v${version}" --repo projectious-work/ainfra \
    --generate-notes "${root}/install.sh" \
    "${root}/dist/ainfra-v${version}-"*linux-gnu.tar.gz*
}

release_host() {
  local version="$1" target tag_commit
  release_validate_version "${version}"
  [[ "$(uname -s)" == Darwin ]] || release_die "release-host requires macOS"
  require_clean_release_source "${version}" v0.x-release
  tag_commit="$(git -C "${root}" rev-parse "v${version}^{commit}")"
  [[ "${tag_commit}" == "$(git -C "${root}" rev-parse HEAD)" ]] \
    || release_die "HEAD must be the exact v${version} source"
  command -v gtar >/dev/null || release_die "GNU tar (gtar) is required"
  export AINFRA_TAR=gtar
  for target in aarch64-apple-darwin x86_64-apple-darwin; do
    build_target "${version}" "${target}"
    verify_native_archive "${version}" "${target}"
  done
  gh release upload "v${version}" --repo projectious-work/ainfra \
    "${root}/dist/ainfra-v${version}-"*apple-darwin.tar.gz*
}

command="${1:-help}"
shift || true
case "${command}" in
  test) (cd "${root}" && exec "${script_dir}/validate-all") ;;
  package)
    [[ "$#" -eq 3 ]] || release_die "package requires VERSION TARGET BIN"
    release_package "${root}" "$1" "$2" "$3"
    ;;
  release)
    [[ "$#" -eq 1 ]] || release_die "release requires VERSION"
    release_linux "$1"
    ;;
  release-host)
    [[ "$#" -eq 1 ]] || release_die "release-host requires VERSION"
    release_host "$1"
    ;;
  help|-h|--help) usage ;;
  *) usage >&2; exit 2 ;;
esac
