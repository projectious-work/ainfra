#!/usr/bin/env bash
# Install a released ainfra binary after verifying its published checksum.

set -euo pipefail

repository='projectious-work/ainfra'
release_origin="${AINFRA_RELEASE_ORIGIN:-https://github.com/$repository}"
release_api="${AINFRA_RELEASE_API:-https://api.github.com/repos/$repository/releases}"
version="${VERSION:-}"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
verify_signature="${VERIFY_SIGNATURE:-auto}"

die() {
  printf 'ainfra installer: %s\n' "$*" >&2
  exit 1
}

for tool in curl tar awk grep mktemp install; do
  command -v "$tool" >/dev/null 2>&1 || die "missing required tool: $tool"
done

case "$verify_signature" in
  auto|0|1) ;;
  *) die 'VERIFY_SIGNATURE must be auto, 0, or 1' ;;
esac

case "$(uname -s)" in
  Linux) target_os=linux ;;
  Darwin) target_os=darwin ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) target_arch=amd64 ;;
  arm64|aarch64) target_arch=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

if [[ -z "$version" ]]; then
  release_data=$(curl -fsSL "$release_api?per_page=20") ||
    die 'cannot query published releases'
  version=$(awk -F '"' '/"tag_name":/ { print $4; exit }' <<<"$release_data")
  [[ -n "$version" ]] || die 'cannot resolve the newest published release'
fi
version=${version#v}
[[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]] ||
  die "invalid release version: $version"

base="ainfra_${version}_${target_os}_${target_arch}"
archive="$base.tar.gz"
download_url="$release_origin/releases/download/v$version"
temporary=$(mktemp -d "${TMPDIR:-/tmp}/ainfra-install.XXXXXX")
cleanup() {
  rm -rf -- "$temporary"
}
trap cleanup EXIT

printf 'Downloading ainfra v%s for %s/%s...\n' \
  "$version" "$target_os" "$target_arch"
curl -fsSL --proto '=https' --proto-redir '=https' \
  "$download_url/$archive" -o "$temporary/$archive"
curl -fsSL --proto '=https' --proto-redir '=https' \
  "$download_url/checksums.sha256" -o "$temporary/checksums.sha256"

grep -F "  $archive" "$temporary/checksums.sha256" \
  >"$temporary/selected.sha256" || die 'archive is absent from checksum manifest'
[[ $(awk 'END { print NR }' "$temporary/selected.sha256") == 1 ]] ||
  die 'checksum manifest contains an ambiguous archive entry'

if command -v sha256sum >/dev/null 2>&1; then
  (cd "$temporary" && sha256sum --check selected.sha256)
elif command -v shasum >/dev/null 2>&1; then
  (cd "$temporary" && shasum -a 256 --check selected.sha256)
else
  die 'missing checksum tool: install sha256sum or shasum'
fi

if [[ "$verify_signature" == 1 ]] ||
  [[ "$verify_signature" == auto ]] && command -v cosign >/dev/null 2>&1; then
  command -v cosign >/dev/null 2>&1 ||
    die 'Cosign is required when VERIFY_SIGNATURE=1'
  curl -fsSL --proto '=https' --proto-redir '=https' \
    "$download_url/checksums.sha256.sigstore.json" \
    -o "$temporary/checksums.sha256.sigstore.json"
  cosign verify-blob "$temporary/checksums.sha256" \
    --bundle "$temporary/checksums.sha256.sigstore.json" \
    --certificate-identity 'info@projectious.work' \
    --certificate-oidc-issuer 'https://github.com/login/oauth'
  printf '%s\n' 'Verified the signed checksum manifest with Cosign.'
elif [[ "$verify_signature" == auto ]]; then
  printf '%s\n' \
    'Cosign is not installed; verified the archive checksum only.' \
    'Install Cosign and set VERIFY_SIGNATURE=1 for identity verification.'
fi

mkdir -p "$install_dir"
tar -xzf "$temporary/$archive" -C "$temporary" "$base/ainfra"
install -m 0755 "$temporary/$base/ainfra" "$install_dir/ainfra"

printf 'Installed ainfra v%s to %s/ainfra\n' "$version" "$install_dir"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add %s to PATH before running ainfra.\n' "$install_dir" ;;
esac
