# Phase 7: Guarded MCP server mode

Status: in progress

Phase 7 exposes ainfra's typed application use cases through the common
`ainfra mcp serve --stdio` entry point. The default registry is read-only;
planning, deployment, and destruction remain absent unless their capability
groups and independent authorization contracts are implemented and enabled.

## Protocol selection

Phase 7 targets MCP `2026-07-28`, the current generally available
specification when implementation began on 2026-08-15. The adapter uses the
official `github.com/modelcontextprotocol/go-sdk` module at `v1.7.0-pre.3`,
the current SDK release line implementing that protocol revision. Protocol
types remain confined to `internal/mcpserver`; domain and application packages
remain protocol-independent.

## Initial slice

The first vertical slice establishes:

- the `ainfra mcp serve --stdio [--project PATH]` command shape;
- newline-delimited stdio transport through the official SDK;
- a default-deny registry containing only typed read-only tools;
- the first versioned `ainfra.version` tool result; and
- unit coverage for dispatch, protocol selection, registry disclosure, and
  structured results.

## Project-isolation slice

The server now resolves a single canonical deployment root and validates its
normal configuration before opening the protocol transport. Conflicting
`--project` and `AINFRA_PROJECT` selections, missing deployments, and symlink
roots fail before any protocol output. Requests cannot replace the selected
root. The default registry exposes that immutable identity through the typed,
read-only `ainfra.project.inspect` tool.

The typed, read-only `ainfra.status` adapter reuses the same retained-status
application logic as the CLI while keeping project selection out of request
arguments. It reads only the fixed startup project and its configured private
run root, preserves inspection-required recovery guidance, and returns typed
`AINFRA-E4601` diagnostics on failure.

Published v1 JSON Schemas are embedded into the binary and exposed through a
closed `ainfra://schemas/v1/...` MCP resource registry. A versioned
`ainfra://contracts/v1` catalog reports supported document, result, and
standard-output versions alongside the known schema URIs and canonical
documentation references. Resource handlers never translate client URIs into
filesystem paths and require neither the source checkout nor network access.

The read-only `ainfra.doctor.deployment` adapter reuses the CLI's resolved
deployment diagnostic core against the startup project and configuration
snapshot. Its closed input has no target, configuration, reconciliation, or
confirmation fields. Failed findings remain typed diagnostics, and tests prove
that a missing runtime directory is reported without creating it.

The compiled-binary suite verifies default registry disclosure, typed results,
rejection of undisclosed mutation tools, and separation of startup diagnostics
from protocol stdout.

The remaining environment, template, run, and sanitized artifact adapters,
bounded concurrency, malformed and oversized frame coverage, capability
groups, and independent mutation authorization remain subsequent Phase 7
slices.
