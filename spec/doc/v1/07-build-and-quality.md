## Direct developer interface

A Makefile or third-party task runner is not required. Go's native commands are
the canonical developer interface and MUST remain independently runnable:

```text
go build ./cmd/ainfra
gofmt -w .
go tool goimports -w .
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

Go-based development tools SHOULD be pinned with the `go.mod` tool mechanism
and invoked with `go tool` where the selected Go version supports it. The aibox
development environment pins non-Go tools. Tool configuration files are
committed beside the source.

A small checked-in script MAY sequence the complete release or integration
check when several ecosystems are involved. Such a script is a transparent
convenience, not a second build system: it prints each underlying command,
stops on failure, accepts no hidden network bootstrap, and documents external
prerequisites. Documentation and release checklists MUST name the native
commands so contributors never need to reverse-engineer an alias.

## Toolchain

- a supported Go version is pinned in `go.mod` and the developer environment;
- `go.mod` and `go.sum` are committed;
- `CGO_ENABLED=0` is used for release builds unless a reviewed dependency
  requires otherwise;
- Linux and macOS release targets are `amd64` and `arm64`;
- Windows compilation and support are out of scope;
- unit tests MUST not require network, cloud credentials, OpenTofu, or Ansible;
- integration tests explicitly declare external prerequisites.

The suite design, isolation rules, regression policy, and release matrix are
defined in the [testing strategy](13-testing-strategy.md).

The aibox developer container SHOULD consume language-scoped add-ons requested
in `projectious-work/aibox#337`, including Go infrastructure, supply-chain, and
release groups.

## Required checks

### Formatting and static analysis

```text
gofmt
go tool goimports
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
```

The `golangci-lint` configuration MUST contain a curated low-noise set. A
warning is either fixed or suppressed at the narrowest location with rationale.
Repository-wide blanket exclusions are prohibited.

### Tests

```text
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

- **AINFRA-BUILD-001:** race testing MUST cover all supported packages on
  Linux in the development/release gate.
- **AINFRA-BUILD-002:** coverage is a change-risk signal, not a gameable global
  percentage. Security, parsers, locks, inventory, run transitions, and exec
  boundaries require direct tests.
- **AINFRA-BUILD-003:** fuzz targets MUST cover YAML/JSON documents, source
  references, archive/path handling, redaction chunking, OpenTofu output, and
  Ansible event parsing.
- **AINFRA-BUILD-004:** fuzz smoke runs use a bounded release budget; longer
  campaigns may run separately.

### Integration tiers

1. **Unit:** no external executables or network.
2. **Process contract:** fake `tofu`, `git`, and `ansible-runner` executables
   assert argv, cwd, environment, cancellation, and redaction.
3. **Local engine:** real OpenTofu with `-backend=false` fixtures and real
   Ansible Runner local/check-mode fixtures; no cloud mutation.
4. **Disposable live:** explicit credentials and cost approval; template-owned
   provision, configure, idempotence, destroy, and independent verification.

Tier 4 MUST never run as an incidental default check.

## Security and dependency tools

Required local/release checks:

- `govulncheck ./...` for reachable Go vulnerabilities;
- `gosec ./...` for Go security analysis;
- `osv-scanner` for repository and lockfile dependency advisories;
- `gitleaks` for repository/history secret scanning according to release policy;
- `syft` for SBOM generation;
- `grype` for release artifact and optional user-built-image scanning;
- `cosign` for release checksum/signature or attestation workflows;
- `shellcheck` for shell scripts;
- `hadolint` for the optional Dockerfile.

Tools described as optional in early drafts are part of the intended aibox
development package set. A missing tool MAY be temporarily reported as an
environment blocker during bootstrap, but release-check MUST require the final
approved set.

## Dependency policy

- prefer the standard library when it yields clear, testable code;
- every direct dependency requires a stated purpose;
- dependencies MUST be pinned by `go.mod`/`go.sum` and reviewed for license,
  maintenance, provenance, and transitive weight;
- avoid importing OpenTofu internals or Ansible internal Python APIs;
- upgrades require tests, vulnerability checks, and changelog review;
- release SBOMs MUST cover the Go binary and document external runtime
  prerequisites separately.

## Build metadata

Release binaries MUST report version, commit, build time policy, Go version,
and supported contract versions. Reproducibility SHOULD avoid embedding
variable timestamps in binary bytes; human release metadata can carry the
publication time externally.
