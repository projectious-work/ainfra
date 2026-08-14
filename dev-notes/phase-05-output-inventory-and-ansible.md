# Phase 5: Output, inventory, and Ansible

Status: in progress.

Phase 5 begins from verified release `v1.0.0-alpha.4` and extends the reviewed
OpenTofu apply boundary into configuration management. It standardizes
non-secret provider output, derives deterministic inventory, runs Ansible with
native variable files, and proves convergence without weakening the immutable
plan and evidence guarantees delivered by Phase 4.

## Completion outcome

A user can apply an exact reviewed infrastructure plan, validate its declared
non-secret output, generate stable inventory from that output, and run the
declared Ansible configuration through controlled subprocess boundaries. A
subsequent run demonstrates zero infrastructure and configuration drift while
durable evidence identifies every executed stage.

## Delivery sequence

1. Freeze the standardized output and inventory contracts.
2. Validate output shape, types, sensitivity, and required fields.
3. Generate deterministic inventory without shell evaluation.
4. Define native Ansible inventory, variable-file, and playbook ordering.
5. Add a controlled Ansible executable boundary and tool fingerprinting.
6. Bind configuration inputs and generated inventory to run evidence.
7. Report stable text and JSON results across infrastructure and configuration
   stages.
8. Prove zero-change convergence with compiled black-box and negative tests.
9. Complete race, cancellation, redaction, and release-gate validation.

Destruction, teardown verification, and broader interruption recovery remain
owned by Phase 6.

Tracking WorkItem:
`BACK-20260814_1013-SteadyDell-phase-five-output-inventory-ansible`.

## Implementation progress

### 2026-08-14 — Standardized handoff and configuration lifecycle

- Extended reviewed-plan bindings to cover every declared Ansible variable
  file and the deployment's `known_hosts` input, so configuration cannot
  consume bytes outside the reviewed run.
- Added bounded OpenTofu output collection that retains only the declared,
  non-sensitive output value and strictly validates the closed v1 schema.
- Added pure deterministic inventory conversion with sorted groups, hosts,
  keys, and stable YAML serialization. Unknown fields, unsupported connection
  shapes, invalid names, and secret-shaped material fail closed.
- Added atomic private `output.json` and `inventory.yaml` publication with
  SHA-256 artifact results and typed `not-applicable` results for
  infrastructure-only templates.
- Added a shell-free Ansible Runner adapter, bounded version negotiation,
  contained project/inventory/input paths, forced SSH host-key checking, and a
  populated bound `known_hosts` requirement for SSH inventories.
- Added normal configuration and independent check-mode verification. Runner
  `playbook_on_stats` evidence must cover the exact expected host set with zero
  changes, failures, and unreachable hosts.
- Added `output`, `inventory`, `configure`, and composed `deploy` command
  dispatch with stable text/JSON results, durable stage events, replay refusal
  for mutating configuration, and retained Runner artifact references.
- Added unit, CLI-contract, schema, and compiled black-box coverage for the
  reviewed apply-to-output-to-inventory-to-Ansible path and the composed
  deploy path.

The remaining Phase 5 release-gate work is adversarial cancellation,
concurrency, corrupted Runner-evidence, SSH-policy, redaction, and race
coverage plus final user documentation and the complete validation sweep.
