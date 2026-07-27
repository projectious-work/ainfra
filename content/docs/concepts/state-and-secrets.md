---
title: State and secrets
weight: 30
description: Backend requirements, secret references, and recovery.
---

## State

Non-disposable environments must use a capability-validated remote backend
providing encryption at rest, locking, version recovery, TLS, and access
control. S3-compatible services qualify only after integration tests prove
those capabilities.

Local state is limited to inputs explicitly marked disposable and produces a
prominent warning. The wrapper does not provision, repair, or silently migrate
a backend.

Backend configuration stays outside version control. The repository ignores
local state, plans, `.terraform/`, and `.ainfra/`.

## Secret references

Contracts accept references:

```yaml
projectTokenRef:
  type: environment
  name: HCLOUD_TOKEN
```

They do not accept a token value. Credentials are resolved only when an
underlying tool needs them and are passed through the child-process
environment, never command arguments.

Raw state, plans, logs, inventories, and live-verification evidence remain in
ignored local storage. Only an explicitly sanitized report may be promoted
into documentation.

## Recovery

Recovery procedures must identify:

- the backend version to restore;
- the target state lineage;
- the operator authorizing the action;
- the reviewed direct OpenTofu command;
- the verification that follows recovery.

Automatic state repair and migration are intentionally out of scope.
