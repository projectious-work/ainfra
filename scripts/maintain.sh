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
    '  scripts/maintain.sh release-host-prepare' \
    '  scripts/maintain.sh release-host [RUN_DIRECTORY] [--dry-run]'
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

case "${1:-}" in
  release-host-prepare)
    [ "$#" -eq 1 ] || {
      usage >&2
      exit 2
    }
    require_preparation_tools
    exec "$repo_root/scripts/prepare-container-gate.py"
    ;;
  release-host)
    shift
    run_dir=""
    dry_run=false
    while [ "$#" -gt 0 ]; do
      case "$1" in
        --dry-run) dry_run=true ;;
        -*)
          usage >&2
          exit 2
          ;;
        *)
          [ -z "$run_dir" ] || {
            usage >&2
            exit 2
          }
          run_dir="$1"
          ;;
      esac
      shift
    done

    if [ -z "$run_dir" ]; then
      require_preparation_tools
      run_dir="$("$repo_root/scripts/prepare-container-gate.py")"
    fi
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
