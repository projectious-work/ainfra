#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "${script_dir}/.." && pwd)"
# shellcheck source=scripts/release-lib.sh
source "${script_dir}/release-lib.sh"

release_evidence_dir() {
  local version="$1" commit="$2"
  printf '%s/dist/release-evidence/v%s/%s\n' \
    "${root}" "${version}" "${commit}"
}

release_tool_version() {
  if command -v "$1" >/dev/null 2>&1; then
    "$@" 2>&1 | head -1
  else
    printf 'missing\n'
  fi
}

release_state_report() {
  local version="$1" commit="$2" evidence_dir="$3"
  local cargo_version branch
  cargo_version="$(sed -nE \
    '0,/^version = "([^"]+)"$/s//\1/p' "${root}/Cargo.toml")"
  branch="$(git -C "${root}" branch --show-current)"
  {
    printf '# Release state: v%s\n\n' "${version}"
    printf -- '- Candidate commit: `%s`\n' "${commit}"
    printf -- '- Candidate branch: `%s`\n' "${branch}"
    printf -- '- Cargo package version: `%s`\n' "${cargo_version}"
    printf -- '- Rust toolchain: `%s`\n' \
      "$(tr '\n' ' ' < "${root}/rust-toolchain.toml")"
    printf -- '- rustc: `%s`\n' "$(release_tool_version rustc -Vv)"
    printf -- '- cargo: `%s`\n' "$(release_tool_version cargo -V)"
    printf -- '- OpenTofu: `%s`\n' "$(release_tool_version tofu version)"
    printf -- '- Ansible: `%s`\n' \
      "$(release_tool_version ansible-playbook --version)"
    printf '\n## Bound inputs\n\n'
    printf '```text\n'
    (
      cd "${root}"
      printf '%s  Cargo.lock\n' "$(release_sha256 Cargo.lock)"
      find schemas templates -type f -print \
        | LC_ALL=C sort \
        | while IFS= read -r input; do
            printf '%s  %s\n' "$(release_sha256 "${input}")" "${input}"
          done
    )
    printf '```\n'
  } > "${evidence_dir}/RELEASE-STATE.md"
  [[ "${cargo_version}" == "${version}" ]] \
    || release_die "Cargo.toml must already contain version ${version}"
}

release_doctor_report() {
  local version="$1" commit="$2" evidence_dir="$3"
  local ainfra_bin="${AINFRA_BIN:-${root}/target/debug/ainfra}"
  local project doctor_output doctor_status
  if [[ ! -x "${ainfra_bin}" ]]; then
    (cd "${root}" && cargo build --locked)
  fi
  project="$(mktemp -d)"
  trap 'rm -rf "${project}"' RETURN
  (
    cd "${project}"
    "${ainfra_bin}" init --name release-doctor --format json >/dev/null
  )
  set +e
  doctor_output="$(
    cd "${project}"
    "${ainfra_bin}" doctor --environment development --format json 2>&1
  )"
  doctor_status=$?
  set -e
  {
    printf '# Release doctors: v%s\n\n' "${version}"
    printf -- '- Candidate commit: `%s`\n' "${commit}"
    printf -- '- Exit status: `%s`\n\n' "${doctor_status}"
    printf '```json\n%s\n```\n' "${doctor_output}"
  } > "${evidence_dir}/RELEASE-DOCTORS.md"
  rm -rf "${project}"
  trap - RETURN
  [[ "${doctor_status}" -eq 0 ]] \
    || release_die "release doctor failed; see ${evidence_dir}"
}

release_docs_report() {
  local version="$1" commit="$2" evidence_dir="$3"
  local notes="${root}/release-notes/v${version}.md"
  local docs_output docs_status docs_destination
  [[ -f "${notes}" ]] \
    || release_die "missing release notes: release-notes/v${version}.md"
  for heading in Added Changed Fixed 'Known limitations' Install Rollback; do
    grep -Fqx "## ${heading}" "${notes}" \
      || release_die "release notes are missing section: ${heading}"
  done
  grep -Fq '## Rollback' "${root}/README.md" \
    || release_die "README.md must document release rollback"
  docs_destination="$(mktemp -d)"
  set +e
  docs_output="$(
    "${root}/docs/scripts/build-docs.sh" \
      --destination "${docs_destination}" 2>&1
  )"
  docs_status=$?
  set -e
  rm -rf "${docs_destination}"
  {
    printf '# Release documentation gate: v%s\n\n' "${version}"
    printf -- '- Candidate commit: `%s`\n' "${commit}"
    printf -- '- Release notes: `release-notes/v%s.md`\n' "${version}"
    printf -- '- Production build exit status: `%s`\n\n' "${docs_status}"
    printf '```text\n%s\n```\n' "${docs_output}"
  } > "${evidence_dir}/RELEASE-DOCS.md"
  [[ "${docs_status}" -eq 0 ]] \
    || release_die "documentation build failed; see ${evidence_dir}"
}

release_phase0() {
  local version="$1" commit evidence_dir
  release_validate_version "${version}"
  commit="$(git -C "${root}" rev-parse HEAD)"
  evidence_dir="$(release_evidence_dir "${version}" "${commit}")"
  mkdir -p "${evidence_dir}"
  release_state_report "${version}" "${commit}" "${evidence_dir}"
  release_doctor_report "${version}" "${commit}" "${evidence_dir}"
  release_docs_report "${version}" "${commit}" "${evidence_dir}"
  (
    cd "${evidence_dir}"
    release_sha256 RELEASE-STATE.md
    release_sha256 RELEASE-DOCTORS.md
    release_sha256 RELEASE-DOCS.md
  ) > "${evidence_dir}/PHASE0.sha256"
  printf 'phase-0 evidence: %s\n' "${evidence_dir}"
}

command="${1:-help}"
shift || true
case "${command}" in
  phase0)
    [[ "$#" -eq 1 ]] || release_die "phase0 requires VERSION"
    release_phase0 "$1"
    ;;
  *)
    printf 'Usage: scripts/release-evidence.sh phase0 VERSION\n' >&2
    exit 2
    ;;
esac
