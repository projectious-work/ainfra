# Phase 4: Reviewed OpenTofu plans

Status: shipped in `v1.0.0-alpha.4`.

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

### 2026-08-13 — Reviewed plan CLI

- Exposed `ainfra plan [DEPLOYMENT] [--destroy]` with trusted project and
  configuration resolution and stable text and v1 JSON results.
- Kept OpenTofu discovery lazy, fingerprinted the selected executable, captured
  its machine-reported version, and launched it with a closed environment.
- Generated timestamped 128-bit random run IDs and emitted the saved-plan,
  locked-template, and aggregate deployment-input digests needed for review.
- Removed incomplete run directories after every post-preparation failure and
  preserved successful private plan evidence for later exact-ID application.
- Added command dispatch coverage for target and destroy intent while retaining
  the existing adapter, lifecycle, security, and schema conformance suites.

The next slice loads an exact reviewed plan ID, reverifies all immutable
bindings immediately before execution, and implements saved-plan-only apply
with durable started and terminal evidence. Destroy execution remains deferred
to Phase 6.

### 2026-08-14 — Exact reviewed-plan apply

- Exposed `ainfra apply [DEPLOYMENT] --plan RUN_ID` while continuing to refuse
  destroy-plan execution through the apply command.
- Acquired the deployment operation lock, reloaded the deployment, and
  immediately reverified the manifest, ordered native inputs, template lock,
  cache and workspace template bytes, OpenTofu executable and version, summary,
  and saved-plan bytes.
- Executed only `tofu apply` with the reviewed `plan.tfplan`; no implicit plan
  generation or argument-string shell boundary was introduced.
- Added exclusive durable `started` evidence plus `succeeded`, `failed`, or
  `cancelled` terminal evidence. Ambiguous interruption additionally records
  `inspection-required` and disables automatic retry.
- Added stale-binding, exact-argument, lifecycle-event, command-dispatch, and
  workspace-runtime-exclusion coverage.

The next slice hardens cancellation and evidence-failure paths with compiled
black-box fixtures, adds concurrent-operation coverage, and completes Phase 4
release-gate conformance before publication.

### 2026-08-14 — Phase 4 release-gate hardening

- Added a compiled CLI fixture that locks a local template, creates a reviewed
  plan, and applies it through a controlled fake OpenTofu executable.
- Proved the child receives only the exact saved-plan apply arguments and that
  started and succeeded evidence survives the process boundary.
- Ran competing apply processes against one reviewed plan: the deployment lock
  and exclusive start record permit exactly one execution and refuse the other.
- Added replay and tampered-plan refusal coverage, cancellation-to-inspection
  assertions, post-execution evidence-failure handling, and race execution.
- Corrected workspace reverification to exclude nested `.terraform` runtime
  directories and generated lock files only when those lock files were absent
  from the reviewed template; injected configuration remains digest-bound.
- Passed the context and shipped-context release audit with zero errors. Its
  sole warning is the existing absence of processkit skills under
  `src/context/skills/`, unrelated to the ainfra Phase 4 deliverable.

Phase 4 implementation is release-gate ready. The roadmap remains
`in_progress` until a Phase 4 release is packaged, host-verified, signed,
published, and independently verified; only then may it transition to
`shipped` and Phase 5 begin.

### 2026-08-14 — Reviewed-plan user documentation

- Replaced the Phase 3-only usage and quick-start guidance with the complete
  plan-review-apply workflow available on the development branch.
- Added configuration, plan intent, stable JSON result, immutable binding,
  exact saved-plan invocation, evidence, replay, concurrency, and recovery
  guidance.
- Kept the root README release claims unchanged because the latest published
  release remains `v1.0.0-alpha.3` without Phase 4 functionality.

This closes the Phase 4 implementation documentation gap. The remaining work
is the continuous Phase 4 release flow and post-publication roadmap transition,
which require an explicit release request.

### 2026-08-14 — Release publication and shipment

Release `v1.0.0-alpha.4` was built from clean release commit `fd9f315`, passed
the native macOS and hardened container gate, and published with four platform
archives, four SPDX JSON SBOMs, a SHA-256 manifest, and its keyless Sigstore
bundle. Publication independently downloaded and verified every checksum and
the signature.

The externally verifiable release satisfies the final shipment gate. Phase 4
is shipped, and Phase 5 now owns standardized non-secret output,
deterministic inventory generation, Ansible execution with native variables,
and zero-change convergence verification.
