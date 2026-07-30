#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$(mktemp -d)"
ARCHIVE_DIR="$(mktemp -d)"
trap 'rm -rf "${BUILD_DIR}" "${ARCHIVE_DIR}"' EXIT
EXPECTED_HEXTRA_COMMIT="8e53e7a7ef3e24348edbac276e03cb9fecf66d0c"
EXPECTED_FLEXSEARCH_SHA256="433e941a8a573ebb9931fc16fc75266ab6b93f569ac2fb4f3dc66882e0416f4c"

ACTUAL_HEXTRA_COMMIT="$(
  git -C "${ROOT_DIR}/themes/hextra" rev-parse HEAD
)"
if [[ "${ACTUAL_HEXTRA_COMMIT}" != "${EXPECTED_HEXTRA_COMMIT}" ]]; then
  echo "Hextra is not pinned to the approved v0.12.3 commit." >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_FLEXSEARCH_SHA256="$(
    sha256sum \
      "${ROOT_DIR}/assets/js/vendor/flexsearch.bundle.min.js" |
      awk '{print $1}'
  )"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL_FLEXSEARCH_SHA256="$(
    shasum -a 256 \
      "${ROOT_DIR}/assets/js/vendor/flexsearch.bundle.min.js" |
      awk '{print $1}'
  )"
else
  echo "sha256sum or shasum is required for documentation tests." >&2
  exit 1
fi

if [[ "${ACTUAL_FLEXSEARCH_SHA256}" != "${EXPECTED_FLEXSEARCH_SHA256}" ]]; then
  echo "The vendored FlexSearch bundle checksum is invalid." >&2
  exit 1
fi

"${ROOT_DIR}/scripts/build-docs.sh" --destination "${BUILD_DIR}"

test -f "${BUILD_DIR}/index.html"
test -f "${BUILD_DIR}/docs/index.html"
test -f "${BUILD_DIR}/docs/getting-started/quickstart/index.html"
test -f "${BUILD_DIR}/docs/concepts/template-strategy/index.html"
test -f \
  "${BUILD_DIR}/docs/how-to/hetzner-baseline-environment/index.html"
test -f "${BUILD_DIR}/docs/guides/authoring-templates/index.html"
test -f "${BUILD_DIR}/docs/reference/templates/index.html"
test -f "${BUILD_DIR}/favicon.svg"
test -f "${BUILD_DIR}/favicon-dark.svg"
test -f "${BUILD_DIR}/site.webmanifest"
test -f "${BUILD_DIR}/en.search-data.json"
compgen -G "${BUILD_DIR}/en.search.*.js" >/dev/null

rg -q 'aria-label=Releases' "${BUILD_DIR}/index.html"
rg -q 'hextra-nav-menu-items' "${BUILD_DIR}/index.html"
rg -q 'main \(latest\)' "${BUILD_DIR}/index.html"
rg -q 'projectious-work.github.io/ainfra/v0.1/' \
  "${BUILD_DIR}/index.html"
rg -q 'hextra-search-input' "${BUILD_DIR}/index.html"
rg -q 'logo/ainfra-light.svg' "${BUILD_DIR}/index.html"
rg -q 'logo/ainfra-dark.svg' "${BUILD_DIR}/index.html"

if rg -n 'ainfra/ainfra/' "${BUILD_DIR}"; then
  echo "Documentation contains a duplicated base path." >&2
  exit 1
fi

if rg -n -i \
  'docsy|data-bs-|fontawesome|jquery|lunr|class=td-' \
  "${BUILD_DIR}" --glob '*.html' --glob '*.css' --glob '*.js'; then
  echo "Documentation contains obsolete Docsy assets or markup." >&2
  exit 1
fi

DOCS_VERSION=v0.test \
DOCS_BASE_URL="https://projectious-work.github.io/ainfra/v0.test/" \
  "${ROOT_DIR}/scripts/build-docs.sh" --destination "${ARCHIVE_DIR}"

rg -q 'Version v0.test' "${ARCHIVE_DIR}/docs/index.html"
rg -q 'archived snapshot' "${ARCHIVE_DIR}/docs/index.html"
rg -q 'View the latest documentation' "${ARCHIVE_DIR}/docs/index.html"
rg -q 'ainfra/v0.test/' "${ARCHIVE_DIR}/index.html"
rg -q 'href=/ainfra/v0.test/' "${ARCHIVE_DIR}/index.html"

test ! -e "${ROOT_DIR}/package.json"
test ! -e "${ROOT_DIR}/package-lock.json"
test -f "${ROOT_DIR}/themes/hextra/theme.toml"

echo "Documentation build and smoke checks passed."
