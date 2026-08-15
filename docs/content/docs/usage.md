---
title: Usage
weight: 50
---

The current alpha CLI validates local contracts, diagnoses deployments, resolves
immutable local or Git template sources, and creates and applies reviewed
OpenTofu plans. Static help and version inspection do not perform project
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

The published `v1.0.0-alpha.5` includes output, inventory, Ansible, and composed
deploy commands; see
[Output, inventory, and Ansible](output-inventory-ansible/). Reviewed destroy,
recovery, retained logs, and operational logging are being completed for
Phase 6. Guarded MCP serving remains a later roadmap phase.
