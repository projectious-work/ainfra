#!/usr/bin/env bash
set -euo pipefail

release_die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

release_validate_version() {
  [[ "${1:-}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] \
    || release_die "version must be X.Y.Z"
}

release_sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

release_package() {
  local root="$1" version="$2" target="$3" binary="$4"
  release_validate_version "${version}"
  case "${target}" in
    x86_64-unknown-linux-gnu|aarch64-unknown-linux-gnu|\
      x86_64-apple-darwin|aarch64-apple-darwin) ;;
    *) release_die "unsupported release target: ${target}" ;;
  esac
  [[ -x "${binary}" ]] || release_die "binary is not executable: ${binary}"
  [[ -f "${root}/LICENSE" ]] || release_die "LICENSE is required"
  command -v file >/dev/null 2>&1 || release_die "file is required"
  local binary_format
  binary_format="$(file -b "${binary}")"
  case "${target}:${binary_format}" in
    x86_64-unknown-linux-gnu:*ELF*x86-64*|\
      aarch64-unknown-linux-gnu:*ELF*ARM\ aarch64*|\
      x86_64-apple-darwin:*Mach-O*x86_64*|\
      aarch64-apple-darwin:*Mach-O*arm64*) ;;
    *) release_die "binary architecture does not match ${target}" ;;
  esac

  local dist="${root}/dist"
  local name="ainfra-v${version}-${target}"
  local archive="${dist}/${name}.tar.gz"
  local stage
  mkdir -p "${dist}"
  stage="$(mktemp -d)"
  trap 'rm -rf "${stage}"' RETURN
  install -m 0755 "${binary}" "${stage}/ainfra"
  install -m 0644 "${root}/LICENSE" "${stage}/LICENSE"

  # GNU tar is used by the Linux release environment. macOS installs gtar.
  local tar_bin="${AINFRA_TAR:-tar}"
  "${tar_bin}" --sort=name --owner=0 --group=0 --numeric-owner \
    --mtime="@${SOURCE_DATE_EPOCH:-0}" -cf - -C "${stage}" LICENSE ainfra \
    | gzip -n > "${archive}"
  printf '%s\n' "$(release_sha256 "${archive}")" > "${archive}.sha256"
  rm -rf "${stage}"
  trap - RETURN
  printf '%s\n' "${archive}"
}
