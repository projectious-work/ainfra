---
title: Quick Start
weight: 20
---

The Phase 2 CLI can validate a local deployment and an already-resolved local
template without contacting providers or changing infrastructure.

Start by checking local prerequisites:

```sh
ainfra doctor environment
```

Create a minimal deployment directory when starting from scratch:

```sh
ainfra init example-deployment
```

This writes `example-deployment/ainfra.yaml` and an idempotent marked
`.gitignore` entry for local `.ainfra/` evidence. It does not create secrets,
backend resources, or infrastructure. Replace the placeholder local template
reference with the source you intend to lock in Phase 3.

Then diagnose a deployment directory or its manifest:

```sh
ainfra doctor deployment path/to/deployment
ainfra doctor path/to/deployment --format json
```

For a checked-out template source, run:

```sh
ainfra doctor template path/to/template
```

Doctor is read-only unless `--reconcile` is explicitly supplied and its plan
is confirmed. Template acquisition and locking arrive in Phase 3; OpenTofu and
Ansible lifecycle execution follows in later phases. Phase 2 therefore does
not turn a doctor result into infrastructure changes.
