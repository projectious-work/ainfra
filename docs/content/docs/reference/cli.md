---
title: CLI
weight: 10
description: The complete ainfra command surface.
---

```text
ainfra init [--name NAME] [--template TEMPLATE]
  [--environment ENVIRONMENT] [--format text|json]
ainfra validate [<path-or-template>] [--input INPUT]
  [--format text|json]
ainfra doctor [--input INPUT | --environment ENV] [--format text|json]
ainfra plan [<template> --input INPUT | --environment ENV]
  [--destroy] [--format text|json]
ainfra apply [<template> --input INPUT | --environment ENV]
  --approve PLAN_ID
ainfra destroy [<template> --input INPUT | --environment ENV]
  --approve-destroy PLAN_ID
ainfra outputs [<template> | --environment ENV] --run PLAN_ID
  [--format json|yaml]
ainfra configure [<template> | --environment ENV] --run PLAN_ID
  --known-hosts FILE [--check]
ainfra status --environment ENV [--format text|json]
ainfra inventory --output OUTPUT --destination DESTINATION
```

## `init`

Initializes the current directory without contacting a provider or invoking
OpenTofu or Ansible. It creates `ainfra.yaml`, `ainfra.lock`, and one example
under `environments/`, then adds `.ainfra/` to `.gitignore`.

Initialization preflights every primary file and refuses to overwrite any of
them. There is intentionally no force flag. The lock pins the selected
built-in template's name, version, source, and content digest.

## `validate`

Validates an `InfrastructureTemplate`, `TemplateInput`, or
`InfrastructureOutput` document. It does not resolve credentials or mutate
infrastructure. With no target, it discovers the nearest ancestor
`ainfra.yaml` and validates the project, lockfile, environment references,
template digest, contracts, and policy.

## `doctor`

Reports local dependency and configuration readiness without mutation.
Providing `--input` adds checks that depend on the target environment's
backend mode and capabilities. `--environment` selects project mode and checks
the configured input plus compatible OpenTofu and Ansible versions. The two
selectors are mutually exclusive.

## `plan`

Validates the template and input, invokes OpenTofu planning, applies policy
checks, and emits a reviewable plan ID. The built-in template is materialized
into an isolated per-plan workspace. `--destroy` creates the corresponding
destruction review without bypassing the exact-plan gate.

Project mode uses `--environment ENV` and discovers the nearest ancestor
`ainfra.yaml`. Explicit compatibility mode remains
`TEMPLATE --input INPUT`; mixing selectors fails before any subprocess.

## `apply`

Applies only the plan identified by `--approve`, from its retained isolated
workspace. A stale, absent, changed, or mismatched identifier fails closed
before an infrastructure process starts.

Project plans use the `ainfra.plan/v1alpha2` protocol and bind the canonical
project root plus the exact bytes of `ainfra.yaml` and `ainfra.lock`. Legacy
explicit records remain `v1alpha1`; records cannot cross-authorize between
the two modes.

## `destroy`

Destroys only through the exact destroy plan identified by
`--approve-destroy`. The implementation applies that reviewed plan rather than
running an unreviewed direct destroy command.

## `outputs`

Collects OpenTofu output for one exact successfully applied run, rejects
sensitive or ownership-inconsistent values, and emits the validated,
sanitized `InfrastructureOutput`. It never substitutes raw state or raw engine
output for that contract.

## `configure`

Builds a private-address inventory from one run's validated output and invokes
the retained Ansible playbook. `--known-hosts` must name an independently
verified, non-group/world-writable host-key file. `--check` adds Ansible check
mode and diff without changing the reviewed infrastructure plan.

## `inventory`

Transforms a validated standardized output into an Ansible inventory at an
explicit destination.

## `status`

Reads only validated project files and durable local run evidence. It never
invokes OpenTofu, Ansible, credential providers, DNS, or provider APIs.

The report classifies the latest project-bound run as `planned`, `applied`,
`output-collected`, `configured`, `destroy-planned`, `destroyed`, `partial`,
`stale`, `corrupt`, `legacy`, or `none`. Started operations without a matching
success or failure event are `partial`; they are never assumed safe to retry.
Machine output uses `ainfra.status/v1alpha1` and includes sanitized checks and
the next safe command or manual-recovery route.
