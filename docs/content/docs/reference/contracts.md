---
title: Contracts
weight: 20
description: Versioned schemas at the boundary between templates and consumers.
---

All schemas use JSON Schema draft 2020-12 and reject unknown fields.

| Contract | Source | Purpose |
|---|---|---|
| `InfrastructureTemplate/v1alpha1` | [`schemas/template-manifest.v1alpha1.json`](https://github.com/projectious-work/ainfra/blob/main/schemas/template-manifest.v1alpha1.json) | Template engines, paths, capabilities, and invariants |
| `TemplateInput/v1alpha1` | [`schemas/template-input.v1alpha1.json`](https://github.com/projectious-work/ainfra/blob/main/schemas/template-input.v1alpha1.json) | Non-secret operator intent and credential references |
| `InfrastructureOutput/v1alpha1` | [`schemas/template-output.v1alpha1.json`](https://github.com/projectious-work/ainfra/blob/main/schemas/template-output.v1alpha1.json) | Stable, non-secret handoff to downstream systems |
| `ainfra.plan/v1alpha1` | [`schemas/plan-record.v1alpha1.json`](https://github.com/projectious-work/ainfra/blob/main/schemas/plan-record.v1alpha1.json) | Exact reviewed-plan binding for lifecycle authorization |

## Compatibility

The `apiVersion` is locked through the v1 series. An incompatible `v2` requires
a full migration rather than silent coercion.

## Validation behavior

- Unsupported `apiVersion` and `kind` values fail.
- Unknown properties fail.
- Secret-shaped standard outputs fail.
- Undeclared capabilities fail.
- Template paths cannot escape the template directory.
- Secret references describe a source but do not contain the secret.

Run:

```sh
uv run ainfra validate DOCUMENT
```

Positive and negative fixtures live under `tests/fixtures/contracts/`.

## Reviewed-plan records

A reviewed-plan record binds the intended operation to the selected template
and version, environment identity, canonical input path and bytes, template
tree, and generated OpenTofu plan bytes. Apply and destroy must reject any
changed binding before starting an infrastructure process.

Rust writes `ainfra.plan/v1alpha1` records and can read normal ten-field plan
records created by the Python implementation. Approval identifiers are exactly
20 lowercase hexadecimal characters. Plan files must resolve beneath their
corresponding `.ainfra/runs/<PLAN_ID>/` directory.
