#!/bin/sh
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

destination=${1:-dist/phase-1}

build_target() {
  target_os=$1
  target_arch=$2
  output="$destination/$target_os/$target_arch/ainfra"
  mkdir -p "$(dirname -- "$output")"
  printf '+ CGO_ENABLED=0 GOOS=%s GOARCH=%s go build -trimpath -o %s ./cmd/ainfra\n' \
    "$target_os" "$target_arch" "$output"
  CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch \
    go build -trimpath -o "$output" ./cmd/ainfra
}

build_target linux amd64
build_target linux arm64
build_target darwin amd64
build_target darwin arm64

printf 'built Linux and macOS binaries under %s\n' "$destination"
