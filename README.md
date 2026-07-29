<div align="center">

<img src="docs/static/logo/ainfra-light.svg" alt="ainfra" width="96" height="96">

# ainfra

**Infrastructure you can inspect, approve, and remove.**

[![Status: early development](https://img.shields.io/badge/status-early_development-E05232)](https://projectious-work.github.io/ainfra/)
[![License: MIT](https://img.shields.io/badge/license-MIT-1d3352)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-projectious--work.github.io-1d3352)](https://projectious-work.github.io/ainfra/)
[![Rust: 1.96.1](https://img.shields.io/badge/rust-1.96.1-546a82)](rust-toolchain.toml)

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

The current v0 implementation includes strict `v1alpha1` contracts, local
security gates, reviewed-plan lifecycle controls, sanitized outputs, Ansible
inventory generation, and a disposable Hetzner Kubernetes-ready baseline.

The full path has been exercised on disposable Hetzner infrastructure,
including deterministic SSH trust, Ansible check mode, an idempotent second
apply, and complete teardown. It remains early-development software: review
plans, understand the costs, and keep teardown ready.

## What it protects

- **Reviewed plans.** Apply requires the exact plan identifier that was
  reviewed.
- **Explicit ownership.** Destroy requires the exact reviewed destroy-plan ID.
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

Install the verified release binary, then initialize a separate infrastructure
project:

```sh
curl --proto '=https' --tlsv1.2 --fail --location \
  --proto-redir '=https' \
  https://github.com/projectious-work/ainfra/releases/latest/download/install.sh \
  -o /tmp/ainfra-install.sh
sh /tmp/ainfra-install.sh
ainfra --help
mkdir ../my-infrastructure
(cd ../my-infrastructure && ainfra init --name my-infrastructure)
(cd ../my-infrastructure && ainfra validate)
(cd ../my-infrastructure && ainfra doctor --environment development)
```

For a disposable Hetzner plan, including cost and teardown guidance, follow the
[Quickstart](https://projectious-work.github.io/ainfra/docs/getting-started/quickstart/).

## Documentation

Full documentation lives at
**[projectious-work.github.io/ainfra](https://projectious-work.github.io/ainfra/)**.

| Section | Contents |
|---|---|
| [Quickstart](https://projectious-work.github.io/ainfra/docs/getting-started/quickstart/) | Install, validate, plan, apply, and tear down |
| [Concepts](https://projectious-work.github.io/ainfra/docs/concepts/) | Architecture, security, state, and secrets |
| [Guides](https://projectious-work.github.io/ainfra/docs/guides/) | Lifecycle operations, local gates, template authoring |
| [Reference](https://projectious-work.github.io/ainfra/docs/reference/) | CLI, schemas, Hetzner baseline, acceptance evidence |
| [Contributing](https://projectious-work.github.io/ainfra/docs/contributing/) | Development and documentation workflow |

Build and serve the Hugo + Docsy site locally:

```sh
docs/scripts/build-docs.sh
docs/scripts/serve-docs.sh
```

## Repository layout

```text
src/                              Rust CLI and lifecycle implementation
schemas/                          Versioned public JSON Schema contracts
templates/hetzner-kubernetes-baseline/
                                  OpenTofu, cloud-init, and Ansible template
tests/                            Contract, policy, and lifecycle tests
docs/                            Self-contained Hugo + Docsy site and tooling
scripts/                          Local gates and documentation commands
```

## Contributing

Issues and pull requests are welcome. Start with the
[contributing guide](https://projectious-work.github.io/ainfra/docs/contributing/)
and run `scripts/validate-all` plus `scripts/test-all` before submitting a
change.

Feature branches target `v0.x-dev`. Tested changes are promoted through
`v0.x-pre-release` and `v0.x-release`; published stable releases are merged
into `main`. See the
[branching strategy](https://projectious-work.github.io/ainfra/docs/contributing/branching/)
for the complete lane responsibilities and release rules.

## Security

Never place credentials, private keys, state, plans, or generated inventories
in an issue or commit. Review the
[security model](https://projectious-work.github.io/ainfra/docs/concepts/security-model/)
before operating live infrastructure.

## Rollback

Reinstall a previously verified CLI release with `AINFRA_VERSION` when a
binary rollback is required. Infrastructure rollback is a separate reviewed
operation: create and approve an exact destroy plan, retain the backend state,
and verify provider cleanup. Never remove `.ainfra/` or OpenTofu state to
simulate rollback.

## License

[MIT](LICENSE) © Bnaard

Brand and design system ©
[projectious.work](https://github.com/projectious-work/brand). The ainfra mark
is derived from that system.
