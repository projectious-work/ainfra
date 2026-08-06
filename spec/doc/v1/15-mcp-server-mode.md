## Status and objective

MCP server mode is planned for v1 after the core infrastructure lifecycle and
hardening phases. It makes ainfra directly usable by MCP-capable agents and
editors without wrapping shell commands or parsing terminal prose. It remains
an adapter over stable typed application use cases and is not a second
execution path.

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

## Initial capability boundary

The initial server exposes only read-only ainfra-owned capabilities:

- doctor without `--reconcile`;
- deployment and template contract inspection;
- status and sanitized run summaries;
- sanitized standardized output and generated inventory; and
- published schemas, supported versions, and documentation references.

Tool handlers call the same typed application use cases as the CLI. They MUST
NOT execute a shell, reconstruct CLI argument strings, parse console output, or
implement a second lifecycle. MCP input and output schemas are derived from the
same typed contracts and versioned machine results used by normal commands.

The initial mode does not expose reconciliation, template writes, lock/update,
plan, apply, configure, deploy, or destroy. MCP annotations describing a tool as
read-only or destructive are useful client metadata but MUST NOT be treated as
authorization controls.

## Future mutation support

Mutating tools MAY be considered only after read-only server operation is
stable. They require a separate accepted security design covering:

- explicit server-start capability allowlisting;
- per-request caller authorization independent of descriptive tool metadata;
- non-interactive confirmation or externally verifiable approval artifacts;
- exact saved-plan and operation binding for apply or destroy;
- deployment operation locking, cancellation, and ambiguous-state recovery;
- complete sanitized audit evidence; and
- proof that an MCP client cannot broaden project roots, environment access,
  credentials, executable selection, or log destinations.

Starting a server with access to credentials MUST NOT by itself authorize a
mutation. No future MCP tool may create an implicit plan or bypass the normal
reviewed-plan protocol.

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
- **AINFRA-MCP-002:** the initial tool registry MUST be read-only and fixed by
  the shipped version, not extended by templates.
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

[mcp-spec]: https://modelcontextprotocol.io/specification/latest
