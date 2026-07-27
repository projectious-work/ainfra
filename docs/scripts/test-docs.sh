#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$(mktemp -d)"
ARCHIVE_DIR="$(mktemp -d)"
trap 'rm -rf "${BUILD_DIR}" "${ARCHIVE_DIR}"' EXIT

"${ROOT_DIR}/scripts/build-docs.sh" --destination "${BUILD_DIR}"

test -f "${BUILD_DIR}/index.html"
test -f "${BUILD_DIR}/docs/index.html"
test -f "${BUILD_DIR}/docs/getting-started/quickstart/index.html"
test -f "${BUILD_DIR}/docs/concepts/template-strategy/index.html"
test -f "${BUILD_DIR}/docs/guides/authoring-templates/index.html"
test -f "${BUILD_DIR}/docs/reference/templates/index.html"
test -f "${BUILD_DIR}/favicon.svg"
rg -q '>Releases<' "${BUILD_DIR}/index.html"
rg -q 'dropdown-item-latest' "${BUILD_DIR}/index.html"
rg -q '>main</a>' "${BUILD_DIR}/index.html"

if rg -n 'ainfra-templates/ainfra-templates/' "${BUILD_DIR}"; then
  echo "Documentation contains a duplicated base path." >&2
  exit 1
fi

DOCS_VERSION=v0.test \
DOCS_BASE_URL="https://projectious-work.github.io/ainfra-templates/v0.test/" \
  "${ROOT_DIR}/scripts/build-docs.sh" --destination "${ARCHIVE_DIR}"

rg -q 'Version v0.test' "${ARCHIVE_DIR}/docs/index.html"
rg -q 'archived snapshot' "${ARCHIVE_DIR}/docs/index.html"
rg -q 'ainfra-templates/v0.test/' "${ARCHIVE_DIR}/index.html"

echo "Documentation build and smoke checks passed."
