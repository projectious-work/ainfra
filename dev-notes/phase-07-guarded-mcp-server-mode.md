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

The closed `ainfra.doctor.run` tool applies the CLI's latest-run evidence
checks to the fixed startup project. It accepts no run ID or path override,
does not inspect OpenTofu state, and reports missing evidence as a typed skip
rather than creating directories or claiming a pass.

The closed `ainfra.doctor.template` tool diagnoses only the verified template
bound to the startup project. It reuses deployment lock/cache facts and the
existing template diagnostic registry, accepts no path or source input, and
reports an absent verified binding as a typed skip without acquisition.

Environment diagnostics now accept caller cancellation through the existing
executable-inspection boundary. MCP captures the read-only environment report
before opening stdio and exposes that immutable result as
`ainfra.doctor.environment`; requests cannot replace configuration, executable
paths, environment values, or log destinations.

The default registry also exposes `ainfra.output.read` and
`ainfra.inventory.read`. These are retained-artifact readers, not adapters for
the mutating CLI output and inventory operations: they never invoke OpenTofu,
acquire an operation lock, or publish a file. Both accept only `runId`, resolve
it beneath the configured private runs root, and require matching v1 run and
apply-plan records bound to the startup deployment. Standardized output is
strictly decoded through the closed, secret-shape-rejecting v1 contract.
Inventory is returned only when its private retained bytes exactly match a
fresh deterministic rendering of that validated output. Traversal,
cross-project bindings, public files, secret-shaped output, and altered
inventory fail closed with typed diagnostics.

All tool and resource handlers now share an eight-request concurrency bound.
Waiting requests remain cancellable, and the bound covers inexpensive
metadata calls as well as retained-artifact and diagnostic reads so a client
cannot bypass it by switching default capabilities. The official SDK remains
responsible for protocol parsing and dispatch.

The stdio adapter places a 4 MiB ceiling on each newline-delimited input frame
before bytes reach the SDK decoder. An oversized frame closes the session with
a sanitized `AINFRA-E5001` failure on stderr and no protocol pollution on
stdout. Compiled-binary tests cover malformed and oversized frames, concurrent
tool calls, log separation, and clean rejection; unit tests prove the exact
concurrency width, cancellation while queued, and frame-boundary behavior.

Admitted tool and resource requests receive a server-local monotonic request
identifier. At `info` level, the existing synchronized operational logger
records paired start and finish events containing only the request ID, tool or
resource URI, fixed deployment name, and safe-format run ID when the input
type has one. Failures are recorded at `error` level. Raw arguments, project
paths, environment values, and result or artifact content are never logged. The
compiled-binary concurrency fixture verifies unique correlations and confirms
that standardized output content and the project root do not enter the audit
sink.

Optional capability selection is a repeatable startup-only
`--capability GROUP` allowlist. Unknown, duplicate, and not-yet-implemented
groups fail before protocol output; the default registry remains unchanged.
The first implemented group, `planning`, discloses exactly one additional
read-only tool: `ainfra.reconciliation.plan`. It calls the existing registered
local-repair planner with application disabled and returns stable relative
paths, modes, rollback limitations, and the accompanying deployment diagnosis.
Compiled tests prove that planning must be explicitly enabled and that calling
it does not create `.ainfra` or otherwise apply the proposed repair.

The compiled-binary suite verifies default registry disclosure, typed results,
rejection of undisclosed mutation tools, and separation of startup diagnostics
from protocol stdout.

The deployment and destruction capability groups, along with independent
mutation authorization, remain subsequent Phase 7 slices and are rejected by
the current startup validator.
