#!/usr/bin/env bash
# Build the complete cross-platform release payload in the development
# container. Host-side signing consumes this script's immutable output; it
# never recompiles or modifies an archive.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
version=""
dry_run=false

die() {
  printf 'release packaging failed: %s\n' "$*" >&2
  exit 1
}

usage() {
  printf '%s\n' \
    'usage: scripts/package-release.sh --version=SEMVER [--dry-run]'
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

cd "$repo_root"

for tool in git go tar gzip syft sha256sum; do
  command -v "$tool" >/dev/null 2>&1 || die \
    "missing required devcontainer tool: $tool; run release-package inside the ainfra devcontainer"
done

release_root="$repo_root/dist/release"
destination="$release_root/$version"

if [[ "$dry_run" == true ]]; then
  printf '%s\n' \
    "release-package dry-run: would create $destination" \
    'targets: linux/amd64 linux/arm64 darwin/amd64 darwin/arm64' \
    'outputs: four archives, four SPDX JSON SBOMs, checksums.sha256'
  exit 0
fi

[[ -z "$(git status --porcelain)" ]] || die 'worktree is not clean'
[[ ! -e "$destination" ]] || die "release directory already exists: $destination"

mkdir -p "$release_root"
staging="$release_root/.${version}.packaging.$$"
[[ ! -e "$staging" ]] || die "staging directory already exists: $staging"
mkdir -m 0700 "$staging"
mkdir -m 0700 "$staging/.go-build-cache"
mkdir -m 0700 "$staging/.go-module-cache"
mkdir -m 0700 "$staging/.syft-cache"

cleanup() {
  if [[ -d "$staging" ]]; then
    chmod -R u+w -- "$staging" 2>/dev/null || true
    rm -rf -- "$staging"
  fi
}
trap cleanup EXIT

source_epoch="$(git show -s --format=%ct HEAD)"

package_target() {
  local target_os=$1
  local target_arch=$2
  local base="ainfra_${version}_${target_os}_${target_arch}"
  local package_root="$staging/package/$base"
  local archive="$staging/$base.tar.gz"
  local sbom="$staging/$base.spdx.json"

  mkdir -p "$package_root"
  printf '+ build %s/%s\n' "$target_os" "$target_arch"
  CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" \
    GOCACHE="$staging/.go-build-cache" \
    GOMODCACHE="$staging/.go-module-cache" \
    go build -trimpath -buildvcs=true \
    -ldflags "-X main.injectedVersion=$version" \
    -o "$package_root/ainfra" ./cmd/ainfra
  cp LICENSE README.md "$package_root/"

  printf '+ archive %s\n' "${archive##*/}"
  tar --sort=name \
    --mtime="@$source_epoch" \
    --owner=0 --group=0 --numeric-owner \
    -C "$staging/package" -cf - "$base" |
    gzip -n >"$archive"

  printf '+ sbom %s\n' "${sbom##*/}"
  SYFT_CHECK_FOR_APP_UPDATE=false \
    SYFT_CACHE_DIR="$staging/.syft-cache" \
    syft "$archive" -o "spdx-json=$sbom"
  rm -rf -- "$package_root"
}

package_target linux amd64
package_target linux arm64
package_target darwin amd64
package_target darwin arm64
rmdir "$staging/package"
rm -rf -- "$staging/.go-build-cache"
chmod -R u+w -- "$staging/.go-module-cache"
rm -rf -- "$staging/.go-module-cache"
rm -rf -- "$staging/.syft-cache"

(
  cd "$staging"
  sha256sum \
    "ainfra_${version}_darwin_amd64.tar.gz" \
    "ainfra_${version}_darwin_amd64.spdx.json" \
    "ainfra_${version}_darwin_arm64.tar.gz" \
    "ainfra_${version}_darwin_arm64.spdx.json" \
    "ainfra_${version}_linux_amd64.tar.gz" \
    "ainfra_${version}_linux_amd64.spdx.json" \
    "ainfra_${version}_linux_arm64.tar.gz" \
    "ainfra_${version}_linux_arm64.spdx.json" \
    >checksums.sha256
  sha256sum --check checksums.sha256
)

mv "$staging" "$destination"
trap - EXIT

printf '%s\n' \
  "release packaging complete: $destination" \
  "source commit: $(git rev-parse HEAD)"
