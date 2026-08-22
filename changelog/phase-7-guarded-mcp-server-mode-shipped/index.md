# Phase 7 shipped: Guarded MCP server mode

> ainfra now exposes read-only inspection and explicitly authorized, plan-bound lifecycle operations through MCP over stdio.


Phase 7 exposed the existing typed application lifecycle through a guarded MCP
adapter without weakening its authorization or evidence contracts.

## What shipped

- An official-SDK MCP server over newline-delimited JSON-RPC stdio.
- A fixed project boundary and read-only default tool registry.
- Sanitized doctor, status, output, inventory, schema, and documentation
  resources.
- Explicit planning and lifecycle capability allowlists.
- Independent authorization for deployment and destruction tools.
- The same locking, plan-binding, recovery, and evidence rules as the CLI.

Read the complete
[Phase 7 development note](https://github.com/projectious-work/ainfra/blob/v1.x-dev/dev-notes/phase-07-guarded-mcp-server-mode.md)
or inspect
[v1.0.0-alpha.7](https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.7).


---
Source: https://projectious-work.github.io/ainfra/changelog/phase-7-guarded-mcp-server-mode-shipped/index.md
