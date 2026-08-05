# 11. Acceptance, migration, and open decisions

## Product acceptance

ainfra v1 is ready only when all mandatory requirements are implemented and
these journeys pass from a clean supported environment:

1. initialize and validate a deployment;
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
14. build Linux/macOS binaries and validate the optional Dockerfile.

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
- unsafe Dockerfile changes.

## Template acceptance

A certified template requires:

- schema/layout conformance;
- OpenTofu fmt/init/validate;
- Ansible syntax and lint checks;
- pinned dependencies;
- complete native variable documentation;
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

## Delivery milestones

1. **Contracts:** schemas, examples, parsing, validation, lockfile, diagnostics.
2. **OpenTofu lifecycle:** init, validate, plan, binding, apply, output,
   destroy, run evidence.
3. **Ansible lifecycle:** inventory, Runner integration, SSH trust, check-mode
   verification, deploy composition.
4. **Sources and hardening:** Git subdirectories, cache, redaction, security
   fixtures, recovery/status.
5. **Template and documentation:** reference template, authoring/AI guide,
   clean-room conformance.
6. **Release candidate:** all gates, Linux/macOS packaging, optional Dockerfile,
   migration guide, disposable acceptance.

Each milestone ends in a usable vertical slice and MUST not reintroduce a
meta-language.

## Open decisions

These details remain intentionally open for implementation design review:

- exact Go CLI parsing library versus standard-library flag composition;
- exact run-ID format;
- canonical JSON/YAML libraries and schema validator;
- whether OCI template sources enter v1.x after Git sources stabilize;
- minimum supported OpenTofu and Ansible Runner versions at implementation
  start;
- release signing/attestation backend;
- precise multi-environment support beyond the preferred one-state-boundary
  deployment.

None of these may weaken the native-variable, process-boundary, source-locking,
or reviewed-plan decisions.
