## Direct developer interface

A Makefile or third-party task runner is not required. Go's native commands are
the canonical developer interface and MUST remain independently runnable:

```text
go build ./cmd/ainfra
go fmt ./...
git ls-files -z '*.go' | xargs -0 go tool goimports -w
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
scripts/validate-v1-spec
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
go fmt ./...
git ls-files -z '*.go' | xargs -0 go tool goimports -w
git diff --exit-code -- '*.go'
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
```

`go fmt ./...` is the canonical recursive package formatter. The separate
`goimports` command enumerates every tracked Go file recursively. The final
diff command turns either formatter's change into a failing check. This avoids
non-portable assumptions that `gofmt -w .` or `goimports -w .` recursively
traverse the module.

The `golangci-lint` configuration MUST contain a curated low-noise set. A
warning is either fixed or suppressed at the narrowest location with rationale.
Repository-wide blanket exclusions are prohibited.

### Tests

```text
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
```

Suite ownership, required coverage, integration dependencies, isolation, and
release applicability are defined only in the
[testing strategy](13-testing-strategy.md). This chapter defines how to invoke
the Go tools, not a second test taxonomy.

### Specification contracts

```text
scripts/validate-v1-spec
```

This maintainer and release command runs the pinned `check-jsonschema` tool
with JSON Schema's non-GPL format dependencies, offline through the locked
development environment. It validates every v1 schema against its Draft
2020-12 metaschema, validates every maintained example through an explicit
schema mapping, keeps standard format assertions enabled, and proves through
negative fixtures that malformed `date-time` and `uri` values fail. Adding a
schema or example requires updating the explicit mapping; the release review
checks completeness.

This script is not an end-user prerequisite. The Go binary embeds the schemas
and applies equivalent structural and format validation during ordinary command
loading and doctor checks.

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
and supported contract versions. In JSON mode, supported contract versions
MUST use the closed structure defined for the `version` result in
[`machine-output.schema.json`](../../schemas/v1/machine-output.schema.json).
Reproducibility SHOULD avoid embedding variable timestamps in binary bytes;
human release metadata can carry the publication time externally.
