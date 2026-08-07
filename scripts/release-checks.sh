#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

die() {
  printf 'release check failed: %s\n' "$*" >&2
  exit 1
}

validate_version() {
  [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] ||
    die "invalid semantic version: $1"
}

release_checks() {
  local version=$1
  validate_version "${version}"

  [[ -z "$(git status --porcelain --untracked-files=no)" ]] ||
    die "tracked worktree is not clean"

  scripts/validate-all
  go test -race ./...
  go test -coverprofile=coverage.out ./...
  scripts/build-targets.sh "dist/v${version}"
  go tool govulncheck ./...
  go tool gosec ./...

  command -v gitleaks >/dev/null || die "gitleaks is required"
  gitleaks git --redact

  printf 'release checks passed for v%s\n' "${version}"
}

case "${1:-}" in
  run)
    [[ $# -eq 2 ]] || die "usage: scripts/release-checks.sh run VERSION"
    release_checks "$2"
    ;;
  *)
    die "usage: scripts/release-checks.sh run VERSION"
    ;;
esac
