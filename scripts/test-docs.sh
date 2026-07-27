#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "${BUILD_DIR}"' EXIT

"${ROOT_DIR}/scripts/build-docs.sh" --destination "${BUILD_DIR}"

test -f "${BUILD_DIR}/index.html"
test -f "${BUILD_DIR}/docs/index.html"
test -f "${BUILD_DIR}/docs/getting-started/quickstart/index.html"
test -f "${BUILD_DIR}/favicon.svg"

if rg -n 'ainfra-templates/ainfra-templates/' "${BUILD_DIR}"; then
  echo "Documentation contains a duplicated base path." >&2
  exit 1
fi

echo "Documentation build and smoke checks passed."
