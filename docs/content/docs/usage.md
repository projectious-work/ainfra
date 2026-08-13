---
title: Usage
weight: 50
---

The Phase 3 CLI validates local contracts, diagnoses deployments, and resolves
immutable local or Git template sources. Static help and version inspection do
not perform project discovery, network access, or child-tool execution:

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

The help surface also lists later lifecycle commands. OpenTofu planning,
apply, Ansible execution, destruction, and MCP serving are not implemented in
this release.
