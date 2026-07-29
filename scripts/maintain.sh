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
  audit-release VERSION        Verify all local release artifacts
  release VERSION              Run phase 0, build, tag, and publish Linux
  release VERSION --steps phase0
                               Write candidate-bound readiness reports only
  release VERSION --steps checks
                               Run or reuse candidate-bound validation logs
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

require_clean_linked_worktrees() {
  local worktree
  while IFS= read -r worktree; do
    [[ -z "$(git -C "${worktree}" status --porcelain)" ]] \
      || release_die "linked worktree is dirty: ${worktree}"
  done < <(
    git -C "${root}" worktree list --porcelain \
      | sed -n 's/^worktree //p'
  )
}

set_version() {
  local version="$1" current
  current="$(sed -nE 's/^version = "([^"]+)"$/\1/p' "${root}/Cargo.toml" \
    | head -1)"
  [[ -n "${current}" ]] || release_die "cannot read Cargo.toml version"
  [[ "${current}" == "${version}" ]] \
    || release_die "version ${version} must be prepared and merged by PR"
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
  local install_root
  release_verify_archive "${root}" "${version}" "${target}"
  install_root="$(mktemp -d)"
  tar -xzf "${archive}" -C "${install_root}"
  "${install_root}/ainfra" --version | grep -Fqx "ainfra ${version}"
  "${install_root}/ainfra" --help >/dev/null
  rm -rf "${install_root}"
}

audit_release() {
  local version="$1"
  shift
  release_validate_version "${version}"
  local targets=("$@") target
  if [[ "${#targets[@]}" -eq 0 ]]; then
    mapfile -t targets < <(release_expected_targets)
  fi
  for target in "${targets[@]}"; do
    release_verify_archive "${root}" "${version}" "${target}"
  done
  printf 'verified %s release artifact(s) for v%s\n' \
    "${#targets[@]}" "${version}"
}

verify_remote_release() {
  local version="$1" asset target
  local assets
  assets="$(gh release view "v${version}" --repo projectious-work/ainfra \
    --json assets --jq '.assets[].name')"
  for target in $(release_expected_targets); do
    for asset in \
      "ainfra-v${version}-${target}.tar.gz" \
      "ainfra-v${version}-${target}.tar.gz.sha256" \
      "ainfra-v${version}-${target}.provenance.json"; do
      grep -Fqx "${asset}" <<< "${assets}" \
        || release_die "GitHub release is missing ${asset}"
    done
  done
  grep -Fqx install.sh <<< "${assets}" \
    || release_die "GitHub release is missing install.sh"
  grep -Fqx LICENSE <<< "${assets}" \
    || release_die "GitHub release is missing LICENSE"
  grep -Fqx "ainfra-v${version}.spdx.json" <<< "${assets}" \
    || release_die "GitHub release is missing SPDX SBOM"
}

generate_release_metadata() {
  local version="$1"
  shift
  local arguments=() target
  for target in "$@"; do
    arguments+=(--target "${target}")
  done
  (cd "${root}" && uv run --frozen python \
    scripts/generate-release-metadata.py \
    --root "${root}" --version "${version}" "${arguments[@]}")
  cp "${root}/LICENSE" "${root}/dist/LICENSE"
}

release_linux() {
  local version="$1" target
  release_validate_version "${version}"
  require_clean_release_source "${version}" v0.x-release
  set_version "${version}"
  require_clean_linked_worktrees
  "${script_dir}/release-evidence.sh" phase0 "${version}"
  "${script_dir}/release-checks.sh" run "${version}"
  for target in aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu; do
    build_target "${version}" "${target}"
  done
  audit_release "${version}" \
    aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu
  generate_release_metadata "${version}" \
    aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu
  case "$(uname -m)" in
    x86_64|amd64) target=x86_64-unknown-linux-gnu ;;
    aarch64|arm64) target=aarch64-unknown-linux-gnu ;;
    *) release_die "unsupported native Linux architecture" ;;
  esac
  verify_native_archive "${version}" "${target}"
  if git -C "${root}" rev-parse --verify --quiet \
    "refs/tags/v${version}" >/dev/null; then
    release_die "tag v${version} already exists"
  fi
  git -C "${root}" tag -a "v${version}" -m "Release v${version}"
  git -C "${root}" push origin v0.x-release "v${version}"
  gh release create "v${version}" --repo projectious-work/ainfra \
    --notes-file "${root}/release-notes/v${version}.md" \
    "${root}/install.sh" "${root}/dist/LICENSE" \
    "${root}/dist/ainfra-v${version}.spdx.json" \
    "${root}/dist/ainfra-v${version}-"*linux-gnu.tar.gz* \
    "${root}/dist/ainfra-v${version}-"*linux-gnu.provenance.json
}

