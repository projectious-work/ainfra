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

Project-root resolution, the remaining read-only application adapters,
bounded concurrency, compiled-binary protocol tests, capability groups, and
independent mutation authorization remain subsequent Phase 7 slices.
