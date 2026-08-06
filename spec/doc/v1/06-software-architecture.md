## Principles

- small, cohesive, self-contained packages;
- dependency direction from application policy toward replaceable adapters;
- explicit data flow and immutable value types at boundaries;
- no template/provider semantics in the CLI core;
- no global mutable state;
- no shell execution;
- interfaces defined by consumers and kept narrow;
- standard library before dependencies when the implementation remains clear;
- package tests alongside code and black-box behavior tests at boundaries;
- no generic `util`, `common`, `manager`, or `helpers` dumping grounds.

See [package-dependencies.svg](package-dependencies.svg).

## Proposed source layout

| Module or path | Purpose |
|---|---|
| `cmd/ainfra` | Compose the executable and map process startup and exit behavior. |
| `internal/command` | Define commands, flags, help, input conversion, and renderer selection. |
| `internal/app` | Coordinate use cases and transaction boundaries without implementing adapters. |
| `internal/config` | Load, validate, merge, and explain typed CLI configuration. |
| `internal/project` | Discover and validate deployment roots, manifests, and native input paths. |
| `internal/template` | Validate template manifests, layouts, compatibility, and content digests. |
| `internal/source` | Acquire contained local or Git template sources through immutable descriptors. |
| `internal/lock` | Parse, compare, canonicalize, and atomically write template lock files. |
| `internal/tofu` | Construct documented OpenTofu invocations and return typed outcomes. |
| `internal/inventory` | Convert validated standardized output into deterministic Ansible inventory. |
| `internal/ansible` | Construct Ansible Runner invocations and return its attributed engine outcomes. |
| `internal/evidence` | Normalize, classify, redact, retain, and render child-engine evidence without interpreting deployment state. |
| `internal/evidence/opentofu` | Decode and version-check OpenTofu machine-readable UI records. |
| `internal/evidence/ansible` | Decode and version-check Ansible Runner job events and reported statistics. |
| `internal/run` | Store run metadata, append events, and enforce recovery state transitions. |
| `internal/diagnostic` | Register doctor checks and define stable typed findings and reports. |
| `internal/reconcile` | Plan and apply permitted local fixes, then verify their outcomes. |
| `internal/migration` | Plan and apply explicit template-contract version transformations. |
| `internal/security` | Enforce containment, file, environment, redaction, and output policies. |
| `internal/exec` | Execute child processes without a shell under explicit IO and environment policy. |
| `internal/logging` | Create redacted structured events and deliver them to configured sinks. |
| `internal/output` | Render typed results as rich text, plain text, or versioned JSON. |
| `schemas` | Embed published schemas required for offline contract validation. |
| `testdata` | Hold non-secret fixtures and fake child executables shared by tests. |
| `test/blackbox` | Exercise the compiled CLI and its observable process contract. |
| `test/integration` | Exercise adapters against real local child tools. |
| `test/e2e` | Exercise opt-in disposable certified-template lifecycles. |

Production code MUST remain under `internal/` until an external Go API has a
real consumer and separate stability commitment.

The `schemas` package MUST configure the selected Go validator to assert every
standard `format` used by published contracts, including `date-time` and `uri`.
Format support is a tested runtime capability, not an annotation-only parser
option. End users receive this validation through normal loading and doctor;
they do not install the repository's development validator.

## Package responsibilities

### `cmd/ainfra`

Composition root only: parse process-global startup facts, construct concrete
adapters, call `command.Run`, and map the result to an exit code. It MUST NOT
contain lifecycle or validation logic.

### `command`

Defines CLI syntax, flags, help, command-specific input conversion, and result
rendering selection. It calls `app` use cases and contains no engine calls.

### `app`

Coordinates product use cases such as doctor, template migration, plan, deploy,
and destroy. It owns sequencing and transaction boundaries but delegates
parsing, persistence, security policy, and engines to focused packages. Files
SHOULD be organized by use case rather than one large lifecycle file.

### `config`

Loads every configuration layer, validates the published schema, applies
deterministic precedence, records value provenance, and returns one immutable
effective configuration. It does not parse native deployment variables or
child-tool credentials.

### `project`

Discovers deployment roots, parses and validates `ainfra.yaml`, resolves native
input paths, and exposes an immutable deployment model. It does not acquire
templates or execute tools.

### `template`

Parses the template manifest, validates layout and compatibility, computes the
canonical template digest, and exposes an immutable template model. It knows no
provider-specific variables.

### `diagnostic`

Defines the check registry, scopes, applicability rules, finding model, and
deterministic report ordering used by doctor, including stable codes, severity,
explanation, and next action. Individual packages provide checks for the
boundaries they own; `diagnostic` does not duplicate their parsing or engine
logic. Checks receive explicit capabilities and cannot obtain network or
process access implicitly. The registry accepts only ainfra-owned contract
checks; it is not a provider or template-specific plugin system.

