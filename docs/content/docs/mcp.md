---
title: Guarded MCP server
weight: 55
---

MCP is ainfra's primary agent interaction surface. ainfra serves its typed
application operations to MCP clients over stdio while the CLI remains a
complete equivalent interface for humans, CI, and recovery.
See [Why ainfra exists](../why-ainfra/) for the division between agent
reasoning and governed execution.
The server is bound to one deployment at startup, is read-only by default, and
never treats tool annotations or conversational claims as authorization.

## Start a read-only server

Run the server from a deployment directory:

```sh
ainfra mcp serve --stdio
```

Or bind it explicitly:

```sh
ainfra mcp serve --stdio --project path/to/deployment
```

Protocol frames are the only data written to stdout. Diagnostics and
operational logs use stderr or the configured logging sinks. The startup
project, configuration, executable paths, cache, run storage, environment
allowlist, and logging destinations cannot be overridden by tool arguments.

The default registry provides read-only tools for:

- build and contract versions;
- project, deployment, and verified template inspection;
- retained status, standardized output, and inventory;
- deployment, run, template, and environment diagnostics; and
- published schemas and contract resources.

## Enable optional capabilities

Additional tools must be allowlisted when the server starts:

```sh
ainfra mcp serve --stdio --capability planning
ainfra mcp serve --stdio --capability deployment
ainfra mcp serve --stdio \
  --capability deployment \
  --capability destruction
```

`planning` adds reconciliation, template lock/update/migration, and reviewed
OpenTofu planning. Planning may create private run evidence or
digest-addressed template cache entries, but it does not apply infrastructure
or publish a template lock.

`deployment` adds independently authorized reconciliation, template writes,
apply, configure, and composed deploy. `destruction` adds destroy execution,
requires `deployment`, and accepts only an approval with destroy intent.
Enabling a capability makes tools discoverable; it does not approve a
mutation.

## Configure signed approvals

Mutation tools require a trust file selected by the operator at startup:

```sh
ainfra mcp serve --stdio \
  --project path/to/deployment \
  --capability deployment \
  --authorization-trust path/to/mcp-trust.json
```

The trust file must be a private, non-symlink JSON file:

```json
{
  "schemaVersion": 1,
  "issuers": [
    {
      "id": "operator-1",
      "publicKey": "BASE64_ED25519_PUBLIC_KEY"
    }
  ]
}
```

An approval is a strict JSON envelope containing `schemaVersion: 1`, the raw
`grant` object, and a base64 Ed25519 signature. Sign the UTF-8 bytes
`ainfra-mcp-authorization-v1\n` followed by the exact raw JSON bytes used as
the envelope's `grant` value.

The grant binds all authority-relevant inputs:

```json
{
  "authorizationId": "approval-2026-08-16-1",
  "issuer": "operator-1",
  "caller": "agent-1",
  "projectRoot": "/absolute/canonical/deployment",
  "operation": "apply",
  "planId": "REVIEWED_PLAN_ID",
  "planDigest": "sha256:...",
  "intent": "apply",
  "approvedAt": "2026-08-16T12:00:00Z",
  "expiresAt": "2026-08-16T12:05:00Z"
}
```

The issuer must differ from the caller. Approval must be current and match the
startup root, canonical operation, plan identity, digest, and intent exactly.
Opaque approval bytes are never logged, returned, or retained. Sanitized
authorization evidence records only the authorization ID, caller, issuer,
binding, and expiry.

## Reviewed mutation workflow

For infrastructure mutations, use the planning capability to create a saved
plan and present its run ID and digest for independent review. Restarting with
both `planning` and `deployment` is allowed, but the same agent request cannot
approve the plan it created. Apply, deploy, and destroy always consume an
existing reviewed plan; no mutation tool creates an implicit replacement.

Template lock/update and reconciliation previews similarly return stable plan
IDs and digests. Their execution tools recompute the complete plan under the
deployment operation lock before writing. Changed source content or stale
preconditions require a new preview and approval.

## Operational safety

- Keep trust files and private keys outside the deployment and source tree.
- Enable only the capabilities needed for the current session.
- Use short approval expiries and unique authorization IDs.
- Treat an interrupted mutation as inspection-required; consult `ainfra.status`
  and retained logs before retrying.
- Do not expose the stdio server through a network bridge. Network transports
  are outside the v1 threat model.
- Close the MCP session normally so the stdio child can shut down cleanly.

The MCP result envelope is versioned independently from the transport. Clients
should inspect `apiVersion`, `tool`, `ok`, `result`, and `diagnostics` rather
than parse display text.
