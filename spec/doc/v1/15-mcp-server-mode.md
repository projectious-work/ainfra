## Status and objective

The initial guarded MCP stdio server shipped in roadmap Phase 7. Roadmap Phase
10 makes MCP the primary agent experience and hardens its capability,
authorization, resource, durable-operation, and adversarial-security contracts.
It remains an adapter over stable typed application use cases and is not a
second execution path. The CLI remains complete and behaviorally equivalent
for direct human, CI, recovery, and break-glass operation.

The implementation follows the current [MCP specification][mcp-spec] selected
at phase start. Protocol dependencies and version support are an
implementation-time selection recorded in that phase's development note.

## Command and transport

The initial command is:

```text
ainfra mcp serve --stdio [--project PATH]
```

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

The operator MAY enable additional capability groups explicitly when starting
the server. Capability selection is allowlist-based; absence means denial:

| Group | Additional operations |
|---|---|
| `planning` | reconciliation planning, template lock/update/migration planning, and saved infrastructure plan creation |
| `deployment` | approved reconciliation, template writes, apply, configure, deploy, and other non-destroy lifecycle mutations |
| `destruction` | exact reviewed destroy-plan execution; requires `deployment` and a separate enablement |

Enabling a group makes tools discoverable but does not approve an individual
mutation. Tool handlers call the same typed application use cases as the CLI.
They MUST NOT execute a shell, reconstruct CLI argument strings, parse console
output, or implement a second lifecycle. MCP input and output schemas derive
from the same typed contracts and versioned machine results as normal commands.

## Agent experience

The server presents bounded infrastructure use cases rather than CLI syntax.
An agent can discover applicable templates and requirements, inspect and
diagnose a deployment, create and retrieve a reviewed plan, request an
independently authorized execution, observe durable operation status, and
retrieve sanitized results, evidence, and recovery advice. It never receives
an arbitrary shell or natural-language deployment tool.

Every result identifies whether the operation applies, stable findings and
outcome codes, blocking and advisory conditions, approval requirements,
durable operation identity, permitted next actions, and recovery state where
relevant. Human prose MAY accompany the typed result but cannot be its only
meaning.

The server exposes version-matched resources for published schemas, supported
contract versions, template-authoring guidance, safe template metadata,
credential-provider capabilities, sanitized results, reviewed plan summaries,
and recovery guidance. Untrusted template prose is labeled as template content
and cannot alter tool or server policy.

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

The server resolves one allowed project root at startup. Requests may select
objects inside that root but cannot supply arbitrary filesystem roots. It loads
normal ainfra configuration once and reports configuration errors before
serving. Request arguments cannot override executable paths, config files,
environment variables, caches, run storage, or logging sinks.

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
- **AINFRA-MCP-011:** MCP MUST be the leading documented agent interaction
  surface without becoming the domain or application package boundary.
- **AINFRA-MCP-012:** mutating tools MUST return or accept a durable operation
  identity that remains queryable after client retry, cancellation, or session
  loss and prevents ambiguous duplicate execution.
- **AINFRA-MCP-013:** tool schemas MUST classify side effects and identify
  required capabilities and independent authorization without treating
  descriptive annotations as enforcement.
- **AINFRA-MCP-014:** resources and tool definitions MUST be embedded,
  versioned, stable for a released binary, and immune to template-controlled
  mutation, name shadowing, and instruction injection.
- **AINFRA-MCP-015:** structured results MUST expose applicable next actions
  from a closed set and MUST NOT return arbitrary executable remediation from
  untrusted content.
- **AINFRA-MCP-016:** a plan explanation MUST identify its source plan and
  sanitization limits and MUST NOT replace the saved plan or approval review.
- **AINFRA-MCP-017:** the initial stdio server MUST NOT depend on model vendor,
  client-specific conversation state, or private chain-of-thought.
- **AINFRA-MCP-018:** a future network transport MUST have a separate remote
  threat model and MUST NOT expose local stdio assumptions directly.

[mcp-spec]: https://modelcontextprotocol.io/specification/latest
