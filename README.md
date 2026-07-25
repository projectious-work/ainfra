# ainfra-templates

`ainfra-templates` is the security-first infrastructure provisioning layer for
projectious.work. It defines versioned infrastructure-template contracts and a
thin local wrapper around OpenTofu and Ansible.

```text
ainfra apply -> provisioned target + non-secret output contract
             -> aibox deploys workloads to that existing target
             -> processkit reconciles content inside the workspace
```

The project provisions targets. It does not build workload images, install
processkit, deploy aibox fleets, or hide OpenTofu and Ansible behavior.

## Current status

Milestone 0 establishes the public `v1alpha1` schemas, fixture-driven
validation, documentation, and Python/uv CLI foundation. Infrastructure
lifecycle commands remain guarded until their implementation milestones.

## Local development

Python 3.12 and
[uv](https://docs.astral.sh/uv/) are required.

```sh
uv sync --all-groups
scripts/validate-all
scripts/test-all
uv run ainfra --help
uv run ainfra validate \
  tests/fixtures/contracts/v1alpha1/valid/template-input.json
```

There are no GitHub Actions or workflow files. Every project gate is exposed
through a local script or the CLI.

## Contracts

- [`InfrastructureTemplate`](schemas/template-manifest.v1alpha1.json)
- [`TemplateInput`](schemas/template-input.v1alpha1.json)
- [`InfrastructureOutput`](schemas/template-output.v1alpha1.json)
- [Contract semantics](docs/architecture.md)
- [Security model](docs/security-model.md)
- [State and secrets](docs/state-and-secrets.md)
- [Template authoring](docs/authoring-templates.md)
- [Operator guide](docs/operator-guide.md)
- [Acceptance matrix](docs/acceptance-matrix.md)

All examples and fixtures are non-secret. Secret values must be referenced,
never placed in committed manifests, command arguments, ordinary outputs, or
logs.

## Branches and releases

`main` contains releases. Work for a minor release integrates through a
version-specific branch such as `v0.1-dev`; feature branches start from and
merge back to that branch. A reviewed release merge promotes the development
branch to `main`, after which the exact release commit receives an annotated
semantic-version tag.

## License

MIT. See [LICENSE](LICENSE).
