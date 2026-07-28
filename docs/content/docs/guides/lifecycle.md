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
ainfra doctor --environment development
```

The input-aware form also checks backend readiness for the intended
environment.

## Plan

```sh
uv run ainfra plan TEMPLATE --input INPUT
```

For an initialized project, use the declared environment:

```sh
ainfra plan --environment development
```

Treat the plan as sensitive. Review resource ownership, addresses, firewall
rules, image selection, and estimated cost. Apply only the returned plan ID:

```sh
uv run ainfra apply TEMPLATE --input INPUT --approve PLAN_ID
```

The corresponding project-mode apply is:

```sh
ainfra apply --environment development --approve PLAN_ID
```

For the complete project workflow, retain the same review boundary and pass
that exact ID to `up`:

```sh
ainfra plan --environment development
# Review the saved plan, then:
ainfra up --environment development --approve PLAN_ID \
  --known-hosts ./known_hosts
```

`up` never plans or chooses an approval implicitly. It composes the exact
apply, validated output collection, configuration, and a separate Ansible
check-mode verification stage. If a stage was interrupted after its `started`
event or ended in failure, `up` stops and routes the operator to manual
recovery.

The Rust implementation embeds the built-in template and materializes a new
workspace beneath `.ainfra/runs/<PLAN_ID>/workspace/`; it does not execute
OpenTofu in a source checkout. Initialization uses the committed lock file in
read-only mode. The generated variables, plan bytes, and versioned plan record
remain together in the isolated run directory.

Before apply, ainfra revalidates the input and verifies the operation,
template identity and version, environment, canonical input path and bytes,
retained workspace contents, remote backend configuration path and bytes, plan
location, and plan bytes. Project mode additionally binds the canonical project
root plus the exact `ainfra.yaml` and `ainfra.lock` bytes. Any mismatch fails
before OpenTofu initialization or apply begins.

Each completed plan has an immutable `run.json` and append-only events beneath
`.ainfra/runs/<PLAN_ID>/events/`. Mutating and configuration phases record a
`started` event before execution and then a `succeeded` or sanitized `failed`
event. A missing terminal event means the process may have been interrupted;
do not infer success or retry automatically.

Inspect local state without contacting a provider:

```sh
ainfra status --environment development
ainfra status --environment development --format json
```

Status validates event ordering and bindings, ignores untrusted run entries,
and emits deterministic next steps. A partial, stale, corrupt, or legacy state
routes to manual recovery instead of an apply or destroy command.

Teardown uses the same two-step review:

```sh
ainfra plan --environment development --destroy
# Review the dedicated destroy plan, then:
ainfra down --environment development --approve-destroy DESTROY_PLAN_ID
```

An apply-plan ID can never authorize `down`, and `down` never runs a direct
unreviewed destroy. A successful destroy process first records
`destroy-applied`; only a subsequent empty `tofu state list` for the reviewed
backend records the environment as `destroyed`.

## Read standardized outputs

```sh
uv run ainfra outputs TEMPLATE --run PLAN_ID --format yaml
uv run ainfra outputs TEMPLATE --run PLAN_ID --format json
ainfra outputs --environment development --run PLAN_ID
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
ainfra configure --environment development \
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
ainfra plan --environment development --destroy
```

Then approve the returned ownership scope:

```sh
uv run ainfra destroy TEMPLATE \
  --input INPUT \
  --approve-destroy PLAN_ID
ainfra destroy --environment development \
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
