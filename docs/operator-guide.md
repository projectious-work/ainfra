# Operator guide

## Validate contracts

```sh
uv run ainfra validate path/to/document.yaml
uv run ainfra validate path/to/document.json --format json
```

Validation reads a document but does not resolve its secret references or run
an infrastructure engine.

## Run local gates

```sh
scripts/validate-all
scripts/test-all
```

Missing required tooling is a failure, not a skipped gate. Security scanners
and infrastructure engines enter these scripts in their delivery milestones.

## Lifecycle safety

`plan`, `apply`, `destroy`, and `outputs` are intentionally guarded during the
foundation milestone. They must not be used until the wrapper implements
reviewed-plan application, output redaction, exact approvals, ownership
checks, and direct-tool equivalents.

Live Hetzner verification is never part of routine validation. It is the
final delivery step and requires a fresh, explicit user approval that names
the project, resources, intended lifetime, and estimated cost.
