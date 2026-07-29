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

release_create_archive() {
  local stage="$1"
  local tar_bin="${AINFRA_TAR:-tar}"
  local epoch="${SOURCE_DATE_EPOCH:-0}"

  if "${tar_bin}" --version 2>/dev/null | grep -Fq 'GNU tar'; then
    "${tar_bin}" --sort=name --owner=0 --group=0 --numeric-owner \
      --mtime="@${epoch}" -cf - -C "${stage}" LICENSE ainfra
    return
  fi

  [[ "$(uname -s)" == Darwin ]] \
    || release_die "deterministic packaging requires GNU tar or macOS tar"
  local timestamp
  # BSD touch interprets -t in local time. Formatting in UTC would shift
  # epoch 0 into 1969 for positive UTC offsets, which USTAR cannot encode.
  timestamp="$(date -r "${epoch}" '+%Y%m%d%H%M.%S')" \
    || release_die "invalid SOURCE_DATE_EPOCH: ${epoch}"
  touch -t "${timestamp}" "${stage}/LICENSE" "${stage}/ainfra"
  COPYFILE_DISABLE=1 "${tar_bin}" --format ustar \
    --uid 0 --gid 0 --uname root --gname root \
    -cf - -C "${stage}" LICENSE ainfra
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
  release_verify_binary_target "${binary}" "${target}"

  local dist="${root}/dist"
  local name="ainfra-v${version}-${target}"
  local archive="${dist}/${name}.tar.gz"
  local stage
  mkdir -p "${dist}"
  stage="$(mktemp -d)"
  trap 'rm -rf "${stage}"' RETURN
  install -m 0755 "${binary}" "${stage}/ainfra"
  install -m 0644 "${root}/LICENSE" "${stage}/LICENSE"

  release_create_archive "${stage}" | gzip -n > "${archive}"
  printf '%s\n' "$(release_sha256 "${archive}")" > "${archive}.sha256"
  rm -rf "${stage}"
  trap - RETURN
  printf '%s\n' "${archive}"
}

release_expected_targets() {
  printf '%s\n' \
    aarch64-unknown-linux-gnu \
    x86_64-unknown-linux-gnu \
    aarch64-apple-darwin \
    x86_64-apple-darwin
}

release_verify_archive() {
  local root="$1" version="$2" target="$3"
  release_validate_version "${version}"
  local archive="${root}/dist/ainfra-v${version}-${target}.tar.gz"
  local checksum="${archive}.sha256" expected actual stage members types
  [[ -f "${archive}" ]] || release_die "missing release archive: ${archive}"
  [[ -f "${checksum}" ]] || release_die "missing checksum: ${checksum}"
  expected="$(sed -n '1p' "${checksum}")"
  [[ "$(wc -l < "${checksum}" | tr -d ' ')" == 1 ]] \
    || release_die "checksum must contain exactly one line: ${checksum}"
  [[ "${expected}" =~ ^[0-9a-f]{64}$ ]] \
    || release_die "checksum must be one lowercase SHA-256 digest: ${checksum}"
  actual="$(release_sha256 "${archive}")"
  [[ "${actual}" == "${expected}" ]] \
    || release_die "checksum verification failed for ${archive}"

  members="$(tar -tzf "${archive}")"
  if [[ "${members}" != $'LICENSE\nainfra' ]]; then
    printf 'archive members were:\n%s\n' "${members}" >&2
    release_die "archive has unexpected members: ${archive}"
  fi
  types="$(tar -tvzf "${archive}" | awk '{print substr($1, 1, 1)}')"
  [[ "${types}" == $'-\n-' ]] \
    || release_die "archive members must be regular files: ${archive}"
  stage="$(mktemp -d)"
  trap 'rm -rf "${stage}"' RETURN
  tar -xzf "${archive}" -C "${stage}"
  [[ -x "${stage}/ainfra" ]] \
    || release_die "archive binary is not executable: ${archive}"
  release_verify_binary_target "${stage}/ainfra" "${target}"
  rm -rf "${stage}"
  trap - RETURN
}

release_verify_binary_target() {
  local binary="$1" target="$2" binary_format
  command -v file >/dev/null 2>&1 || release_die "file is required"
  binary_format="$(file -b "${binary}")"
  case "${target}:${binary_format}" in
    x86_64-unknown-linux-gnu:*ELF*x86-64*|\
      aarch64-unknown-linux-gnu:*ELF*ARM\ aarch64*|\
      x86_64-apple-darwin:*Mach-O*x86_64*|\
      aarch64-apple-darwin:*Mach-O*arm64*) ;;
    *) release_die "binary architecture does not match ${target}" ;;
  esac
}
