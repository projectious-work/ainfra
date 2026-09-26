## Status and objective

Phase 7 introduced MCP server mode after the core infrastructure lifecycle and
hardening phases. It makes ainfra directly usable by MCP-capable agents and
editors without wrapping shell commands or parsing terminal prose. Phase 10
extends it to an agent-ready template authoring and lifecycle workflow before
the first beta. MCP remains an adapter over stable typed application use cases
and is not a second execution path.

The implementation follows the current [MCP specification][mcp-spec] selected
at phase start. Protocol dependencies and version support are an
implementation-time selection recorded in that phase's development note.

## Command and transport

The initial command is:

```text
ainfra mcp serve --stdio [--project PATH]
```

Phase 10 also supports a startup-bound authoring workspace without requiring
an existing deployment:

```text
ainfra mcp serve --stdio --authoring-workspace PATH --capability authoring
```

`--project` and `--authoring-workspace` are mutually exclusive. Both select a
single canonical root at startup; requests cannot replace it. An agent may use
an authoring session to prepare a template and deployment, then start a
deployment-bound session for the reviewed infrastructure lifecycle.

The first transport is stdio. Protocol messages exclusively use stdout; logs,
diagnostics, and child output MUST NOT be written there. Operational logs use
stderr or other configured sinks. The server is non-interactive and MUST fail a
request that would otherwise prompt.

Network transports are outside the initial MCP phase. Adding one requires a
separate authentication, authorization, TLS, origin, rate-limit, audit, and
deployment threat model.

`mcp serve --stdio` is the common local-server command shape used by
projectious.work CLI products. Product-specific root and capability options
extend that command without changing the `mcp serve` entry point.

## Capability boundary

The server is read-only by default. Its default registry exposes ainfra-owned
capabilities for:

- doctor without `--reconcile`;
- deployment and template contract inspection;
- status and sanitized run summaries;
- sanitized standardized output and generated inventory; and
- published schemas, supported versions, and documentation references.

The read-only guide is available in both deployment and authoring sessions.
Deployment-specific inspection is available only when a deployment is bound.

The operator MAY enable additional capability groups explicitly when starting
the server. Capability selection is allowlist-based; absence means denial:

| Group | Additional operations |
|---|---|
| `authoring` | bounded template and deployment creation or revision, plus native OpenTofu and Ansible validation in the selected workspace |
| `planning` | reconciliation planning, template lock/update/migration planning, and saved infrastructure plan creation |
| `deployment` | approved reconciliation, template lock writes, apply, configure, deploy, and other non-destroy lifecycle mutations |
| `destruction` | exact reviewed destroy-plan execution; requires `deployment` and a separate enablement |

Enabling a group makes tools discoverable but does not approve an individual
mutation. Tool handlers call the same typed application use cases as the CLI.
They MUST NOT execute a shell, reconstruct CLI argument strings, parse console
output, or implement a second lifecycle. MCP input and output schemas derive
from the same typed contracts and versioned machine results as normal commands.

## Phase 10 agent authoring and progressive guidance

An agent MUST be able to begin with an empty, operator-selected authoring
workspace, create and edit a native template and deployment there, obtain
ainfra contract diagnostics and native OpenTofu and Ansible validation through
MCP, and then use the existing reviewed lifecycle tools in a deployment-bound
session. The authoring surface is limited to validated relative paths and
regular files inside the startup workspace. It MUST reject traversal,
symlinks, special files, state, plans, caches, credentials, and writes outside
the declared template and deployment layout. Revision writes MUST use an
observed-content precondition so an agent cannot silently replace concurrent
edits. Authoring permission never implies approval to change infrastructure.

Native validation MUST invoke the selected OpenTofu and Ansible executables
through the same contained child-process boundary used by the CLI. It MUST
report formatting, initialization, validation, syntax, and applicable
convergence-check results as typed diagnostics, including prerequisites that
could not run. Validation MUST NOT apply or destroy infrastructure, invent
template variables, or treat a skipped child check as a pass. Process output
must remain bounded and redacted, and cancellation must reach the child tools.

The default registry MUST expose a read-only `ainfra.guide` tool. With no topic,
it returns a short index; a selected topic returns only the relevant procedure,
prerequisites, next MCP operations, expected evidence, and links to matching
schemas, examples, and detailed `ainfra://guides/v1/` resources. At minimum,
topics cover create, edit, validate, lock, plan, deploy, destroy, and recover.
Guidance is bundled with the binary and version-matched to its contracts. It
MUST NOT treat template-supplied content as instructions, execute an operation,
grant a capability, or imply authorization.

