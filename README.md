<div align="center">

<img src="static/logo/ainfra-light.svg" alt="ainfra" width="96" height="96">

# ainfra

**Infrastructure you can inspect, approve, and remove.**

[![Status: early development](https://img.shields.io/badge/status-early_development-E05232)](https://projectious-work.github.io/ainfra-templates/)
[![License: MIT](https://img.shields.io/badge/license-MIT-1d3352)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-projectious--work.github.io-1d3352)](https://projectious-work.github.io/ainfra-templates/)
[![Python: 3.12](https://img.shields.io/badge/python-3.12-546a82)](pyproject.toml)

</div>

---

`ainfra` is the security-first infrastructure provisioning layer for
projectious.work. It defines versioned infrastructure-template contracts and a
thin local wrapper around OpenTofu and Ansible.

```text
ainfra apply -> provisioned target + non-secret output contract
             -> aibox deploys workloads to that existing target
             -> processkit reconciles content inside the workspace
```

The project provisions targets. It does not build workload images, install
Kubernetes or processkit, deploy aibox fleets, or hide OpenTofu and Ansible
behavior.

## Current status

The `v0.1-dev` implementation includes strict `v1alpha1` contracts, local
security gates, reviewed-plan lifecycle controls, sanitized outputs, Ansible
inventory generation, and a disposable Hetzner Kubernetes-ready baseline.

The full path has been exercised on disposable Hetzner infrastructure,
including deterministic SSH trust, Ansible check mode, an idempotent second
apply, and complete teardown. It remains early-development software: review
plans, understand the costs, and keep teardown ready.

## What it protects

- **Reviewed plans.** Apply requires the exact plan identifier that was
  reviewed.
- **Explicit ownership.** Destroy requires the exact resource-scope token.
- **Secret boundaries.** Inputs reference credentials; standardized outputs do
  not contain them.
- **Private defaults.** Public address allocation and management ingress are
  deliberate choices.
- **Visible automation.** OpenTofu owns infrastructure and Ansible owns host
  configuration; direct-tool behavior remains inspectable.
- **Local gates.** Contract, policy, formatting, type, security, and
  infrastructure checks run without GitHub Actions.

## Architecture at a glance

```text
Template + non-secret input
            │
            ▼
  contract and policy checks
            │
            ▼
   reviewed OpenTofu plan
            │
            ▼
 infrastructure + hardened hosts
            │
            ▼
 non-secret output contract
            │
            ▼
 downstream workload deployment
```

## Quick start

Python 3.12 and
[uv](https://docs.astral.sh/uv/) are required.

```sh
uv sync --all-groups
scripts/validate-all
scripts/test-all
uv run ainfra --help
uv run ainfra validate \
  tests/fixtures/contracts/v1alpha1/valid/template-input.json
uv run ainfra doctor
```

For a disposable Hetzner plan, including cost and teardown guidance, follow the
[Quickstart](https://projectious-work.github.io/ainfra-templates/docs/getting-started/quickstart/).

## Documentation

Full documentation lives at
**[projectious-work.github.io/ainfra-templates](https://projectious-work.github.io/ainfra-templates/)**.

| Section | Contents |
|---|---|
| [Quickstart](https://projectious-work.github.io/ainfra-templates/docs/getting-started/quickstart/) | Install, validate, plan, apply, and tear down |
| [Concepts](https://projectious-work.github.io/ainfra-templates/docs/concepts/) | Architecture, security, state, and secrets |
| [Guides](https://projectious-work.github.io/ainfra-templates/docs/guides/) | Lifecycle operations, local gates, template authoring |
| [Reference](https://projectious-work.github.io/ainfra-templates/docs/reference/) | CLI, schemas, Hetzner baseline, acceptance evidence |
| [Contributing](https://projectious-work.github.io/ainfra-templates/docs/contributing/) | Development and documentation workflow |

Build and serve the Hugo + Docsy site locally:

```sh
scripts/build-docs.sh
scripts/serve-docs.sh
```

## Repository layout

```text
src/ainfra/                       Python CLI and lifecycle orchestration
schemas/                          Versioned public JSON Schema contracts
templates/hetzner-kubernetes-baseline/
                                  OpenTofu, cloud-init, and Ansible template
tests/                            Contract, policy, and lifecycle tests
content/ assets/ layouts/ static/ Hugo + Docsy documentation site
scripts/                          Local gates and documentation commands
```

## Contributing

Issues and pull requests are welcome. Start with the
[contributing guide](https://projectious-work.github.io/ainfra-templates/docs/contributing/)
and run `scripts/validate-all` plus `scripts/test-all` before submitting a
change.

## Security

Never place credentials, private keys, state, plans, or generated inventories
in an issue or commit. Review the
[security model](https://projectious-work.github.io/ainfra-templates/docs/concepts/security-model/)
before operating live infrastructure.

## License

[MIT](LICENSE) © Bnaard

Brand and design system ©
[projectious.work](https://github.com/projectious-work/brand). The ainfra mark
is derived from that system.
