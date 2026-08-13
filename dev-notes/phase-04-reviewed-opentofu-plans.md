# Phase 4: Reviewed OpenTofu plans

Status: in progress.

Phase 4 begins from verified release `v1.0.0-alpha.3` and composes the Phase 3
immutable template boundary into native OpenTofu planning and apply. It does
not yet standardize provider output, generate inventory, run Ansible, or
implement destruction and recovery; those remain in Phases 5 and 6.

## Completion outcome

A user can create an apply plan from a verified locked template and declared
native inputs, review a sanitized structural summary, and apply only that exact
saved plan by ID. Every plan binding and saved-plan byte is reverified before
execution, and durable evidence distinguishes started, succeeded, failed,
cancelled, and ambiguous outcomes.

## Requirement ownership

| Requirement group | Phase 4 disposition |
|---|---|
| `AINFRA-LOCK-002`–`003` | Enforce template digest at plan time and consume only plan-bound materialization during apply. |
| `AINFRA-PLAN-001`–`004` | Implement collision-resistant IDs, sensitive saved-plan storage, complete summaries, and strict apply/destroy intent separation. |
| `AINFRA-APPLY-001`–`005` | Require exact reviewed IDs, reverify bindings and bytes, apply saved plans only, record lifecycle evidence, and route ambiguous interruption to inspection. |
| OpenTofu subprocess controls | Use the controlled executable/argument/environment boundary and declared backend/variable-file ordering. |
| Output, inventory, Ansible | Defer to Phase 5. |
| Destroy and recovery hardening | Defer to Phase 6. |

## Delivery sequence

1. Freeze plan record, run evidence, and machine-result contracts.
2. Add the controlled OpenTofu adapter and executable fingerprinting.
3. Create private run workspaces from verified locked templates.
4. Hash deployment metadata, native inputs, template identity, and tools.
5. Run `tofu init` with declared backend configuration in source order.
6. Create saved apply plans with declared variable files in source order.
7. Produce sanitized structural summaries and immutable plan records.
8. Require an exact plan ID and reverify every binding immediately before
   `tofu apply <saved-plan>`.
9. Record durable started and terminal evidence, including ambiguous
   interruption handling.
10. Add compiled black-box, cancellation, corruption, race, redaction, and
    conformance coverage before the Phase 4 release gate.

## First implementation slice

The first dependency-ordered slice freezes the plan/run schemas and implements
the shell-free OpenTofu adapter with controlled fake-executable tests. It must
not create provider infrastructure until the saved-plan binding and evidence
contracts are in place.

Tracking WorkItem:
`BACK-20260813_1736-SmartArch-implement-phase-four-reviewed-opentofu-plans`.

## Implementation record

### 2026-08-13 — Plan contracts and controlled OpenTofu adapter

- Added closed v1 schemas and conforming examples for immutable reviewed-plan
  bindings and lifecycle run metadata.
- Bound plan intent, deployment and template digests, ordered native inputs,
  OpenTofu version/executable digest, saved-plan bytes, and summary path.
- Added the `internal/tofu` adapter as a shell-free argument-array boundary for
  `tofu init` and saved `tofu plan`, preserving declared backend and variable
  file order and making apply versus destroy intent explicit.
- Reused executable fingerprint, contained working-directory, explicit
  environment, cancellation, and sanitized child IO policy from `internal/exec`.
- Added controlled-runner tests for exact argument boundaries, ordering,
  destroy separation, traversal refusal, and non-zero child outcomes.

The next slice creates private run workspaces from verified locked templates
and computes the complete deployment, input, template, and tool binding set
before invoking this adapter.

### 2026-08-13 — Verified run workspaces and pre-engine bindings

- Added exclusive, owner-only run allocation under a trusted runs root with
  cleanup of failed partial preparation.
- Reverified the digest-addressed cache entry, copied only validated template
  content into the private run workspace, and reverified the resulting tree.
- Bound deployment manifest bytes, ordered OpenTofu backend and variable input
  bytes, locked template source/resolution/digest, and the fingerprinted
  OpenTofu executable before any engine invocation.
- Refused unsafe run IDs, existing run reuse, poisoned cache content, escaping
  or non-regular native inputs, and changed executable bytes.
- Added tests for private permissions, complete deterministic binding order,
  workspace equality, poisoned-cache refusal, exclusive IDs, and partial-run
  cleanup.

The next slice composes preparation with `tofu init` and saved apply planning,
then atomically publishes the immutable reviewed-plan record and structural
summary without yet enabling apply.

### 2026-08-13 — Saved apply-plan creation

- Snapshotted declared backend and variable files into the private run before
  engine execution so OpenTofu cannot consume mutable deployment inputs after
  their digests are bound.
- Corrected adapter path validation to evaluate arguments from the actual
  engine working directory while still proving containment by the run root.
- Composed verified preparation, ordered `tofu init`, saved apply planning, and
  `tofu show -json` through the controlled adapter.
- Reduced raw plan JSON in memory to action counts only; values and other
  potentially secret fields are neither represented nor persisted.
- Published private `plan.json`, `plan-record.json`, and `run.json` only after
  successful initialization, planning, summary decoding, and saved-plan
  hashing. The plan record binds every pre-engine input and the final plan
  bytes.
- Added end-to-end application coverage for argument paths and ordering,
  plan-byte binding, private record permissions, and secret non-persistence.

The next slice exposes `ainfra plan` through trusted configuration and stable
text/JSON results, with generated collision-resistant run IDs and complete
failure cleanup. Apply remains unavailable until exact reviewed-plan loading
and immediate binding reverification are implemented.
