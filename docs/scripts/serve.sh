#!/usr/bin/env bash
set -euo pipefail

docs_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tailwind_bin="$docs_dir/node_modules/.bin/tailwindcss"

[[ -x "$tailwind_bin" ]] || {
  echo "error: documentation dependencies are missing; run npm install in docs/" >&2
  exit 1
}

exec env PATH="$docs_dir/node_modules/.bin:$PATH" hugo server \
  --source "$docs_dir" \
  --baseURL http://localhost:1314/ainfra/ \
  --bind 0.0.0.0 \
  --port 1314 \
  --buildDrafts \
  --disableFastRender \
  --navigateToChanged \
  "$@"
