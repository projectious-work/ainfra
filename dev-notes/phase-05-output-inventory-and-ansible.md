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

## Requirement and disposition matrix

| Normative requirement | Disposition | Phase 5 implementation and evidence |
|---|---|---|
| `AINFRA-CONTRACT-022`–`027` | implement | Preserve declared Ansible variable-file order and opaque bytes; consume only reviewed snapshots without merging, renaming, defaulting, generating, or interpreting native values. |
| `AINFRA-CONTRACT-030`–`038` | implement | Require a non-sensitive declared OpenTofu output, strictly validate the closed v1 shape, reject unknown versions, fields, connections, names, counts, reserved facts, and secret-shaped material, then generate only the documented deterministic inventory mapping. |
| `AINFRA-INV-001`–`004` | implement | Pure sorted inventory conversion, atomic private output/inventory publication, structural and secret-pattern refusal, and typed infrastructure-only applicability behavior. |
| `AINFRA-TPL-002`, `005`, `009`–`010` | supporting | Native variable transport and minimized output remain template contracts; configure/check consume them without reinterpretation and require exact-host zero-change Runner evidence after convergence. |
| `AINFRA-CLI-001`–`005`, `007`–`013` | supporting | Output, inventory, configure, and deploy use established target resolution, static help, text/JSON selection, stream separation, cancellation, confirmation, invocation, and exit contracts. Direct child-tool documentation is retained with the user workflow. |
| `AINFRA-OUTPUT-001`–`008`, `010`–`015` | implement | Shared typed results, one closed JSON envelope, stable text, centralized redaction, schema validation, engine attribution, and explicit `applicable` or `not-applicable` results. |
| `AINFRA-SEC-010`–`015`, `020`–`027`, `040`–`042` | implement | Shell-free argv, fingerprinted tools, closed environments, contained directories, sensitive evidence, structural/secret scanning, forced host-key verification, bound populated `known_hosts`, no key generation, and no implicit agent forwarding. |
| `AINFRA-SEC-030`–`032` | supporting | Output is obtained through OpenTofu's output protocol without reading state; all native inputs remain opaque reviewed-plan bindings. |
| `AINFRA-TEST-001`–`003`, `010`–`018` | implement for Phase 5 behavior | Unit, schema, CLI-contract, compiled black-box, malformed/corrupt/symlink, cancellation, concurrency, race, coverage, vulnerability, security, and container gates cover the owned parsers and lifecycle. |
| `AINFRA-DOC-010`–`013`; applicable `AINFRA-DOC-001`–`005` | implement | The phase record and user documentation describe commands, applicability, native boundaries, security controls, evidence, convergence, and release state without claiming unpublished support. |
| Destruction, consolidated logs, recovery hardening, and live provider certification | deferred | These remain explicitly owned by Phases 6, 9, and later template-certification work and are not used to claim Phase 5 completion. |

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

### 2026-08-14 — Phase 5 release-gate hardening

- Forced Ansible's host-key policy through a closed environment and proved
  inherited disabling overrides cannot cross the child-process boundary.
- Added negative coverage for missing and empty `known_hosts`, while proving
  local inventory neither requires nor receives SSH configuration.
- Added cancellation propagation, corrupt and symlinked Runner artifact
  refusal, duplicate terminal-summary refusal, and bounded root-scoped event
  traversal.
- Ran competing configuration starts under the deployment operation lock and
  proved exactly one can enter the mutating stage.
- Passed the full normal, race, coverage, static-analysis, schema, container,
  vulnerability, and security validation gates.
- Passed the live and shipped-context structural release audit with zero
  errors. Its sole warning is the existing absence of processkit skills under
  `src/context/skills/`, unrelated to the ainfra Phase 5 deliverable.

Phase 5 implementation is release-gate ready. The roadmap remains
`in_progress` until a Phase 5 release is packaged, host-verified, signed,
published, and independently verified; only then may it transition to
`shipped` and Phase 6 begin.

### 2026-08-14 — Independent requirement sweep and gap closure

The release sweep checked implementation, tests, schemas, examples, user
documentation, and retained evidence against every row above. It found one
blocking `AINFRA-INV-004` lifecycle-reporting gap: for `inventory: none`, the
composed deploy path omitted its skipped post-apply stages instead of recording
them as not applicable. Deploy results now carry ordered typed stage reports;
infrastructure-only runs record output, inventory, configure, and
configure-check as `not-applicable`, while applicable runs record the actual
stage outcomes. Schema, unit, and compiled CLI coverage protect the result.

After that correction, the sweep found no unexplained skip, specification
inconsistency, overclaim, or unowned Phase 5 implementation gap. Release
packaging may begin, but shipment still requires the owner-executed host gate,
keyless signing, publication, and independent artifact verification.
