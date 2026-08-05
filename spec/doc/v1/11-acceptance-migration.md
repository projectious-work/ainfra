## Product acceptance

ainfra v1 is ready only when all mandatory requirements are implemented and
these journeys pass from a clean supported environment:

1. initialize and doctor a deployment;
2. lock and use a local template;
3. lock a Git tag/commit plus subdirectory and detect mutable-source drift;
4. plan using the unmodified native `terraform.tfvars`;
5. reject apply after any reviewed binding changes;
6. apply the exact saved plan;
7. validate standard non-secret output and deterministically generate
   inventory;
8. run Ansible Runner with the unmodified `ansible-vars.yaml`;
9. prove a zero-change check-mode verification;
10. report meaningful status after success, failure, cancellation, and an
    ambiguous interrupted event;
11. reject an apply plan for destroy and vice versa;
12. apply an exact destroy plan and verify empty state;
13. complete machine JSON and exit-code compatibility fixtures;
14. diagnose environment, deployment, template/source, and run fixtures with
    stable findings and correct pass/skip/warning/fail states;
15. reconcile a safe local finding idempotently, reject an unsafe finding, and
    preserve evidence for a partially failed reconciliation;
16. produce a dry-run template migration plan, apply a safe local migration,
    validate it, and reject an ambiguous or stateful migration; and
17. build Linux/macOS binaries and validate the optional Dockerfile.

## Security acceptance

Negative fixtures MUST prove rejection or redaction for:

- traversal and symlink escape;
- special files and unsafe permissions;
- mutable or digest-mismatched template content;
- command/argument injection strings;
- secret-shaped template input/output/inventory;
- known secret values split across output chunks;
- disabled SSH host-key checking;
- backend and plan binding changes;
- missing or changed child executable;
- malformed Ansible events and partial host recap;
- corrupt or reordered run events;
- poisoned cache entry;
- unsafe Dockerfile changes;
- reconciliation path escape, precondition race, partial failure, and attempts
  to modify native inputs, plans, state, credentials, or remote resources.

## Template acceptance

A conforming or certified template requires:

- successful validation of every ainfra-owned document against its declared
  schema, with no unknown fields, plus layout conformance;
- OpenTofu fmt/init/validate;
- Ansible syntax and lint checks;
- pinned dependencies;
- complete native variable documentation;
- standard variable-reference structure with documented required/default,
  type, constraint, sensitivity, and cross-variable information;
- standard output fixtures;
- security policy tests;
- cost-approved disposable plan/apply/configure/check/reapply/destroy;
- zero-change verification after convergence;
- independent provider and state confirmation after teardown;
- redacted evidence with date, versions, operator, limitations, and result.

## Rust-era migration

The rewrite is intentionally not a port.

- Rust-era `TemplateInput`, compiled tfvars, embedded-template discovery,
  capability allowlist, and run protocols are not v1 contracts.
- v1 MAY provide a read-only diagnostic that identifies old files and links to
  migration documentation; it MUST NOT silently translate or execute them.
- users create a new deployment with native tfvars and Ansible variables,
  select and lock a conforming template, and make a new reviewed plan.
- existing infrastructure requires an explicit state adoption/import guide;
  migration MUST never assume ownership from filenames alone.
- the old implementation remains reachable through version control and
  releases for forensic recovery.

## Delivery roadmap

The phased delivery sequence and potential future directions are maintained in
the [implementation roadmap data](roadmap.yaml). Each phase ends in a usable
vertical slice and MUST not reintroduce a meta-language.

## Implementation-time selections

The product specification does not preselect dependencies or external versions
whose suitability depends on the implementation date. Each selection is made
at the start of the phase that first needs it, after requirements are concrete:

| Selection | Decision point | Required criteria |
|---|---|---|
| CLI parser or standard `flag` composition | Go foundation | Prefer the standard library; add a library only when the specified hierarchy, help, completion, or error behavior would otherwise require substantial custom framework code. |
| JSON, YAML, and JSON Schema implementations | Contracts and doctor | Standards compliance, strict unknown-field behavior, maintained security posture, deterministic output, low dependency weight, and required Go-version support. |
| Minimum OpenTofu and Ansible Runner versions | First real-tool integration | Required CLI/event features, upstream support and security status, availability in the developer environment, and an affordable compatibility matrix. |
| Release signing or attestation backend | Release packaging implementation | Public verification, identity and key-rotation model, automation without exported long-lived secrets, platform availability, and recovery documentation. |

The phase note records candidates considered, selected versions, rationale,
licenses, security review, and validation evidence before the dependency or
backend becomes production-critical. A separate ADR is added when the choice
creates a durable cross-package or release-infrastructure constraint.

- **AINFRA-IMPL-001:** implementation-time selection MUST preserve every
  behavioral and security requirement in this specification.
- **AINFRA-IMPL-002:** standard-library functionality is preferred when it
  remains clear and maintainable; avoiding all dependencies is not itself a
  reason to build a private framework.
- **AINFRA-IMPL-003:** selected child-tool minimums MUST be documented, tested,
  and reported by `ainfra doctor environment` before mutating commands ship.

Multi-environment coordination beyond one deployment and one state boundary is
not part of v1. Introducing it requires a future product decision and MUST NOT
shape the initial manifest or execution architecture prematurely.
