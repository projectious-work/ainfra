---
title: Usage
weight: 50
---

The current alpha CLI validates local contracts, diagnoses deployments,
resolves immutable local or Git template sources, creates and executes reviewed
OpenTofu plans, configures hosts through Ansible, and retains sanitized
lifecycle evidence. Static help and version inspection do not perform project
discovery, network access, or child-tool execution:

```sh
ainfra help
ainfra help version
ainfra version
ainfra --format json version
```

Validate and diagnose local state with:

```sh
ainfra init example-deployment
ainfra doctor environment
ainfra doctor deployment example-deployment
ainfra doctor template path/to/materialized-template
ainfra doctor all example-deployment
```

Resolve and bind template content with:

```sh
ainfra template lock example-deployment
ainfra template update example-deployment
```

Both mutations accept `--config`, follow trusted configuration precedence,
and emit stable text or JSON results. `template lock` creates the initial
binding and refuses to replace a changed lock. `template update` is the
explicit path for accepting a new source identity, Git commit, or tree digest.
Doctor never resolves a mutable Git ref and remains read-only unless guarded
local reconciliation is explicitly requested.

Create a private, immutable saved plan and apply that exact reviewed plan with:

```sh
ainfra plan example-deployment
ainfra apply example-deployment --plan RUN_ID
```

The plan result supplies `RUN_ID` and the complete follow-up command. Apply
refuses missing, destroy-intent, replayed, or stale plans and never creates an
implicit replacement plan. See [Reviewed plans](reviewed-plans/) for binding,
evidence, JSON-output, and interruption details.

Destruction requires its own reviewed plan and exact run ID:

```sh
ainfra plan example-deployment --destroy
ainfra destroy example-deployment --plan RUN_ID
```

Inspect retained evidence and ambiguous recovery state without reading
OpenTofu state:

```sh
ainfra status example-deployment
ainfra logs example-deployment --run RUN_ID
ainfra doctor run example-deployment
```

The published `v1.0.0-alpha.7` includes reviewed destroy, recovery, retained
logs, and operational logging in addition to output, inventory, Ansible, and
composed deploy commands; see
[Output, inventory, and Ansible](output-inventory-ansible/). Phase 7 adds
guarded MCP serving. The interface is read-only by
default and fixes one project root for the lifetime of the process:

```sh
ainfra mcp serve --stdio --project /path/to/deployment
```

Saved plan creation is discoverable only when explicitly enabled:

```sh
ainfra mcp serve --stdio --project /path/to/deployment \
  --capability planning
```

Enabling a capability does not approve a lifecycle mutation. Deployment and
destruction require independent, operation-bound authorization.

Deployment capability startup requires `AINFRA_MCP_APPROVAL_KEY` containing at
least 32 bytes of verifier key material. Each mutation request must carry an
externally issued `ainfra.approval/v1` artifact signed with HMAC-SHA256. The
artifact binds the canonical root, operation, saved plan ID and digest, intent,
caller, independent approver, issue time, expiry, and nonce. ainfra exposes no
approval-signing command or MCP tool.

The default registry exposes read-only diagnostics, status, sanitized retained
output and inventory, all published v1 schemas, and the normative MCP server
contract. It never exposes raw engine streams.

With an authorization provider configured, `deployment` exposes apply,
configure, convergence-check, and deploy operations. `destruction` additionally
exposes exact reviewed destroy-plan execution and cannot be enabled without
`deployment`. Every stdio request frame is limited to 1 MiB.

Tool and resource execution is bounded to eight concurrent requests. Client
cancellation propagates into planning and lifecycle operations, including any
permitted child process through the existing application execution contract.
