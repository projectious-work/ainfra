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
checks, and emits a reviewable plan ID. `--destroy` creates the corresponding
destruction review.

## `apply`

Applies only the plan identified by `--approve`. A stale, absent, or mismatched
identifier fails closed.

## `destroy`

Destroys only the resource ownership scope identified by
`--approve-destroy`.

## `outputs`

Reads the validated, sanitized `InfrastructureOutput`. It never substitutes raw
OpenTofu state or outputs for that contract.

## `inventory`

Transforms a validated standardized output into an Ansible inventory at an
explicit destination.
