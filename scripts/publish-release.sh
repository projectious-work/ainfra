#!/usr/bin/env bash
# Publish an already built, smoke-tested, and Sigstore-signed release. The
# script deliberately performs no compilation: publication transfers the exact
# artifacts validated by the container and host gates.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
identity='info@projectious.work'
issuer='https://github.com/login/oauth'
version=''
dry_run=false

die() {
  printf 'release publication failed: %s\n' "$*" >&2
  exit 1
}

usage() {
  printf '%s\n' \
    'usage: scripts/publish-release.sh --version=SEMVER [--dry-run]'
}

while (($# > 0)); do
  case "$1" in
    --version=*) version="${1#--version=}" ;;
    --dry-run) dry_run=true ;;
    *) usage >&2; exit 2 ;;
  esac
  shift
done

version="${version#v}"
[[ "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]] ||
  die "invalid semantic version: $version"
tag="v$version"

cd "$repo_root"
for tool in cosign gh git; do
  command -v "$tool" >/dev/null 2>&1 || die "missing required tool: $tool"
done
if command -v sha256sum >/dev/null 2>&1; then
  checksum_command=(sha256sum --check)
elif command -v shasum >/dev/null 2>&1; then
  checksum_command=(shasum -a 256 --check)
else
  die 'missing checksum tool: install sha256sum or shasum'
fi

[[ -z "$(git status --porcelain)" ]] || die 'worktree is not clean'
head_commit="$(git rev-parse HEAD)"
release_commit="$(git rev-parse origin/v1.x-release)"
[[ "$head_commit" == "$release_commit" ]] ||
  die 'HEAD is not the exact origin/v1.x-release commit'

release_dir="$repo_root/dist/release/$version"
manifest="$release_dir/checksums.sha256"
bundle="$release_dir/checksums.sha256.sigstore.json"
notes="$repo_root/docs/releases/$tag.md"
[[ -d "$release_dir" && ! -L "$release_dir" ]] ||
  die "missing or unsafe release directory: $release_dir"
[[ -s "$manifest" && ! -L "$manifest" ]] || die 'missing checksum manifest'
[[ -s "$bundle" && ! -L "$bundle" ]] || die 'missing Sigstore bundle'
[[ -s "$notes" && ! -L "$notes" ]] || die 'missing release notes'

(
  cd "$release_dir"
  "${checksum_command[@]}" checksums.sha256
)
cosign verify-blob "$manifest" \
  --bundle "$bundle" \
  --certificate-identity "$identity" \
  --certificate-oidc-issuer "$issuer"

assets=(
  "$release_dir/ainfra_${version}_darwin_amd64.spdx.json"
  "$release_dir/ainfra_${version}_darwin_amd64.tar.gz"
  "$release_dir/ainfra_${version}_darwin_arm64.spdx.json"
  "$release_dir/ainfra_${version}_darwin_arm64.tar.gz"
  "$release_dir/ainfra_${version}_linux_amd64.spdx.json"
  "$release_dir/ainfra_${version}_linux_amd64.tar.gz"
  "$release_dir/ainfra_${version}_linux_arm64.spdx.json"
  "$release_dir/ainfra_${version}_linux_arm64.tar.gz"
  "$manifest"
  "$bundle"
)
for asset in "${assets[@]}"; do
  [[ -s "$asset" && ! -L "$asset" ]] || die "missing release asset: $asset"
done

if [[ "$dry_run" == true ]]; then
  printf '%s\n' \
    "release-publish dry-run complete for $tag at $head_commit" \
    'no tag, push, GitHub release, or asset upload performed'
  exit 0
fi

[[ -z "$(git tag --list "$tag")" ]] || die "local tag already exists: $tag"
[[ -z "$(git ls-remote --tags origin "refs/tags/$tag")" ]] ||
  die "remote tag already exists: $tag"
if gh release view "$tag" >/dev/null 2>&1; then
  die "GitHub release already exists: $tag"
fi

git tag --annotate "$tag" --message "ainfra $tag" "$head_commit"
git push origin "refs/tags/$tag"
gh release create "$tag" "${assets[@]}" \
  --verify-tag \
  --prerelease \
  --title "$tag" \
  --notes-file "$notes"

verify_dir=$(mktemp -d "${TMPDIR:-/tmp}/ainfra-release-verify.XXXXXX")
cleanup_verify() {
  rm -rf -- "$verify_dir"
}
trap cleanup_verify EXIT
gh release download "$tag" --dir "$verify_dir"
(
  cd "$verify_dir"
  "${checksum_command[@]}" checksums.sha256
  cosign verify-blob checksums.sha256 \
    --bundle checksums.sha256.sigstore.json \
    --certificate-identity "$identity" \
    --certificate-oidc-issuer "$issuer"
)
cleanup_verify
trap - EXIT

printf 'release published and independently verified: %s\n' "$tag"
