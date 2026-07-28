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

### Python-prototype recovery

Treat an old Python checkout and its `.ainfra/` directory as read-only
evidence. Start with:

```sh
ainfra legacy inspect --root /path/to/old/ainfra-checkout --format json
```

The report can validate legacy record structure and saved plan bytes, but it
cannot prove whether apply or destroy ran, which backend was active at that
time, or which resources currently exist. Its lifecycle state is therefore
always `unknown`.

Before any recovery action:

1. preserve the original checkout, `.ainfra/`, state, and backend metadata;
2. identify the exact backend, backend version, and state lineage through
   operator-reviewed backend-native tooling;
3. identify the operator authorizing recovery and record a reviewed direct
   OpenTofu command;
4. compare `tofu state list` with provider-owned resources and ownership
   labels;
5. initialize a new ainfra project only after its configuration references the
   same verified backend and intended environment;
6. create and review a fresh project-bound plan—never reuse a legacy plan ID;
7. independently verify provider resources and state after apply or teardown.

For legacy local state, continue using the preserved Python checkout or an
operator-approved OpenTofu state recovery procedure until the environment is
retired or its state has been deliberately moved. ainfra does not copy,
rewrite, import, or infer that state automatically.
