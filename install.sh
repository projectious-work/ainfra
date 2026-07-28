#!/bin/sh
set -eu

version="${AINFRA_VERSION:-latest}"
install_dir="${AINFRA_INSTALL_DIR:-${HOME}/.local/bin}"
repo=projectious-work/ainfra
base_url="https://github.com/${repo}/releases"

fail() {
  printf 'ainfra installer: %s\n' "$*" >&2
  exit 1
}

valid_version() {
  printf '%s\n' "$1" | awk -F. '
    NF == 3 &&
    $1 ~ /^(0|[1-9][0-9]*)$/ &&
    $2 ~ /^(0|[1-9][0-9]*)$/ &&
    $3 ~ /^(0|[1-9][0-9]*)$/ { valid = 1 }
    END { exit valid ? 0 : 1 }
  '
}

case "${version}" in
  latest) ;;
  *) valid_version "${version}" || fail "version must be X.Y.Z or latest" ;;
esac

[ "$(id -u)" -ne 0 ] || fail "refusing to install as root"
case "$(uname -s):$(uname -m)" in
  Linux:x86_64|Linux:amd64) target=x86_64-unknown-linux-gnu ;;
  Linux:aarch64|Linux:arm64) target=aarch64-unknown-linux-gnu ;;
  Darwin:x86_64) target=x86_64-apple-darwin ;;
  Darwin:arm64|Darwin:aarch64) target=aarch64-apple-darwin ;;
  *) fail "unsupported platform: $(uname -s) $(uname -m)" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"
if [ "${version}" = latest ]; then
  version="$(curl --fail --silent --show-error --proto '=https' --tlsv1.2 \
    "https://api.github.com/repos/${repo}/releases/latest" \
    | sed -nE 's/.*"tag_name"[[:space:]]*:[[:space:]]*"v([^"]+)".*/\1/p')"
  valid_version "${version}" || fail "could not resolve the latest stable version"
fi
release_path="download/v${version}"
if command -v sha256sum >/dev/null 2>&1; then
  checksum_file() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
  checksum_file() { shasum -a 256 "$1" | awk '{print $1}'; }
else
  fail "sha256sum or shasum is required"
fi

name="ainfra-v${version}-${target}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT HUP INT TERM
url="${base_url}/${release_path}/${name}"
curl --fail --location --proto '=https' --tlsv1.2 \
  --proto-redir '=https' \
  --output "${tmp}/${name}" "${url}"
curl --fail --location --proto '=https' --tlsv1.2 \
  --proto-redir '=https' \
  --output "${tmp}/${name}.sha256" "${url}.sha256"

expected="$(awk \
  'NF == 1 && length($1) == 64 && $1 ~ /^[0-9a-fA-F]+$/ { print $1 }' \
  "${tmp}/${name}.sha256")"
[ "$(printf '%s\n' "${expected}" | sed '/^$/d' | wc -l | tr -d ' ')" -eq 1 ] \
  || fail "checksum file must contain exactly one canonical entry"
actual="$(checksum_file "${tmp}/${name}")"
[ "${actual}" = "${expected}" ] || fail "checksum verification failed"

members="$(tar -tzf "${tmp}/${name}")"
[ "${members}" = "$(printf 'LICENSE\nainfra')" ] \
  || fail "archive contains unexpected members"
types="$(tar -tvzf "${tmp}/${name}" | awk '{print substr($1, 1, 1)}')"
[ "${types}" = "$(printf '%s\n%s' - -)" ] \
  || fail "archive members must be regular files"
tar -xzf "${tmp}/${name}" -C "${tmp}"
[ -f "${tmp}/ainfra" ] && [ ! -L "${tmp}/ainfra" ] \
  || fail "archive does not contain a regular ainfra binary"
chmod 0755 "${tmp}/ainfra"
mkdir -p "${install_dir}"
[ ! -L "${install_dir}" ] || fail "install directory must not be a symlink"
install -m 0755 "${tmp}/ainfra" "${install_dir}/.ainfra.new"
mv -f "${install_dir}/.ainfra.new" "${install_dir}/ainfra"
"${install_dir}/ainfra" --version
"${install_dir}/ainfra" --help >/dev/null
printf 'Installed ainfra to %s/ainfra\n' "${install_dir}"
case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *) printf 'Add %s to PATH.\n' "${install_dir}" ;;
esac
