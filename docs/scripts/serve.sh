#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."
exec hugo server \
  --baseURL http://localhost:1314/ainfra/ \
  --bind 127.0.0.1 \
  --port 1314 \
  --buildDrafts \
  --disableFastRender \
  "$@"
