# Changelog

All notable changes to ainfra are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-alpha.5] - 2026-08-14

### Added

- Added strict standardized OpenTofu output collection and deterministic
  Ansible inventory generation with typed applicability results.
- Added controlled `output`, `inventory`, `configure`, and composed `deploy`
  commands with stable text and v1 JSON results.
- Added Ansible configuration and independent exact-host, zero-change
  check-mode verification using retained Runner event evidence.

### Changed

- Extended reviewed plan bindings to include native Ansible variable files and
  the deployment's SSH `known_hosts` input.
- Added ordered deploy-stage reports, including explicit `not-applicable`
  records for infrastructure-only templates.
- Added the output, inventory, Ansible, convergence, and evidence workflow to
  user documentation.

### Security

- Added closed child environments, forced SSH host-key checking, populated
  bound `known_hosts` enforcement, shell-free Runner invocation, and private
  atomic run artifacts.
- Added secret-shaped output refusal, corrupt and symlinked Runner evidence
  refusal, cancellation propagation, replay refusal, and deployment operation
  locking.
- Added compiled black-box, race, schema, vulnerability, static-security, and
  independent requirement-sweep coverage.

## [1.0.0-alpha.4] - 2026-08-14

### Added

- Added `ainfra plan` with collision-resistant run IDs, private saved plans,
  immutable review records, and sanitized structural action summaries.
- Added `ainfra apply --plan RUN_ID` for exact reviewed-plan execution with
  durable started, terminal, cancellation, and inspection-required evidence.
- Added stable text and v1 JSON results for planning and apply execution.

### Changed

- Extended the deployment lifecycle from immutable template resolution through
  controlled OpenTofu initialization, planning, and reviewed apply.
- Added reviewed-plan usage, recovery, concurrency, and sensitive-evidence
  guidance to the documentation.

### Security

- Bound deployment manifests, ordered native inputs, template content,
  OpenTofu executable bytes and version, and saved-plan bytes before apply.
- Added deployment operation locking, replay refusal, destroy-intent separation,
  stale-binding detection, closed child environments, and shell-free argument
  execution.
- Added tampering, cancellation, evidence-failure, concurrent-operation, race,
  vulnerability, static-security, and compiled black-box coverage.

## [1.0.0-alpha.3] - 2026-08-13

### Added

- Added immutable local and Git-subdirectory template acquisition through
  `ainfra template lock` and explicit `ainfra template update`.
- Added the normative template-tree digest, verified private content-addressed
  cache, canonical lock publication, and lock/cache/source doctor findings.
- Added trusted cache and Git configuration for template mutations.

### Changed

- Expanded deployment and aggregate doctor scopes to verify immutable source
  bindings, cached content, digests, and local source drift.
- Updated documentation and conformance evidence for the Phase 3 boundary.

### Security

- Added contained source resolution, argument-array Git execution, immutable
  commit checkout, cache verification, and adversarial path/file-type checks.
- Prevented credential persistence by redacting source userinfo and sensitive
  query values from lockfiles, diagnostics, invocations, and machine output.
- Added cancellation, poisoned-cache, malformed-lock, traversal, race,
  vulnerability, and static-security coverage.

## [1.0.0-alpha.2] - 2026-08-13

### Added

- Added strict deployment and local-template contract validation with
  deterministic project discovery and typed configuration precedence.
- Added `ainfra doctor` environment, deployment, template, retained-run, and
  aggregate scopes with stable text and JSON findings.
- Added guarded local reconciliation and conflict-first minimal deployment
  initialization.

### Changed

- Expanded Phase 2 documentation, schemas, examples, and conformance evidence
  for the contracts-and-doctor boundary.
- Updated the development environment and processkit tooling used for release
  preparation and verification.

### Fixed

- Preserved complete failure evidence when an individual reconciliation action
  remains unsuccessful.
- Made compiled black-box builds resolve Go's effective module cache while
  retaining offline execution.
- Replaced ambiguous shell guards in container and release validation paths
  with explicit fail-closed conditionals.

### Security

- Enforced contained, non-symlink, regular-file inputs for deployments,
  templates, retained runs, and local repairs.
- Added adversarial, fuzz, race, secret, vulnerability, and static-security
  evidence for Phase 2 behavior.

## [1.0.0-alpha.1] - 2026-08-12

### Added

- Introduced the normative ainfra v1 specification, schemas, examples, and
  machine-readable CLI result contracts.
- Added the Go CLI foundation with version, help, validation, planning,
  execution, and guarded MCP server lifecycle behavior.
- Added Linux and macOS release archives for amd64 and arm64.
- Added SPDX JSON SBOMs, SHA-256 checksums, and keyless Sigstore signing for
  the release checksum manifest.
- Added the owner-reviewed macOS/Linux host container gate with Docker runtime
  hardening, SBOM generation, vulnerability scanning, and retained evidence.
- Added the versioned Hextra documentation site and Phase 1 roadmap material.

### Security

- Added secret, dependency, Go vulnerability, static-security, container,
  artifact, and release-integrity gates.
- Bound release signatures to `info@projectious.work` through GitHub OIDC,
  Fulcio, and Rekor, with verification pinned to the GitHub issuer.
- Isolated release build, module, Python, Syft, and host-gate mutable state.

### Changed

- Established the v1 development, pre-release, release, and stable promotion
  lanes.
- Promoted MCP server mode into the v1 product contract.
- Standardized local-only validation and release execution without hosted CI.

[1.0.0-alpha.1]: https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.1
[1.0.0-alpha.2]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.1...v1.0.0-alpha.2
[1.0.0-alpha.3]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.2...v1.0.0-alpha.3
[1.0.0-alpha.4]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.3...v1.0.0-alpha.4
[1.0.0-alpha.5]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.4...v1.0.0-alpha.5
