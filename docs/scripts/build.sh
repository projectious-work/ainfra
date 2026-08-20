#!/usr/bin/env bash
set -euo pipefail

docs_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tailwind_bin="$docs_dir/node_modules/.bin/tailwindcss"

[[ -x "$tailwind_bin" ]] || {
  echo "error: documentation dependencies are missing; run npm install in docs/" >&2
  exit 1
}

build_site() {
  PATH="$docs_dir/node_modules/.bin:$PATH" hugo \
    --source "$docs_dir" \
    --destination "$docs_dir/public" \
    --cleanDestinationDir \
    --gc \
    --minify \
    "$@"
}

# Tailwind consumes Hugo's class inventory from the first pass.
build_site --quiet
build_site "$@"
