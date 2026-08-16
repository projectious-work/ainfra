<div align="center">

<img src="docs/static/logo/ainfra-light.svg" alt="ainfra" width="96"
  height="96">

# ainfra

**Immutable, auditable infrastructure lifecycle orchestration.**

[![Status: alpha](https://img.shields.io/badge/status-alpha-1d3352)](SECURITY.md)
[![License: MIT](https://img.shields.io/badge/license-MIT-1d3352)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-projectious--work.github.io-E05232)](https://projectious-work.github.io/ainfra/)

</div>

---

> [!NOTE]
> **Maturity:** alpha — the v1 specification is complete. The current release
> validates and locks deployments, creates reviewed OpenTofu saved plans, and
> applies only the exact verified plan. Later lifecycle stages remain under
> development.

---

ainfra turns infrastructure templates into a reviewed, reproducible lifecycle.
The current alpha carries that trust boundary through reviewed provider
execution: strict contracts, deterministic diagnostics, immutable source
resolution, content-addressed materialization, bound saved plans, drift
detection, and durable lifecycle evidence.

## Why ainfra

## Install

Download the archive for Linux or macOS from the latest GitHub release, verify
it against the published checksum manifest, and place `ainfra` on your `PATH`.
See the [installation guide](https://projectious-work.github.io/ainfra/docs/installation/)
for source builds and signature verification.

## Quick Start

```sh
ainfra doctor environment
ainfra init example-deployment
ainfra template lock example-deployment
ainfra doctor deployment example-deployment
ainfra plan example-deployment
# Review the plan result, then run its exact next command:
ainfra apply example-deployment --plan RUN_ID
```

`template lock` resolves the configured local or Git source, verifies its
content, and writes the canonical `ainfra.lock`. Use `template update` when an
intentional source or content change should replace that binding.

## Core Workflow

The shipped `v1.0.0-alpha.6` supports contract validation, initialization,
doctor scopes, guarded local reconciliation, immutable template lock/update
operations, standardized output, deterministic inventory, native Ansible
configuration, composed deployment, and exact reviewed apply and destroy
execution. Retained status and logs provide inspection guidance after failed,
cancelled, or ambiguous mutations without interpreting OpenTofu state.

Guarded MCP server mode is the current in-progress roadmap phase and is not
part of the published alpha yet.

## Documentation

The versioned documentation is published at
<https://projectious-work.github.io/ainfra/>.

## Development

## Repository Structure

## Contributing

## License
