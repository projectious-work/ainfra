# Changelog

All notable changes to ainfra are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-alpha.9] - 2026-08-22

### Added

- Added the Phase 9 Hetzner Kubernetes baseline with a private network,
  three-node K3s control plane, optional workers, and a temporary
  source-restricted administration bastion.
- Added pinned, checksum-verified K3s and Cloudflare Tunnel installation,
  deterministic native Ansible inventory, and documented pod and service
  network inputs.
- Added disposable live-certification coverage for provisioning, converged
  re-planning, check-mode configuration, bastion removal, tunnel-only SSH,
  exact-plan destruction, and independent provider-side leak detection.
- Defined the future versioned consumer-target projection and signed
  deployment-provenance boundary while reserving live hardware attestation for
  a separate confidential-computing phase.

### Changed

- OpenTofu now receives the allowlisted Hetzner credential while continuing
  to exclude unrelated parent-process secrets.
- The baseline's OpenTofu engine is self-contained and uses protected local
  state for disposable certification; production deployments must lock a
  derivative with a capability-tested encrypted remote backend.
- K3s initializes the first server before joining the remaining nodes and
  binds every node to its declared private address.

### Fixed

- Accepted both current and legacy `ansible-runner --version` output formats.
- Prevented K3s pod and service CIDRs from overlapping the Hetzner private
  network and corrected systemd rendering for initial and joining servers.

### Security

- Retained disabled root and password authentication, operator-verified SSH
  host keys, private node administration, and Cloudflare Service Auth before
  removing the temporary bastion.
- Certified teardown by applying the exact reviewed destroy plan and then
  independently confirming that no Phase 9 Hetzner resources remained.

## [1.0.0-alpha.8] - 2026-08-20

### Added

- Added Phase 8 template-authoring conformance diagnostics for native variable
  declarations, applicable clean-room fixtures, the standard variable
  reference, and required lifecycle documentation.
- Added a self-contained human and AI authoring package covering manifest and
  output schemas, native dependency and variable rules, security, validation,
  disposable live acceptance, and sanitized evidence requirements.
- Added an independently authored provider-free clean-room template whose
  OpenTofu format, init, validate, plan, apply, destroy-plan, and destroy-apply
  lifecycle is executable without provider credentials.

### Changed

- Template doctor failures now produce a typed partial-failure result and a
  non-zero dependency exit while native-tool checks remain explicitly
  delegated to each template's clean-room validation script.
- Published the branded Hugo documentation site, installation guidance,
  template-authoring workflow, clean-room tutorial, and AI completion
  checklist through GitHub Pages.

### Fixed

- Inventory-free templates no longer need artificial output declarations or
  standardized-output fixtures.
- Updated compiled black-box coverage to verify the complete Phase 8 doctor
  summary instead of the pre-Phase 8 finding count.

### Security

- Scoped authoring-file inspection beneath the validated template root and
  retained the provider/tool boundary without adding provider-specific logic.
- Required fresh explicit approval before any billable live lifecycle and
  prohibited live-support claims without disposable evidence and independent
  teardown confirmation.

## [1.0.0-alpha.7] - 2026-08-17

### Added

- Added `ainfra mcp serve --stdio` with a fixed project boundary, read-only
  diagnostics, status, sanitized retained artifacts, and bundled v1 contract
  resources by default.
- Added explicit planning, deployment, and destruction capability groups that
  reuse the typed application lifecycle rather than shelling through the CLI.
- Added externally issued, Ed25519-signed approval artifacts bound to the
  project root, operation, reviewed plan, caller, independent issuer, and
  validity window for every lifecycle mutation.

### Changed

- Reviewed apply and destroy execution can independently verify an authorized
  plan digest and intent before executable discovery or child invocation.
- MCP cancellation now propagates through planning and lifecycle operations,
  while tool and resource execution is capped at eight concurrent requests.

### Security

- Added closed capability registries, separate destruction gating, a 4 MiB
  frame limit, protocol-only stdout, and refusal of raw engine streams.
- Added authorization mismatch, replay, expiry, self-approval, traversal,
  symlink, malformed-frame, oversized-frame, concurrency, cancellation, and
  compiled-binary black-box coverage.

## [1.0.0-alpha.6] - 2026-08-15

### Added

- Added separately reviewed destroy plans and exact-plan destruction through
  `ainfra destroy --plan RUN_ID`, with durable destructive lifecycle evidence.
- Added `ainfra status` and `ainfra logs` for private retained-run inspection,
  typed recovery guidance, combined timelines, source filtering, and guarded
  raw child-stream access.
- Added configurable structured operational logging with verbosity controls,
  private rotating files, local syslog, and correlated records.

### Changed

- Interrupted apply, destroy, and configuration stages now require inspection
  and cannot be repeated automatically.
- Certified-template guidance now requires independent provider-side teardown
  verification rather than treating a successful destroy exit as proof.
- OpenTofu structured error filtering explicitly reports `unavailable` when
  only retained raw streams and child-process facts exist.

### Security

- Added common credential-shape and chunk-safe exact-value redaction across
  operational and child-process boundaries.
- Added owner-only cache-tree verification, reviewed-plan binding
  reverification, directory-scoped log compression, and symlink-resistant
  retained-artifact reads.
- Added hostile cache, permission, evidence, concurrency, compiled-CLI, fuzz,
  race, vulnerability, and static-security coverage.

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
[1.0.0-alpha.6]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.5...v1.0.0-alpha.6
[1.0.0-alpha.7]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.6...v1.0.0-alpha.7
[1.0.0-alpha.8]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.7...v1.0.0-alpha.8
[1.0.0-alpha.9]: https://github.com/projectious-work/ainfra/compare/v1.0.0-alpha.8...v1.0.0-alpha.9
