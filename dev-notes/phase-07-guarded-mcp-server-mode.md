# Phase 7: Guarded MCP server mode

Status: in progress

Phase 7 exposes ainfra's typed application use cases through the common
`ainfra mcp serve --stdio` entry point. The default registry is read-only;
planning and the first independently authorized deployment operation are
available only through explicit capability groups. Destruction additionally
requires its own capability group.

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
The first implemented group, `planning`, discloses exactly two additional
tools. The read-only `ainfra.reconciliation.plan` tool calls the existing
registered local-repair planner with application disabled and returns stable
relative paths, modes, rollback limitations, and the accompanying deployment
diagnosis. Compiled tests prove that planning must be explicitly enabled and
that calling it does not create `.ainfra` or otherwise apply the proposed
repair.

The non-idempotent `ainfra.plan.create` tool creates an apply- or
destroy-intent saved plan through the same application planning core as the
CLI. Its closed input contains only the intent. Project identity,
configuration, executable selection, cache root, run root, and environment
remain fixed at server startup; request cancellation reaches OpenTofu. The
result retains the normal private run evidence and supplies its run ID and plan
digest for review, but the tool neither authorizes nor executes infrastructure
mutation. Its finish audit event safely correlates the newly created run.

The application layer now defines a provider-neutral independent mutation
authorization boundary without enabling a mutation registry. Opaque approval
material is capped at 64 KiB and passed only to a trusted provider fixed at
server startup. The provider returns a typed grant; ainfra independently
requires exact project root, canonical operation, plan ID and SHA-256 digest,
intent, caller, issuer, approval time, and expiry bindings. The issuer must
differ from the caller, apply/deploy grants require apply intent, destroy
requires destroy intent, and future or expired grants fail closed. Provider
errors are sanitized, cancellation is preserved, and only authorization ID,
caller, issuer, and expiry survive as safe evidence. Session verification
always replaces any supplied root with the canonical startup project root.
Adversarial tests cover absent providers and approval material, oversized
material, self-approval, stale grants, every binding mismatch, invalid
operation/intent pairs, cancellation, and provider error leakage.

The `deployment` capability now discloses the first mutation tool,
`ainfra.apply.execute`. Its input is limited to an exact saved-plan ID, caller
identity, and opaque approval material. Before any OpenTofu invocation, ainfra
strictly reads the startup-project plan binding and requires the independent
provider to approve that exact apply operation, plan ID, plan digest, caller,
root, intent, and validity window. Conversational claims and absent providers
fail closed.

After authorization, ainfra exclusively retains a private
`authorization.json` containing only the sanitized authorization identity and
exact binding. Replay cannot overwrite this evidence. The adapter then calls
the same application apply core as the CLI with the startup configuration
snapshot; full plan reverification, operation locking, a final approval-expiry
check immediately before execution, cancellation, child invocation, outcome
evidence, and ambiguous-state recovery remain shared.
Opaque approval bytes are never logged, returned, or retained.

The compiled binary composes an Ed25519 verifier only when the operator adds
`--authorization-trust PATH` at server startup. The absolute, non-symlink,
non-group-writable trust file is strict JSON with this closed shape:

```json
{
  "schemaVersion": 1,
  "issuers": [
    {"id": "operator-1", "publicKey": "BASE64_ED25519_PUBLIC_KEY"}
  ]
}
```

Approval material is a strict JSON envelope containing `schemaVersion: 1`, a
`grant`, and a base64 signature. The closed grant fields are
`authorizationId`, `issuer`, `caller`, `projectRoot`, `operation`, `planId`,
`planDigest`, `intent`, `approvedAt`, and `expiresAt`. The Ed25519 signature is
calculated over the UTF-8 domain prefix
`ainfra-mcp-authorization-v1\n` followed by the exact raw JSON bytes embedded
as the envelope's `grant` value. The verifier signs and strictly decodes those
same bytes, avoiding language-specific canonicalization. Unknown fields,
trailing values, unknown or duplicate issuers, invalid
keys, invalid signatures, malformed times, and cancelled verification fail
closed. The trusted keys are copied into the immutable serving session;
protocol requests cannot select or reload the trust store.

The `destruction` capability is now available only when `deployment` is also
enabled and discloses `ainfra.destroy.execute`. It accepts the same closed
plan/caller/approval input shape but requires a destroy-intent saved plan and
an independent grant whose canonical operation and intent are both exactly
`destroy`. Apply-intent plans and approvals fail before authorization evidence
or child invocation. Successful execution reuses the CLI destroy core with the
startup configuration snapshot, full under-lock plan reverification, immediate
pre-execution expiry check, exclusive sanitized authorization evidence,
cancellation, terminal outcome evidence, and inspection-required recovery.
Neither capability enablement nor an apply grant can authorize destruction.

The compiled-binary suite verifies default registry disclosure, typed results,
rejection of undisclosed mutation tools, and separation of startup diagnostics
from protocol stdout.

The default registry now also includes explicit typed contract inspection.
`ainfra.deployment.inspect` returns a detached semantic view of the validated
startup deployment, including its template reference and ordered native-input
pointers. `ainfra.template.inspect` reads only the startup project's lock,
requires its source/ref binding to match that deployment, verifies the
digest-addressed cache, loads the strict template manifest, and returns the
immutable binding plus engine and inventory declarations. It fails closed when
the lock, binding, cache, digest, or template contract is invalid. Neither tool
returns materialized cache paths, manifest paths, or template roots.

The `planning` capability now additionally discloses
`ainfra.template.plan`. Its closed input selects only `lock` or `update`; the
startup project, cache, source-acquisition policy, and configuration cannot be
overridden. The application layer resolves and validates the candidate through
the same template-lock core as the CLI but does not publish `ainfra.lock`.
Digest-addressed cache materialization is permitted planning evidence. Existing
lock/update preconditions are preserved, unchanged candidates report
`changed: false`, Git acquisition receives MCP cancellation, and invalid
operations return a typed failure.

The `deployment` capability now also discloses
`ainfra.reconciliation.execute`. Reconciliation previews carry a deterministic
plan ID and SHA-256 digest over the startup deployment identity and ordered,
sanitized actions. Execution recomputes that binding, rejects stale plan IDs,
and requires an independently verified grant bound exactly to the
`reconcile` operation and intent, plan ID, digest, caller, startup root, and
expiry. Authorization IDs are single-use within the serving session. The
existing reconciliation planner rechecks the complete action set under the
normal deployment operation lock before writing. Refused, stale, absent,
expired, or replayed approvals cannot create the runtime directory, and action
failures return sanitized evidence without host paths.

The `deployment` capability now additionally discloses
`ainfra.configure.execute`. It accepts only an applied run ID, caller, and
opaque approval. The handler derives the apply-plan digest from strict retained
evidence and requires an independent grant bound to the `configure` operation,
apply intent, plan ID and digest, caller, startup root, and expiry. Sanitized
approval evidence is retained exclusively as
`authorization-configure.json`; it cannot overwrite apply authorization or be
replayed. Configuration now has a fixed-deployment/fixed-settings application
entry point so MCP cannot reload a changed project configuration. The normal
under-lock run reverification, private output/inventory validation, controlled
Ansible Runner adapter, cancellation, outcome evidence, and recovery semantics
remain shared with the CLI. Approval validity and cancellation are checked
again immediately before the first configuration write or child invocation.

Template migration planning and remaining deployment mutations
(template writes and composed deploy)
remain subsequent Phase 7 slices. Final AINFRA-MCP-001 through
AINFRA-MCP-010 conformance closure and end-user documentation also remain.
Undeclared capability groups are rejected by the current startup validator.
