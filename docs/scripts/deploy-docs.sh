#!/usr/bin/env bash
set -euo pipefail

# Build locally and publish the generated site from the gh-pages branch.
# The repository intentionally does not use GitHub Actions.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PAGES_BRANCH="${PAGES_BRANCH:-gh-pages}"
BUILD_DIR="${ROOT_DIR}/public"
DOCS_BASE_URL="${DOCS_BASE_URL:-https://projectious-work.github.io/ainfra/}"
DOCS_LATEST_URL="https://projectious-work.github.io/ainfra/"
DOCS_VERSION="${DOCS_VERSION:-main}"
DOCS_DEPLOY_DRY_RUN="${DOCS_DEPLOY_DRY_RUN:-false}"

if [[ ! "${DOCS_VERSION}" =~ ^[A-Za-z0-9._-]+$ ]]; then
  echo "DOCS_VERSION must contain only letters, numbers, dots, underscores, or hyphens." >&2
  exit 1
fi

if [[ "${DOCS_DEPLOY_DRY_RUN}" != "true" &&
  "${DOCS_DEPLOY_DRY_RUN}" != "false" ]]; then
  echo "DOCS_DEPLOY_DRY_RUN must be true or false." >&2
  exit 1
fi

if [[ "${DOCS_VERSION}" != "main" &&
  "${DOCS_BASE_URL}" == "${DOCS_LATEST_URL}" ]]; then
  DOCS_BASE_URL="${DOCS_BASE_URL}${DOCS_VERSION}/"
fi

BUILD_ARGS=()
if [[ "${DOCS_VERSION}" != "main" ]]; then
  BUILD_DIR="${BUILD_DIR}/${DOCS_VERSION}"
fi

DOCS_BASE_URL="${DOCS_BASE_URL}" DOCS_VERSION="${DOCS_VERSION}" \
  "${ROOT_DIR}/scripts/build-docs.sh" "${BUILD_ARGS[@]}" \
  --destination "${BUILD_DIR}"

WORKTREE_DIR="$(mktemp -d)"
cleanup() {
  git -C "${ROOT_DIR}" worktree remove --force "${WORKTREE_DIR}" \
    >/dev/null 2>&1 || true
  rmdir "${WORKTREE_DIR}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

if git show-ref --verify --quiet "refs/heads/${PAGES_BRANCH}"; then
  git -C "${ROOT_DIR}" worktree add "${WORKTREE_DIR}" "${PAGES_BRANCH}"
elif git ls-remote --exit-code origin "refs/heads/${PAGES_BRANCH}" \
  >/dev/null 2>&1; then
  git -C "${ROOT_DIR}" fetch origin \
    "${PAGES_BRANCH}:${PAGES_BRANCH}"
  git -C "${ROOT_DIR}" worktree add "${WORKTREE_DIR}" "${PAGES_BRANCH}"
else
  git -C "${ROOT_DIR}" worktree add --detach "${WORKTREE_DIR}"
  git -C "${WORKTREE_DIR}" checkout --orphan "${PAGES_BRANCH}"
fi

if [[ "${DOCS_VERSION}" == "main" ]]; then
  V01_TREE_BEFORE=""
  if git -C "${WORKTREE_DIR}" cat-file -e HEAD:v0.1 2>/dev/null; then
    V01_TREE_BEFORE="$(git -C "${WORKTREE_DIR}" rev-parse HEAD:v0.1)"
  fi

  for ENTRY in \
    "${WORKTREE_DIR}"/* \
    "${WORKTREE_DIR}"/.[!.]* \
    "${WORKTREE_DIR}"/..?*; do
    [[ -e "${ENTRY}" ]] || continue
    ENTRY_NAME="${ENTRY##*/}"
    if [[ "${ENTRY_NAME}" == ".git" ||
      "${ENTRY_NAME}" =~ ^v[0-9]+([.][0-9]+)*$ ]]; then
      continue
    fi
    rm -rf -- "${ENTRY}"
  done
  cp -R "${BUILD_DIR}/." "${WORKTREE_DIR}/"
else
  VERSION_DIR="${WORKTREE_DIR}/${DOCS_VERSION}"
  mkdir -p "${VERSION_DIR}"
  find "${VERSION_DIR}" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
  cp -R "${BUILD_DIR}/." "${VERSION_DIR}/"
fi
: > "${WORKTREE_DIR}/.nojekyll"

git -C "${WORKTREE_DIR}" add -A
if [[ "${DOCS_VERSION}" == "main" && -n "${V01_TREE_BEFORE:-}" ]]; then
  INDEX_TREE="$(git -C "${WORKTREE_DIR}" write-tree)"
  V01_TREE_AFTER="$(
    git -C "${WORKTREE_DIR}" rev-parse "${INDEX_TREE}:v0.1"
  )"
  if [[ "${V01_TREE_AFTER}" != "${V01_TREE_BEFORE}" ]]; then
    echo "Refusing deployment: the v0.1 archive changed." >&2
    exit 1
  fi
fi

if [[ "${DOCS_DEPLOY_DRY_RUN}" == "true" ]]; then
  git -C "${WORKTREE_DIR}" diff --cached --stat
  echo "Documentation deployment dry run passed; nothing was pushed."
  exit 0
fi

if git -C "${WORKTREE_DIR}" diff --cached --quiet; then
  echo "No documentation changes to deploy."
else
  git -C "${WORKTREE_DIR}" commit \
    -m "docs: deploy Hugo site $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  git -C "${WORKTREE_DIR}" push origin "${PAGES_BRANCH}"
  echo "Documentation deployed to ${PAGES_BRANCH}."
fi
