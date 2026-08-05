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

```text
cmd/
└── ainfra/
    └── main.go
internal/
├── app/
├── command/
├── project/
├── template/
├── source/
├── lock/
├── tofu/
├── inventory/
├── ansible/
├── run/
├── security/
├── exec/
├── diagnostic/
└── output/
schemas/
testdata/
```

Production code MUST remain under `internal/` until an external Go API has a
real consumer and separate stability commitment.

## Package responsibilities

### `cmd/ainfra`

Composition root only: parse process-global startup facts, construct concrete
adapters, call `command.Run`, and map the result to an exit code. It MUST NOT
contain lifecycle or validation logic.

### `command`

Defines CLI syntax, flags, help, command-specific input conversion, and result
rendering selection. It calls `app` use cases and contains no engine calls.

### `app`

Coordinates product use cases such as doctor, plan, deploy, and destroy. It
owns sequencing and transaction boundaries but delegates parsing, persistence,
security policy, and engines to focused packages. Files SHOULD be organized by
use case rather than one large lifecycle file.

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
deterministic report ordering used by doctor. Individual packages
provide checks for the boundaries they own; `diagnostic` does not duplicate
their parsing or engine logic. Checks receive explicit capabilities and cannot
obtain network or process access implicitly. The registry accepts only
ainfra-owned contract checks; it is not a provider or template-specific plugin
system.

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

Owns Ansible Runner invocation, runner input-directory construction, event
interpretation, expected-host verification, and check-mode outcomes. It does
not provision infrastructure or acquire templates.

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

### `diagnostic`

Defines stable error/finding codes, severity, component, explanation, and next
action. Concrete packages create typed diagnostics; this package contains no
business policy.

### `output`

Versioned human and JSON result envelopes. Rendering MUST be separate from use
case execution so JSON mode never inherits incidental terminal prose.

## Dependency rules

- `command` may depend on `app`, `diagnostic`, and `output`.
- `app` may depend on domain packages and consumer-defined interfaces.
- domain packages may depend on `diagnostic` and narrowly on `security`.
- engine/source adapters may depend on `exec`.
- `exec`, `diagnostic`, and pure inventory conversion MUST NOT depend on `app`
  or `command`.
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