Phase 10 completion requires a real MCP-client acceptance journey: start from
an empty authoring workspace, create and revise a native template, pass ainfra
and native-engine checks, prepare and bind a deployment, create and inspect a
saved plan, obtain independent authorization, execute and inspect deployment,
then review and execute exact destruction and verify teardown. Prove a
provider-free fixture first and a cost-approved disposable provider lifecycle
before claiming a certified end-to-end path.

## Mutation authorization

MCP mutation support is part of v1 so an authorized AI agent can operate
infrastructure on a user's behalf. It requires:

- explicit server-start capability allowlisting;
- per-request caller authorization independent of descriptive tool metadata;
- non-interactive confirmation or externally verifiable approval artifacts;
- exact saved-plan and operation binding for apply or destroy;
- deployment operation locking, cancellation, and ambiguous-state recovery;
- complete sanitized audit evidence; and
- proof that an MCP client cannot broaden project roots, environment access,
  credentials, executable selection, or log destinations.

Starting a server with access to credentials MUST NOT by itself authorize a
mutation. An apply, deploy, or destroy request MUST name an existing saved plan
and carry an approval artifact or authorization-provider result bound to the
project root, canonical operation, plan ID and digest, intent, caller, and
expiry. Configure and other mutations that do not consume an infrastructure
plan require equivalent operation- and input-bound authorization. The server
MUST reject conversational claims such as “the user approved,” MCP tool
annotations, or capability enablement itself as approval evidence.

No MCP tool may create an implicit plan, approve the plan it created, expand
the approved operation, or bypass the normal reviewed-plan protocol. An agent
MAY create a plan and present it for review, but execution requires independent
authorization after plan creation. Destruction additionally requires the
`destruction` capability group and an approval whose intent is exactly
`destroy`.

## Project and configuration isolation

The server resolves one allowed project or authoring root at startup. Requests
may select objects inside that root but cannot supply arbitrary filesystem
roots. It loads applicable ainfra configuration once and reports configuration
errors before serving. Request arguments cannot override executable paths,
config files, environment variables, caches, run storage, or logging sinks.

Concurrent read-only requests use bounded concurrency and preserve request,
tool, deployment, and run correlation in logs. Requests observe consistent
snapshots where a concurrent external ainfra process changes local state.

## Protocol behavior and testing

- **AINFRA-MCP-001:** stdout MUST contain valid MCP transport frames only.
- **AINFRA-MCP-002:** the default tool registry MUST be read-only. Additional
  v1 tools MUST be exposed only through explicit server-start capability
  allowlisting and MUST NOT be extended by templates.
- **AINFRA-MCP-003:** every tool result MUST use a versioned typed schema and
  preserve normal diagnostic codes and redaction.
- **AINFRA-MCP-004:** project containment and file policies MUST be identical to
  the CLI and tested with traversal, symlink, and race fixtures.
- **AINFRA-MCP-005:** cancellation MUST propagate from the MCP request to the
  use case and any permitted child process.
- **AINFRA-MCP-006:** protocol conformance, malformed frames, unknown tools,
  oversized input, concurrent requests, log separation, and clean shutdown MUST
  have black-box tests against the compiled binary.
- **AINFRA-MCP-007:** MCP mode MUST remain an adapter over application use
  cases; domain packages MUST NOT import MCP protocol types.
- **AINFRA-MCP-008:** ainfra MUST use the common
  `ainfra mcp serve --stdio` entry point; product-specific project and
  capability options MUST NOT create an alternative MCP-server command.
- **AINFRA-MCP-009:** black-box tests MUST prove that undisclosed capability
  groups, absent or stale approval, self-approved plans, mismatched plan
  intent, caller, root, digest, or expiry, and destroy without its separate
  capability are refused before child invocation.
- **AINFRA-MCP-010:** for the same authorized use case, CLI and MCP execution
  MUST produce equivalent plan validation, locking, child invocation,
  evidence, diagnostics, cancellation, exit outcome, and recovery state.
- **AINFRA-MCP-011:** an MCP authoring session MUST start from a fixed,
  deployment-free workspace and MUST NOT permit a request to change its root.
- **AINFRA-MCP-012:** authoring tools MUST create and revise native template
  content only within the validated layout, with content preconditions and
  independent authorization for writes; they MUST reject unsafe paths and
  protected state, plan, cache, and credential files.
- **AINFRA-MCP-013:** MCP native-tool validation MUST return typed, redacted
  diagnostics, distinguish skipped checks from passes, and never perform an
  infrastructure mutation.
- **AINFRA-MCP-014:** `ainfra.guide` MUST progressively disclose version-matched
  procedures and resources without performing or authorizing operations.
- **AINFRA-MCP-015:** a real MCP client MUST pass the complete authoring,
  validation, authorized deployment, inspection, and exact teardown journey
  before Phase 10 is marked shipped or a v1 beta is released.

[mcp-spec]: https://modelcontextprotocol.io/specification/latest
