#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCS_BASE_URL="${DOCS_BASE_URL:-https://projectious-work.github.io/ainfra/}"
DOCS_LATEST_URL="https://projectious-work.github.io/ainfra/"
DOCS_VERSION="${DOCS_VERSION:-main}"
BUILD_DIR="${ROOT_DIR}/public"
HUGO_CACHEDIR="${HUGO_CACHEDIR:-${ROOT_DIR}/.cache/hugo}"
export HUGO_CACHEDIR

BUILD_ARGS=("$@")
VERSION_CONFIG=""
VERSION_CONFIG_DIR=""
cleanup() {
  if [[ -n "${VERSION_CONFIG}" ]]; then
    rm -f "${VERSION_CONFIG}"
  fi
  if [[ -n "${VERSION_CONFIG_DIR}" ]]; then
    rmdir "${VERSION_CONFIG_DIR}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

if [[ ! "${DOCS_VERSION}" =~ ^[A-Za-z0-9._-]+$ ]]; then
  echo "DOCS_VERSION contains unsupported characters." >&2
  exit 1
fi

if [[ "${DOCS_VERSION}" != "main" ]]; then
  VERSION_CONFIG_DIR="$(
    mktemp -d "${TMPDIR:-/tmp}/ainfra-docs-version.XXXXXX"
  )"
  VERSION_CONFIG="${VERSION_CONFIG_DIR}/version.yaml"
  {
    printf 'params:\n'
    printf '  version: "%s"\n' "${DOCS_VERSION}"
    printf '  archived_version: true\n'
    printf '  url_latest_version: "%s"\n' "${DOCS_LATEST_URL}"
    printf '  banner:\n'
    printf '    key: "archived-%s"\n' "${DOCS_VERSION}"
    printf '    message: |\n'
    printf '      **Version %s** is an archived snapshot. ' \
      "${DOCS_VERSION}"
    printf '[View the latest documentation](%s).\n' "${DOCS_LATEST_URL}"
  } >"${VERSION_CONFIG}"
  BUILD_ARGS+=(
    --config
    "${ROOT_DIR}/hugo.yaml,${VERSION_CONFIG}"
  )
fi

for ((i = 0; i < ${#BUILD_ARGS[@]}; i++)); do
  case "${BUILD_ARGS[i]}" in
    --destination)
      if ((i + 1 >= ${#BUILD_ARGS[@]})); then
        echo "--destination requires a value." >&2
        exit 1
      fi
      BUILD_DIR="${BUILD_ARGS[i + 1]}"
      ;;
    --destination=*)
      BUILD_DIR="${BUILD_ARGS[i]#--destination=}"
      ;;
  esac
done

if [[ "${BUILD_DIR}" != /* ]]; then
  BUILD_DIR="${ROOT_DIR}/${BUILD_DIR}"
fi

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
hugo --gc --minify --cleanDestinationDir --baseURL "${DOCS_BASE_URL}" \
  "${BUILD_ARGS[@]}"
: > "${BUILD_DIR}/.nojekyll"
