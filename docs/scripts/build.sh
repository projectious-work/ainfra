#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."
exec hugo --gc --minify --cleanDestinationDir "$@"