### `reconcile`

Selects registered reconcilers for doctor findings, builds the confirmed change
plan, rechecks preconditions, applies permitted local fixes, records partial
failure, and requests verification. It cannot access provider, backend, state,
credential, or remote-host operations.

### `migration`

Plans explicit contract-version transitions and applies deterministic local
source transformations through small version-to-version migrators. It produces
a change set before writing, does not operate on provider state, and delegates
post-migration conformance to `template` and `diagnostic`.

### `source`

Implements local and Git source acquisition behind a small `Resolver`
interface. It owns cache containment and returns an immutable resolved source
descriptor. Git execution is delegated to `exec`.

### `lock`

Parses, canonicalizes, compares, and atomically writes `ainfra.lock`. It has no
network access.

### `tofu`

Owns construction and interpretation of documented OpenTofu CLI calls. It
accepts native paths and typed operation options, delegates processes to
`exec`, and returns typed results. It must not parse or rewrite tfvars.

### `inventory`

Pure transformation from validated standardized output to deterministic
Ansible inventory. It must have no filesystem or process dependency in its
core conversion function.

### `ansible`

Owns Ansible Runner invocation, runner input-directory construction,
expected-host verification, and check-mode outcomes. It does
not provision infrastructure, acquire templates, parse fact caches, or model
desired or current host state. Check-mode outcome evaluation consumes
normalized Runner evidence and cannot make a stronger convergence claim than
Ansible reports.

### `evidence`

Owns the common normalized event model, error-only selection, redaction,
retention references, and human rendering for child-engine evidence. Framing,
OpenTofu protocol-version negotiation, Ansible Runner compatibility checks,
structural validation, resource limits, path safety, and sensitive-field
handling remain typed Go adapter code. Built-in,
schema-validated profiles MAY declare stable field pointers and classification
mappings for each supported protocol. Profiles are embedded ainfra assets,
not template or deployment inputs, and cannot execute expressions or override
security policy. Unknown event types remain attributed, unclassified evidence.

The profile contract is
[`../../schemas/v1/engine-evidence-profile.schema.json`](../../schemas/v1/engine-evidence-profile.schema.json),
with maintained OpenTofu and Ansible Runner examples under
[`../../examples/v1/evidence-profiles/`](../../examples/v1/evidence-profiles/).
Adding a future engine requires an explicit adapter and reviewed profile; a
YAML file alone cannot make an arbitrary executable trusted or supported.

### `run`

Allocates run IDs/directories, persists immutable metadata, appends events,
loads status, and enforces recovery state transitions. It does not decide
business sequencing.

### `security`

Contains cross-boundary policies with focused components: path containment,
file-type policy, sensitive-value registry/redaction, output scanning, and
environment allowlisting. Security checks SHOULD also be invoked at the
affected package boundary; this package is not a late monolithic audit pass.

### `exec`

The only generic child-process implementation. It takes executable, argument
array, cwd, allowed environment, IO policy, cancellation context, and sensitive
values. It forwards signals, streams output, records sanitized evidence, and
returns typed exit information. It never invokes a shell.


### `logging`

Defines structured operational events, verbosity filtering, correlation, and
stderr, rotating-file, and local-syslog sinks. It receives only values already
processed by the security redaction boundary and does not persist run evidence.

### `output`

Owns semantic command results and rich, plain, and versioned JSON renderers.
Rendering MUST be separate from use-case execution and operational logging so
JSON mode never inherits terminal prose or log events.

## Dependency rules

- `command` may depend on `app`, `config`, `diagnostic`, `logging`, and
  `output`.
- `app` may depend on domain packages and consumer-defined interfaces.
- domain packages may depend on `diagnostic` and narrowly on `security`.
- engine/source adapters may depend on `exec`; engine adapters may emit native
  records to `evidence` but MUST NOT depend on a renderer.
- `exec`, `config`, `diagnostic`, `evidence`, `logging`, `output`, and pure
  inventory conversion MUST NOT depend on `app` or `command`.
- adapters MUST NOT call each other; `app` owns sequencing.
- cycles are forbidden and checked in the build gate.

## Testability boundaries

Filesystem, clock, ID generation, environment, and subprocess execution MUST
be replaceable at use-case boundaries. Interfaces SHOULD model a needed
operation—not mirror a concrete type. Most unit tests use in-memory values,
temporary directories, and fake engine executables. Real OpenTofu/Ansible
integration tests are separate and explicitly marked.

## Concurrency

The first implementation SHOULD remain sequential for mutating operations.
Concurrency MAY be used for independent read-only checks or downloads only
when cancellation, ordering, log correlation, and the race detector are
covered. Simplicity wins over throughput.

## Size and review signals

A package whose purpose cannot be explained in one sentence, a file that mixes
multiple use cases, or an interface with unrelated methods triggers design
review. These are review signals, not arbitrary line-count failures.
