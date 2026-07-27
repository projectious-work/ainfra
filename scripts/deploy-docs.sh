#!/usr/bin/env bash
set -euo pipefail

# Build locally and publish the generated site from the gh-pages branch.
# The repository intentionally does not use GitHub Actions.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PAGES_BRANCH="${PAGES_BRANCH:-gh-pages}"
BUILD_DIR="${ROOT_DIR}/public"
DOCS_BASE_URL="${DOCS_BASE_URL:-https://projectious-work.github.io/ainfra-templates/}"

DOCS_BASE_URL="${DOCS_BASE_URL}" "${ROOT_DIR}/scripts/build-docs.sh"

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

find "${WORKTREE_DIR}" -mindepth 1 -maxdepth 1 ! -name .git \
  -exec rm -rf {} +
cp -R "${BUILD_DIR}/." "${WORKTREE_DIR}/"
: > "${WORKTREE_DIR}/.nojekyll"

git -C "${WORKTREE_DIR}" add -A
if git -C "${WORKTREE_DIR}" diff --cached --quiet; then
  echo "No documentation changes to deploy."
else
  git -C "${WORKTREE_DIR}" commit \
    -m "docs: deploy Hugo site $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  git -C "${WORKTREE_DIR}" push origin "${PAGES_BRANCH}"
  echo "Documentation deployed to ${PAGES_BRANCH}."
fi
