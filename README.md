<div align="center">

<img src="docs/static/logo/ainfra-light.svg" alt="ainfra" width="96"
  height="96">

# ainfra

**The agent-native execution boundary for reviewed infrastructure.**

[![Status: alpha](https://img.shields.io/badge/status-alpha-1d3352)](SECURITY.md)
[![Release: v1.0.0-alpha.9](https://img.shields.io/badge/release-v1.0.0--alpha.9-E05232)](https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.9)
[![License: MIT](https://img.shields.io/badge/license-MIT-1d3352)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-projectious--work.github.io-E05232)](https://projectious-work.github.io/ainfra/)

[Contributing](CONTRIBUTING.md) · [Security](SECURITY.md) ·
[Releases](https://github.com/projectious-work/ainfra/releases)

</div>

---

ainfra lets AI agents and humans operate infrastructure through standard,
inspectable OpenTofu and Ansible templates. Agents create and adapt templates;
ainfra validates, locks, plans, authorizes, executes, and records them through
MCP or CLI. It provides a clear approval and recovery boundary without hiding
the native tools.

ainfra is also the infrastructure provisioning component of the
[projectious.work](https://projectious.work/) stack.

## Features

- MCP-first agent operations with read-only defaults, explicit capabilities,
  independently authorized mutations, and equivalent CLI behavior.
- Immutable local and Git template sources, pinned by revision and content
  digest.
- Reviewed OpenTofu apply and destroy plans with exact, replay-resistant
  execution.
- Strict, versioned inputs and non-secret standardized outputs.
- Deterministic Ansible inventory, configuration, and convergence checks.
- Durable, redacted lifecycle logs with interruption and recovery diagnostics.
- Linux and macOS release archives for AMD64 and ARM64, with SBOMs, checksums,
  and a signed checksum manifest.

## Current Status

The v0 line culminated in `v0.1.0`, the original Rust implementation. It is a
legacy line retained for existing users; only narrowly scoped maintenance is
expected, and new deployments should evaluate the v1 line.

The Go-based v1 rewrite is under active alpha development. Phases 0 through 9
are released as `v1.0.0-alpha.9`: contracts, diagnostics, immutable sources,
reviewed apply and destroy, output and Ansible workflows, recovery hardening,
guarded MCP mode, template-authoring conformance, and the first provider-backed
Hetzner private K3s baseline are implemented. Phase 9 passed a cost-approved
disposable lifecycle including convergence, Cloudflare tunnel-only access,
temporary-bastion removal, exact destruction, and an independent leak check.
See the
[alpha 9 release notes](https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.9)
for the shipped artifacts, SBOMs, checksums, and Sigstore verification bundle.

## Installation

The official installer selects the latest Linux or macOS release for AMD64 or
ARM64 and verifies its checksum before installation:

```sh
curl -fsSL https://raw.githubusercontent.com/projectious-work/ainfra/v1.x-release/scripts/install.sh | bash
```

It installs to `~/.local/bin` by default. Pin a release or choose another
destination with `VERSION` and `INSTALL_DIR`:

```sh
curl -fsSL https://raw.githubusercontent.com/projectious-work/ainfra/v1.x-release/scripts/install.sh \
  | VERSION=1.0.0-alpha.9 INSTALL_DIR=/usr/local/bin bash
```

Checksum verification is mandatory. If Cosign is installed, the installer
also verifies the manifest's Sigstore identity automatically. Set
`VERIFY_SIGNATURE=1` to require that verification or follow the
[installation guide](https://projectious-work.github.io/ainfra/docs/installation/)
for manual and source-install options.

## Usage

For an agent, bind the read-only MCP server to one deployment. Enable planning
or mutation capabilities only for the session that needs them:

```sh
ainfra mcp serve --stdio --project example-deployment
```

For direct CLI operation, check the host, initialize a deployment, and bind its
configured template:

```sh
ainfra doctor environment
ainfra init example-deployment
# Edit example-deployment/ainfra.yaml to select a local or Git template.
ainfra template lock example-deployment
ainfra doctor deployment example-deployment
```

Once the template and its native OpenTofu inputs are ready, create a saved
plan. Review the reported actions and use the exact run ID returned by the
command:

```sh
ainfra plan example-deployment
ainfra apply example-deployment --plan RUN_ID
```

ainfra refuses to apply a missing, changed, stale, replayed, or destroy-intent
plan. Start with the [Quick Start](https://projectious-work.github.io/ainfra/docs/quick-start/)
before provisioning real infrastructure.

## Documentation

The maintained documentation is published at
[projectious-work.github.io/ainfra](https://projectious-work.github.io/ainfra/).
It includes the quick start, installation and configuration guides, the
[product rationale and agent/execution division of labour](https://projectious-work.github.io/ainfra/docs/why-ainfra/),
[template-authoring guide](https://projectious-work.github.io/ainfra/docs/templates/),
an [AI authoring checklist](https://projectious-work.github.io/ainfra/docs/template-authoring-ai/),
reviewed-plan contracts, operational references, change log, and roadmap.

Build or serve the site locally with:

```sh
docs/scripts/build.sh
docs/scripts/serve.sh --port 1314
```

## Architecture at a glance

```text
         Agent or human intent
                 │
          MCP or CLI adapter
                 │
                 ▼
Template source + native, non-secret inputs
                    │
                    ▼
       contracts, policy, immutable lock
                    │
                    ▼
          reviewed OpenTofu saved plan
                    │
                    ▼
       infrastructure + configured hosts
                    │
                    ▼
 standardized output + inventory + evidence
```

OpenTofu owns infrastructure changes and Ansible owns host configuration.
ainfra's shared application core coordinates those tools, verifies the
boundaries between lifecycle steps, and records sanitized evidence; it is not
another infrastructure language or a general-purpose cluster manager.

## Development

Build and run the local validation and test suites with:

```sh
go build ./cmd/ainfra
scripts/validate-all
scripts/test-all
docs/scripts/build.sh
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for prerequisites, branch conventions,
the local-only quality gates, and the guarded release process.

## Repository structure

| Path | Purpose |
|---|---|
| `cmd/`, `internal/` | Go CLI and lifecycle implementation |
| `spec/` | Normative v1 specification, schemas, examples, and roadmap |
| `docs/` | Hugo end-user documentation and projectious.work theme |
| `scripts/` | Local validation, installer, and release tooling |
| `test/` | Black-box CLI and lifecycle fixtures |

## Contributing

Issues and pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md)
and run the relevant local checks before submitting a change.

## Security

Do not disclose credentials, state, plans, private infrastructure details, or
suspected vulnerabilities in a public issue. Follow [SECURITY.md](SECURITY.md)
to report security concerns privately.

## License

[MIT](LICENSE) © Bnaard

Brand and design system ©
[projectious.work](https://github.com/projectious-work/brand). The ainfra mark
is derived from that system.
