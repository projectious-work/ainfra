# Changelog

All notable changes to ainfra are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
