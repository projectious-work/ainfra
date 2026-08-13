---
title: Quick Start
weight: 20
---

The Phase 3 CLI can validate a local deployment, resolve and lock a local or
Git template source, and diagnose the resulting immutable binding without
contacting infrastructure providers.

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
reference with the source you intend to lock.

Resolve and materialize the source, then create its canonical lock:

```sh
ainfra template lock example-deployment
```

For Git sources this is the only step that resolves the requested mutable ref.
It records the immutable commit and verified content digest. For intentional
source, ref, or content changes, use the explicit update path:

```sh
ainfra template update example-deployment
```

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
is confirmed. It verifies the lock, contained cache entry, digest, and local
source drift without reacquiring mutable Git refs. OpenTofu and Ansible
lifecycle execution follows in later phases, so this release does not turn a
doctor result into infrastructure changes.
