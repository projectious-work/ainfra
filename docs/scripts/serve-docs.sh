#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCS_BASE_URL="${DOCS_BASE_URL:-http://localhost:1313/ainfra/}"
HUGO_CACHEDIR="${HUGO_CACHEDIR:-${ROOT_DIR}/.cache/hugo}"
export HUGO_CACHEDIR

command -v hugo >/dev/null 2>&1 || {
  echo "Hugo extended is required: https://gohugo.io/installation/" >&2
  exit 1
}

if [[ ! -f "${ROOT_DIR}/themes/hextra/theme.toml" ]]; then
  git -C "${ROOT_DIR}/.." submodule update --init --recursive \
    docs/themes/hextra
fi

cd "${ROOT_DIR}"
mkdir -p "${HUGO_CACHEDIR}"
hugo server --buildDrafts --disableFastRender --baseURL "${DOCS_BASE_URL}" "$@"