release_host() {
  local version="$1" target tag_commit
  release_validate_version "${version}"
  [[ "$(uname -s)" == Darwin ]] || release_die "release-host requires macOS"
  require_clean_release_source "${version}" v0.x-release
  require_clean_linked_worktrees
  tag_commit="$(git -C "${root}" rev-parse "v${version}^{commit}")"
  git -C "${root}" merge-base --is-ancestor "${tag_commit}" HEAD \
    || release_die "v${version} is not an ancestor of v0.x-release"
  [[ "$(git -C "${root}" show \
    "v${version}:Cargo.toml" \
    | sed -nE 's/^version = "([^"]+)"$/\1/p' \
    | head -1)" == "${version}" ]] \
    || release_die "tag v${version} has different Cargo metadata"
  git -C "${root}" diff --quiet "${tag_commit}" HEAD -- \
    Cargo.toml Cargo.lock LICENSE install.sh schemas src templates \
    || release_die "release inputs differ from v${version}"
  for target in aarch64-apple-darwin x86_64-apple-darwin; do
    build_target "${version}" "${target}"
    verify_native_archive "${version}" "${target}"
  done
  audit_release "${version}" \
    aarch64-apple-darwin x86_64-apple-darwin
  gh release download "v${version}" --repo projectious-work/ainfra \
    --dir "${root}/dist" --clobber \
    --pattern "ainfra-v${version}-*linux-gnu.tar.gz" \
    --pattern "ainfra-v${version}-*linux-gnu.tar.gz.sha256"
  audit_release "${version}" \
    aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu
  AINFRA_RELEASE_SOURCE_COMMIT="${tag_commit}" \
    generate_release_metadata "${version}" \
    aarch64-unknown-linux-gnu x86_64-unknown-linux-gnu \
    aarch64-apple-darwin x86_64-apple-darwin
  gh release upload "v${version}" --repo projectious-work/ainfra \
    --clobber "${root}/dist/LICENSE" \
    "${root}/dist/ainfra-v${version}.spdx.json" \
    "${root}/dist/ainfra-v${version}-"*apple-darwin.tar.gz* \
    "${root}/dist/ainfra-v${version}-"*.provenance.json
  verify_remote_release "${version}"
}

command="${1:-help}"
shift || true
case "${command}" in
  test) (cd "${root}" && exec "${script_dir}/validate-all") ;;
  package)
    [[ "$#" -eq 3 ]] || release_die "package requires VERSION TARGET BIN"
    release_package "${root}" "$1" "$2" "$3"
    ;;
  audit-release)
    [[ "$#" -ge 1 ]] || release_die "audit-release requires VERSION"
    audit_release "$@"
    ;;
  release)
    [[ "$#" -ge 1 ]] || release_die "release requires VERSION"
    version="$1"
    shift
    if [[ "$#" -eq 2 && "$1" == --steps && "$2" == phase0 ]]; then
      release_phase0_version="${version}"
      require_clean_release_source "${release_phase0_version}" v0.x-release
      set_version "${release_phase0_version}"
      "${script_dir}/release-evidence.sh" phase0 \
        "${release_phase0_version}"
    elif [[ "$#" -eq 2 && "$1" == --steps && "$2" == checks ]]; then
      release_checks_version="${version}"
      require_clean_release_source "${release_checks_version}" v0.x-release
      set_version "${release_checks_version}"
      "${script_dir}/release-checks.sh" run \
        "${release_checks_version}"
    elif [[ "$#" -eq 0 ]]; then
      release_linux "${version}"
    else
      release_die "supported release steps: phase0, checks"
    fi
    ;;
  release-host)
    [[ "$#" -eq 1 ]] || release_die "release-host requires VERSION"
    release_host "$1"
    ;;
  help|-h|--help) usage ;;
  *) usage >&2; exit 2 ;;
esac
