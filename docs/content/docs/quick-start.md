---
title: Quick Start
weight: 20
---

The current alpha CLI can validate a local deployment, lock a local or Git
template source, create a reviewed OpenTofu plan, and apply only that exact
saved plan.

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
source drift without reacquiring mutable Git refs.

Create a reviewed apply plan:

```sh
ainfra plan path/to/deployment
```

Review the structural action counts and digests in the result. Then copy its
exact next command, for example:

```sh
ainfra apply path/to/deployment \
  --plan 20260814T120000Z-0123456789abcdef0123456789abcdef
```

Immediately before OpenTofu starts, ainfra reverifies the deployment, native
input files, template binding, workspace, OpenTofu executable and version, and
saved-plan bytes. Ansible execution and standardized output follow in Phase 5.
