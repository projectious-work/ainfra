# 12. Implementation roadmap

## How to read this roadmap

ainfra is built in small, documented phases. Each phase is a usable vertical
slice with one clear outcome, not a promise that every listed detail ships at
once. Phase status is one of:

- **In progress:** active specification or implementation work;
- **Planned:** accepted direction, ordered provisionally; or
- **Possible:** a speculative future option that requires a separate decision.

The grouping and concise phase summaries are inspired by the
[Tau roadmap](https://twotimespi.dev/roadmap/) and its public YAML-backed
presentation. This document remains Markdown because it is a versioned product
planning artifact, not input to the ainfra CLI.

## The first usable release

### Phase 0 — Product contract

**In progress.** Finalize the v1 product boundary, native-file contracts,
security model, Go package architecture, schemas, examples, and acceptance
journeys.

### Phase 1 — Go project foundation

**Planned.** Establish the Go module, command shell, typed results, process
runner, filesystem abstractions, test fixtures, developer toolchain, and
Linux/macOS build path.

### Phase 2 — Projects, templates, and doctor

**Planned.** Discover deployments, parse manifests and native files, validate
paths and layouts, and deliver the single doctor surface with stable text and
JSON findings.

### Phase 3 — Immutable template sources

**Planned.** Resolve local and Git-subdirectory sources, create deterministic
digests and lock files, materialize contained workspaces, and detect source or
cache drift.

### Phase 4 — Reviewed OpenTofu lifecycle

**Planned.** Run init and native-file validation, create bound saved plans,
apply only the reviewed plan, collect standardized output, and record durable
run evidence.

### Phase 5 — Inventory and Ansible lifecycle

**Planned.** Convert standardized non-secret output into deterministic
inventory, invoke Ansible Runner with native variables, interpret structured
events, and prove zero-change convergence.

### Phase 6 — Destruction, recovery, and hardening

**Planned.** Add reviewed destroy plans, empty-state verification, interruption
recovery, redaction, executable and plan binding, cache defenses, and the
negative security fixture suite.

### Phase 7 — Reference template and authoring experience

**Planned.** Certify the Hetzner Kubernetes, Cloudflare tunnel, and temporary
bastion reference template; complete human and AI authoring guidance; and add
contract-version diagnosis and safe template migration assistance.

### Phase 8 — v1 release readiness

**Planned.** Complete disposable end-to-end acceptance, dependency and supply-
chain checks, Linux/macOS packaging, changelog and migration documentation, and
validation of the optional development Dockerfile.

## After the v1 foundation

### Phase 9 — Template ecosystem polish

**Possible.** Improve template test harnesses, compatibility fixtures,
deprecation reporting, migration previews, and certified-template evidence
without adding a provider-specific plugin layer to the CLI.

### Phase 10 — Catalog and trust metadata

**Possible.** Define interoperable catalog metadata for discovering certified
Git-hosted templates, including publisher identity, compatibility, signatures,
attestations, and revocation information while keeping source resolution
explicit and locked.

### Phase 11 — Operational evidence export

**Possible.** Export sanitized, signed deployment evidence for audits and
support cases, with clear provenance and retention controls and without
exporting plans, state, credentials, or raw sensitive output.

### Phase 12 — Authoring and editor integration

**Possible.** Publish schema bundles, completions, explainable diagnostics, and
editor integrations that consume doctor JSON rather than introducing another
configuration language.

## Speculative horizon

### Phase 13 — Additional immutable source transports

**Possible.** Evaluate OCI artifacts or content-addressed archives after Git
source locking is proven, using the same resolved-source and digest contracts.

### Phase 14 — External policy evaluation

**Possible.** Allow organizations to evaluate plans and manifests with an
external policy engine at explicit lifecycle gates. ainfra would transport
sanitized inputs and enforce the result, not own a policy language.

### Phase 15 — Deployment-set coordination

**Possible.** Explore coordinating several independent deployments while
preserving separate state, plans, locks, failure boundaries, and approvals.
This must not turn `ainfra.yaml` into an infrastructure meta-language.

### Phase 16 — Remote execution protocol

**Possible.** Define a process-isolated execution protocol for controlled build
or operations environments. Any design must preserve local CLI semantics,
explicit credentials, plan review, and auditable child-tool boundaries.

## Roadmap governance

- Phase numbers express implementation order, not release numbers or dates.
- A phase may be split when that creates a safer usable slice.
- Possible phases are not commitments and MUST NOT shape v1 abstractions
  prematurely.
- Moving a possible phase into the planned group requires a recorded product
  decision, threat-model review, and updated acceptance criteria.
- Completed phases SHOULD link their release, changelog entry, and detailed
  implementation notes.
