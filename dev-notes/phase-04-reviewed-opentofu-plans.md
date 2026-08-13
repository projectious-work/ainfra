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
