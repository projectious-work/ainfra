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
