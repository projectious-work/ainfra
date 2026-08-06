#!/usr/bin/env bash
set -euo pipefail

project_root=$(git -C "$(dirname "$0")/../.." rev-parse --show-toplevel)
remote_url=$(git -C "${project_root}" remote get-url origin)
source_branch=$(git -C "${project_root}" branch --show-current)
source_commit=$(git -C "${project_root}" rev-parse --short HEAD)
source_tree=$(git -C "${project_root}" status --porcelain)

if [[ -n "${source_tree}" ]]; then
  echo "error: documentation deployment requires a clean worktree" >&2
  exit 1
fi

"${project_root}/docs/scripts/build.sh"

deploy_dir=$(mktemp -d)
trap 'rm -rf "${deploy_dir}"' EXIT

cp -R "${project_root}/docs/public/." "${deploy_dir}/"
touch "${deploy_dir}/.nojekyll"

git -C "${deploy_dir}" init -q
git -C "${deploy_dir}" switch -q --orphan gh-pages
git -C "${deploy_dir}" add -A

git_user=$(git -C "${project_root}" config user.name || true)
git_email=$(git -C "${project_root}" config user.email || true)
git_user=${git_user:-ainfra-release-bot}
git_email=${git_email:-release@ainfra.local}

git -C "${deploy_dir}" \
  -c "user.name=${git_user}" \
  -c "user.email=${git_email}" \
  commit -q -m \
  "docs: deploy from ${source_branch}@${source_commit} ($(date -u +%Y-%m-%dT%H:%M:%SZ))"

git -C "${deploy_dir}" push --force "${remote_url}" gh-pages:gh-pages
echo "deployed ${source_branch}@${source_commit} to gh-pages"
