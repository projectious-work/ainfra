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
> validates local contracts, diagnoses deployments, and locks immutable local
> or Git template sources. Infrastructure execution remains under development.

---

ainfra turns infrastructure templates into a reviewed, reproducible lifecycle.
The current alpha establishes the trust boundary before provider execution:
strict contracts, deterministic diagnostics, immutable source resolution,
content-addressed materialization, and drift detection.

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
```

`template lock` resolves the configured local or Git source, verifies its
content, and writes the canonical `ainfra.lock`. Use `template update` when an
intentional source or content change should replace that binding.

## Core Workflow

The shipped alpha supports contract validation, initialization, doctor scopes,
guarded local reconciliation, and immutable template lock/update operations.
Reviewed OpenTofu planning and apply are the next roadmap phase.

## Documentation

The versioned documentation is published at
<https://projectious-work.github.io/ainfra/>.

## Development

## Repository Structure

## Contributing

## License
