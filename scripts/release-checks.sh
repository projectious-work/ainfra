#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "${script_dir}/.." && pwd)"
# shellcheck source=scripts/release-lib.sh
source "${script_dir}/release-lib.sh"

gate_command() {
  case "$1" in
    rust-quality)
      cargo fmt --check
      cargo clippy --all-targets --all-features -- -D warnings
      ;;
    rust-tests)
      cargo test --all-targets --all-features
      cargo audit
      ;;
    repository)
      uv run --frozen ruff format --check src tests scripts
      uv run --frozen ruff check src tests scripts
      uv run --frozen pytest
      ;;
    infrastructure)
      scripts/test-ansible
      scripts/security-all
      docs/scripts/test-docs.sh
      ;;
    *) release_die "unknown release gate: $1" ;;
  esac
}

write_binding() {
  local version="$1" commit="$2" destination="$3"
  {
    printf 'version=%s\n' "${version}"
    printf 'candidate_commit=%s\n' "${commit}"
    printf 'cargo_lock_sha256=%s\n' \
      "$(release_sha256 "${root}/Cargo.lock")"
    printf 'rustc_fingerprint=%s\n' \
      "$(rustc -Vv | release_sha256 /dev/stdin)"
    printf 'environment_scope=local-release\n'
  } > "${destination}"
}

evidence_reusable() {
  local evidence_dir="$1" expected_binding="$2"
  local digest file
  [[ -f "${evidence_dir}/CHECKS.binding" ]] || return 1
  [[ -f "${evidence_dir}/CHECKS.sha256" ]] || return 1
  cmp -s "${expected_binding}" "${evidence_dir}/CHECKS.binding" || return 1
  while read -r digest file; do
    file="${file#\*}"
    file="${file# }"
    [[ -f "${evidence_dir}/${file}" ]] || return 1
    [[ "$(release_sha256 "${evidence_dir}/${file}")" == "${digest}" ]] \
      || return 1
  done < "${evidence_dir}/CHECKS.sha256"
}

run_gate() {
  local gate="$1" evidence_dir="$2"
  (
    cd "${root}"
    set +e
    gate_command "${gate}" > "${evidence_dir}/${gate}.log" 2>&1
    printf '%s\n' "$?" > "${evidence_dir}/${gate}.status"
  )
}

release_checks() {
  local version="$1" commit evidence_dir binding jobs gate count
  local gates=(rust-quality rust-tests repository infrastructure)
  local pids=()
  release_validate_version "${version}"
  commit="$(git -C "${root}" rev-parse HEAD)"
  evidence_dir="${root}/dist/release-evidence/v${version}/${commit}"
  mkdir -p "${evidence_dir}"
  binding="$(mktemp)"
  trap 'rm -f "${binding}"' RETURN
  write_binding "${version}" "${commit}" "${binding}"
  if evidence_reusable "${evidence_dir}" "${binding}"; then
    printf 'reusing verified release checks: %s\n' "${evidence_dir}"
    return
  fi

  jobs="${AINFRA_RELEASE_JOBS:-2}"
  [[ "${jobs}" =~ ^[1-4]$ ]] \
    || release_die "AINFRA_RELEASE_JOBS must be between 1 and 4"
  count=0
  for gate in "${gates[@]}"; do
    run_gate "${gate}" "${evidence_dir}" &
    pids+=("$!")
    count=$((count + 1))
    if [[ "${count}" -eq "${jobs}" ]]; then
      for pid in "${pids[@]}"; do
        wait "${pid}"
      done
      pids=()
      count=0
    fi
  done
  for pid in "${pids[@]}"; do
    wait "${pid}"
  done

  local failed=0 status
  for gate in "${gates[@]}"; do
    status="$(sed -n '1p' "${evidence_dir}/${gate}.status")"
    [[ "${status}" == 0 ]] || failed=1
  done
  [[ "${failed}" -eq 0 ]] \
    || release_die "release checks failed; see ${evidence_dir}"
  cp "${binding}" "${evidence_dir}/CHECKS.binding"
  (
    cd "${evidence_dir}"
    for gate in "${gates[@]}"; do
      printf '%s  %s\n' \
        "$(release_sha256 "${gate}.log")" "${gate}.log"
      printf '%s  %s\n' \
        "$(release_sha256 "${gate}.status")" "${gate}.status"
    done
  ) > "${evidence_dir}/CHECKS.sha256"
  rm -f "${binding}"
  trap - RETURN
  printf 'release checks evidence: %s\n' "${evidence_dir}"
}

command="${1:-help}"
shift || true
case "${command}" in
  run)
    [[ "$#" -eq 1 ]] || release_die "run requires VERSION"
    release_checks "$1"
    ;;
  *)
    printf 'Usage: scripts/release-checks.sh run VERSION\n' >&2
    exit 2
    ;;
esac
