---
title: CLI
weight: 10
description: The complete ainfra command surface.
---

```text
ainfra validate <path> [--format text|json]
ainfra doctor [--format text|json] [--input INPUT]
ainfra plan <template> --input INPUT [--destroy] [--format text|json]
ainfra apply <template> --input INPUT --approve PLAN_ID
ainfra destroy <template> --input INPUT --approve-destroy PLAN_ID
ainfra outputs <template> [--format json|yaml]
ainfra inventory --output OUTPUT --destination DESTINATION
```

## `validate`

Validates an `InfrastructureTemplate`, `TemplateInput`, or
`InfrastructureOutput` document. It does not resolve credentials or mutate
infrastructure.

## `doctor`

Reports local dependency and configuration readiness without mutation.
Providing `--input` adds checks that depend on the target environment's
backend mode and capabilities.

## `plan`

Validates the template and input, invokes OpenTofu planning, applies policy
checks, and emits a reviewable plan ID. The built-in template is materialized
into an isolated per-plan workspace. `--destroy` creates the corresponding
destruction review without bypassing the exact-plan gate.

## `apply`

Applies only the plan identified by `--approve`, from its retained isolated
workspace. A stale, absent, changed, or mismatched identifier fails closed
before an infrastructure process starts.

## `destroy`

Destroys only through the exact destroy plan identified by
`--approve-destroy`. The implementation applies that reviewed plan rather than
running an unreviewed direct destroy command.

## `outputs`

Reads the validated, sanitized `InfrastructureOutput`. It never substitutes raw
OpenTofu state or outputs for that contract.

## `inventory`

Transforms a validated standardized output into an Ansible inventory at an
explicit destination.
