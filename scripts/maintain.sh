#!/bin/bash
# ainfra maintainer command dispatcher.
#
# This keeps the host release workflow memorable while leaving the reviewed
# preparation and execution stages as independently testable tools.

set -eu

repo_root="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd -P)"

usage() {
  printf '%s\n' \
    'usage:' \
    '  scripts/maintain.sh release-host-prepare --version=SEMVER' \
    '  scripts/maintain.sh release-host --version=SEMVER [--dry-run]'
}

require_preparation_tools() {
  missing=""
  for tool in git go; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      missing="${missing}${missing:+, }${tool}"
    fi
  done
  [ -z "$missing" ] || {
    printf 'release-host preparation failed: missing host tools: %s\n' \
      "$missing" >&2
    case "$(uname -s)" in
      Darwin)
        printf '%s\n' 'On macOS, install them with: brew install git go' >&2
        ;;
      Linux)
        printf '%s\n' \
          'On Linux, install Git and Go with your system package manager.' >&2
        ;;
    esac
    exit 1
  }
}

parse_release_options() {
  release_version=""
  dry_run=false
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --version=*) release_version="${1#--version=}" ;;
      --dry-run) dry_run=true ;;
      *)
        usage >&2
        exit 2
        ;;
    esac
    shift
  done
  case "$release_version" in
    "" | *[!0-9A-Za-z.+-]*)
      printf '%s\n' 'release-host failed: invalid or missing version' >&2
      exit 2
      ;;
  esac
}

resolve_prepared_run() {
  version_dir="$repo_root/tmp/container-gate/$release_version"
  [ -d "$version_dir" ] || {
    printf 'release-host failed: no prepared run for version %s\n' \
      "$release_version" >&2
    exit 1
  }
  newest=""
  for candidate in "$version_dir"/*; do
    [ -d "$candidate" ] || continue
    candidate_name="${candidate##*/}"
    case "$candidate_name" in
      [0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]T*Z-*) ;;
      *) continue ;;
    esac
    if [ -z "$newest" ] || [ "$candidate_name" \> "${newest##*/}" ]; then
      newest="$candidate"
    fi
  done
  [ -n "$newest" ] || {
    printf 'release-host failed: no prepared run for version %s\n' \
      "$release_version" >&2
    exit 1
  }
  [ -d "$newest/input" ] || {
    printf '%s\n' 'release-host failed: newest run is incomplete' >&2
    exit 1
  }
  [ ! -e "$newest/runtime" ] && [ ! -e "$newest/evidence" ] || {
    printf '%s\n' 'release-host failed: newest run was already attempted' >&2
    exit 1
  }
  printf '%s\n' "$newest"
}

case "${1:-}" in
  release-host-prepare)
    shift
    parse_release_options "$@"
    [ "$dry_run" = false ] || {
      usage >&2
      exit 2
    }
    require_preparation_tools
    exec "$repo_root/scripts/prepare-container-gate.py" \
      "--version=$release_version"
    ;;
  release-host)
    shift
    parse_release_options "$@"
    run_dir="$(resolve_prepared_run)"
    if [ "$dry_run" = true ]; then
      printf '%s\n' \
        'release-host dry-run: evidence only; no commit, tag, push, or publish' \
        >&2
    fi
    exec "$repo_root/scripts/container-gate-host" "$run_dir"
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
