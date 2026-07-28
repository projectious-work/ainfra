---
title: Lifecycle operations
weight: 10
description: Validate, plan, apply, inspect, and destroy infrastructure.
---

## Validate

```sh
uv run ainfra validate path/to/document.yaml
uv run ainfra validate path/to/document.json --format json
```

Validation reads a document but does not resolve secret references or run an
infrastructure engine.

## Check readiness

```sh
uv run ainfra doctor
uv run ainfra doctor --format json
uv run ainfra doctor --input .ainfra/hetzner.input.yaml
```

The input-aware form also checks backend readiness for the intended
environment.

## Plan

```sh
uv run ainfra plan TEMPLATE --input INPUT
```

Treat the plan as sensitive. Review resource ownership, addresses, firewall
rules, image selection, and estimated cost. Apply only the returned plan ID:

```sh
uv run ainfra apply TEMPLATE --input INPUT --approve PLAN_ID
```

The Rust implementation embeds the built-in template and materializes a new
workspace beneath `.ainfra/runs/<PLAN_ID>/workspace/`; it does not execute
OpenTofu in a source checkout. Initialization uses the committed lock file in
read-only mode. The generated variables, plan bytes, and versioned plan record
remain together in the isolated run directory.

Before apply, ainfra revalidates the input and verifies the operation,
template identity and version, environment, canonical input path and bytes,
retained workspace contents, remote backend configuration path and bytes, plan
location, and plan bytes. Any mismatch fails before OpenTofu initialization or
apply begins.

## Read standardized outputs

```sh
uv run ainfra outputs TEMPLATE --run PLAN_ID --format yaml
uv run ainfra outputs TEMPLATE --run PLAN_ID --format json
```

The standardized output is designed for downstream automation. It must not be
confused with raw provider outputs or state. Output collection requires a
successful apply marker for the same exact run and writes
`.ainfra/runs/PLAN_ID/output.json`.

## Configure hosts

```sh
uv run ainfra configure TEMPLATE \
  --run PLAN_ID \
  --known-hosts .ainfra/known_hosts
```

The command generates a run-local inventory using private management
addresses, then runs the retained Ansible playbook. Verify SSH fingerprints
through an independent channel and populate the protected `known_hosts` file
before running it. Use `--check` for an explicit Ansible check-mode pass.

## Destroy

Create the destroy plan first:

```sh
uv run ainfra plan TEMPLATE --input INPUT --destroy
```

Then approve the returned ownership scope:

```sh
uv run ainfra destroy TEMPLATE \
  --input INPUT \
  --approve-destroy PLAN_ID
```

Destroy applies the exact reviewed destroy plan through `tofu apply`; it never
uses an unreviewed direct `tofu destroy` operation. An apply plan cannot
authorize destroy, and a destroy plan cannot authorize apply.

After a disposable test, confirm zero owned resources in both Hetzner and
OpenTofu state. A successful command without this independent check is not
complete teardown evidence.

## Direct-tool escape hatch

`ainfra` remains a thin wrapper. If diagnosis requires direct OpenTofu or
Ansible commands, use the template working directories and preserve the same
input, state, and ownership scope. Never bypass the reviewed-plan or teardown
controls merely for convenience.
