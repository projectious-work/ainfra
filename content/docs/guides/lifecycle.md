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

## Read standardized outputs

```sh
uv run ainfra outputs TEMPLATE --format yaml
uv run ainfra outputs TEMPLATE --format json
```

The standardized output is designed for downstream automation. It must not be
confused with raw provider outputs or state.

## Generate inventory

```sh
uv run ainfra inventory \
  --output .ainfra/output.json \
  --destination .ainfra/inventory.yml
```

The inventory generator selects management addresses from a validated output
contract. Verify SSH host keys before using it.

## Destroy

Create the destroy plan first:

```sh
uv run ainfra plan TEMPLATE --input INPUT --destroy
```

Then approve the returned ownership scope:

```sh
uv run ainfra destroy TEMPLATE \
  --input INPUT \
  --approve-destroy SCOPE_TOKEN
```

After a disposable test, confirm zero owned resources in both Hetzner and
OpenTofu state. A successful command without this independent check is not
complete teardown evidence.

## Direct-tool escape hatch

`ainfra` remains a thin wrapper. If diagnosis requires direct OpenTofu or
Ansible commands, use the template working directories and preserve the same
input, state, and ownership scope. Never bypass the reviewed-plan or teardown
controls merely for convenience.
