# Phase 1: Go project foundation

## Planning entry — 2026-08-07

Phase 1 establishes the Go project foundation described by the canonical v1
specification. The implementation source of truth is every file under
`spec/doc/v1/`. The Rust-era implementation is not an implementation baseline
or compatibility constraint.

The requested path `spec/docs/v1/` does not exist in this repository. This plan
uses the canonical path `spec/doc/v1/` named by the specification itself.

### Outcome

Phase 1 delivers:

- a pinned root Go module;
- a functional `ainfra` command shell;
- the required `cmd/ainfra` to `command` to `app` dependency direction;
- semantic results with rich, plain, and versioned JSON rendering;
- complete exit-code and diagnostic foundations;
- secure filesystem and subprocess boundaries;
- offline unit, component, adversarial, fuzz, and black-box tests;
- reproducible Linux and macOS builds for amd64 and arm64;
- Go-native developer and quality tooling; and
- this append-only implementation note.

Phase 1 does not implement deployment contracts, doctor, template sources,
lifecycle operations, OpenTofu, Ansible, run evidence, operational logging
sinks, or MCP mode.

### Universal specification-conformance gate

Every acceptance check in this plan has two inseparable parts:

1. verify the implemented code and observable behavior named by the check; and
2. review that implementation against each and every file in `spec/doc/v1/`,
   including the numbered chapters, `README.md`, `roadmap.yaml`, and both SVG
   architecture diagrams.

An acceptance check cannot pass on tests or tool output alone. Its conformance
review must identify which requirements apply to the implemented slice, record
non-applicable later-phase requirements as deferred rather than ignored, and
confirm that the implementation does not obstruct them.

Any delta between implementation and specification must be rectified before
the check passes. The specification is authoritative. Implementation choices,
existing code, tests, scripts, documentation, and this plan must change when
they conflict with it. A genuine specification ambiguity or contradiction
stops the affected work for an explicit specification decision; it is not
silently resolved in favor of the implementation.

The conformance result and any rectification are appended to this phase note at
each acceptance checkpoint.

### Wave 0: primary-agent preparation

Before parallel work begins, the primary agent will:

1. identify the active version line, source branch, target branch, intended
   version, and operation type;
2. verify that `v1.x-dev` is clean and synchronized;
3. create `feat/phase-01-go-foundation` from `v1.x-dev`;
4. select and pin the supported Go version;
5. evaluate standard-library CLI composition against an external parser;
6. select and pin required Go development tools;
7. record candidates, versions, licenses, rationale, and validation evidence;
8. define shared exit-code, result, diagnostic, renderer, process, filesystem,
   and security contracts before parallel work; and
9. run the universal specification-conformance gate for the Wave 0 selections
   and rectify every delta before accepting the foundation.

Standard-library CLI composition is the preferred starting point. A dependency
is selected only if the specified hierarchy, help, completion, or error
behavior would otherwise require substantial private framework code.

### Parallel work assignments

The three agents work in parallel after the Wave 0 contracts are fixed. The
repository policy limits sub-agents to read-only work, so they produce detailed
patch proposals and test designs. The primary agent applies, reconciles, and
validates all mutations.

| Agent | Model | Effort | Responsibility |
|---|---|---:|---|
| A: Go foundation and builds | `gpt-5.6-terra` | Low | Module, tools, scripts, metadata, and cross-builds |
| B: CLI and output contracts | `gpt-5.6-terra` | Medium | Command shell, application layer, results, renderers, diagnostics, and black-box tests |
| C: Security and process boundaries | `gpt-5.6-terra` | Medium | Filesystem and process policy, adversarial fixtures, component tests, and fuzzing |

`gpt-5.6-terra` is the cost-efficient model for all three streams. Low effort
is sufficient for Agent A's primarily mechanical work. Agents B and C use
medium effort because public machine contracts and security boundaries require
more careful reasoning.

### Agent A: Go foundation and builds

Agent A owns `go.mod`, `go.sum`, developer tool configuration, build
configuration, `scripts/validate-all`, and `scripts/release-checks.sh`.

Tasks:

1. initialize the root module with the canonical repository module path;
2. pin the Go version and Go-based development tools;
3. review every direct dependency for purpose, license, maintenance,
   provenance, security posture, and transitive weight;
4. preserve independently runnable native Go commands for build, formatting,
   imports, vet, static analysis, lint, tests, race tests, and coverage;
5. keep `scripts/validate-v1-spec` in the validation chain;
6. make sequencing scripts transparent, fail-fast, and free of hidden network
   bootstrap;
7. integrate applicable vulnerability, security, dependency, secret, and
   shell checks;
8. build with `CGO_ENABLED=0` for Linux and macOS on amd64 and arm64;
9. report version, commit, Go version, build-time policy, and supported
   contract versions without variable timestamps in reproducible bytes; and
10. add no Makefile, third-party task runner, GitHub workflow, or GitHub
    Action.

Agent A acceptance checks:

- the root module builds independently;
- every pinned native tool command is independently runnable;
- all four supported target binaries compile with CGO disabled;
- build metadata is injectable and deterministic;
- validation scripts contain no Rust-era commands; and
- specification validation remains part of the normal validation chain.

Each individual Agent A check also runs the universal conformance review
against every file in `spec/doc/v1/`. Any delta is rectified in favor of the
specification before that check is accepted.

### Agent B: CLI, application, and output contracts

Agent B owns `cmd/ainfra`, `internal/app`, `internal/command`,
`internal/diagnostic`, `internal/output`, `test/blackbox`, and machine-result
fixtures for implemented commands.

Tasks:

1. keep `cmd/ainfra` a composition root that constructs dependencies, calls
   `command.Run`, and maps its result to an exit code;
2. enforce the `cmd/ainfra` to `command` to `app` dependency direction;
3. implement only genuine foundation commands: help and version, including
   their conventional flag forms;
4. expose no placeholder lifecycle command;
5. define exit codes zero through six with their specified meanings;
6. establish stable `AINFRA-E####` diagnostic-code support;
7. place semantic command results and renderers in `internal/output`;
8. render the same semantic result as rich text, stable plain text, and one
   versioned JSON protocol object;
9. keep requested results on stdout and diagnostics on stderr;
10. default non-TTY output to plain text and degrade terminal features safely;
11. report required version and build information through a focused use case;
12. validate every implemented JSON success and failure fixture against the
    published machine-result schema;
13. add golden fixtures for rich, plain, and JSON rendering; and
14. add black-box coverage for help, version, syntax errors, unknown commands,
    rendering, stream separation, exit codes, non-TTY behavior, and
    cancellation at the composition boundary.

Agent B acceptance checks:

- `cmd/ainfra` contains no product or validation logic;
- `command` owns syntax and renderer selection;
- `app` owns use-case sequencing;
- `output` owns semantic results and renderers;
- every implemented command and exit path has an applicable machine fixture;
  and
- black-box tests invoke the compiled executable from outside `internal` in a
  controlled environment with bounded timeouts.

Each individual Agent B check also runs the universal conformance review
against every file in `spec/doc/v1/`. Any delta is rectified in favor of the
specification before that check is accepted.

### Agent C: security, filesystem, and process boundaries

Agent C owns `internal/security`, `internal/exec`, `testdata/bin`, tests beside
those packages, and applicable fuzz targets.

Tasks:

1. implement focused components for canonical paths, containment, symlink
   refusal, regular files, special-file refusal, restrictive permissions,
   environment allowlisting, executable identity, and sensitive-value
   registration;
2. avoid a generic filesystem abstraction and define interfaces at their
   consumer boundaries;
3. make `internal/exec` the only generic child-process runner;
4. require executable, argument array, canonical working directory, allowed
   environment, I/O policy, cancellation context, and sensitive values;
5. never invoke a shell or accept an interpolated command string;
6. resolve executables to regular files and obtain their versions when needed;
7. return typed process outcomes rather than interpreting prose;
8. forward cancellation and termination signals without claiming rollback;
9. preserve API seams for later evidence retention and stream redaction;
10. create fake executables for arguments, streams, exit codes, delays,
    signals, partial writes, environment, injection-shaped inputs, and chunked
    output;
11. test injection, environment isolation, containment, symlinks, special
    files, permissions, executable changes, cancellation, bounded shutdown,
    and partial output; and
12. fuzz argument arrays, path containment, environment filtering, and
    executable resolution where useful.

Agent C acceptance checks:

- no process path reaches a shell;
- injection-shaped values remain opaque arguments;
- filesystem decisions fail closed;
- parent environment variables do not leak by default;
- cancellation reaches the child process;
- fixtures contain no secrets and use reserved invalid identifiers; and
- fuzz regressions are retained as permanent fixtures.

Each individual Agent C check also runs the universal conformance review
against every file in `spec/doc/v1/`. Any delta is rectified in favor of the
specification before that check is accepted.

### Primary-agent integration

The primary agent integrates the proposals in this order:

1. apply the module and toolchain foundation;
2. establish shared result, diagnostic, exit-code, and process types;
3. apply the CLI, application, and output implementation;
4. apply the security and process implementation;
5. verify dependency direction against `package-dependencies.svg`;
6. reject dependency cycles, generic dumping-ground packages, adapter-to-
   adapter calls, and cross-cutting packages that import `app` or `command`;
7. add requirement identifiers to tests or traceability where useful;
8. append actual results, decisions, deltas, and rectifications to this note;
   and
9. update user documentation only for behavior proven by tests.

Integration acceptance requires a review of the combined implementation
against every file in `spec/doc/v1/`. Integration stops and the implementation
is rectified whenever it differs from the authoritative specification.

### Final Phase 1 validation

The final validation includes:

```text
go build ./cmd/ainfra
go fmt ./...
go tool goimports
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
scripts/validate-v1-spec
scripts/validate-all
scripts/release-checks.sh
```

It also checks:

- an offline test run;
- reviewed golden output;
- machine-result schema validation;
- bounded applicable fuzz smoke tests;
- security and vulnerability scans;
- `git diff --check`;
- all four CGO-disabled target builds;
- `ainfra help` and `ainfra version` smoke behavior;
- absence of GitHub workflows and Actions; and
- absence of restored Rust source, tests, or configuration.

Every command and every additional check above is followed by a conformance
review of the affected implementation against each and every file in
`spec/doc/v1/`. A command succeeding does not override a specification delta.
All deltas are rectified with the specification as the authoritative source,
and the check is rerun before it can pass.

### Deferred packages and capabilities

Phase 1 does not create empty placeholders for later responsibilities. The
following remain deferred until their roadmap phase provides a real use case:

- `internal/config`;
- `internal/project`;
- `internal/template`;
- `internal/source`;
- `internal/lock`;
- `internal/tofu`;
- `internal/inventory`;
- `internal/ansible`;
- `internal/evidence`;
- `internal/run`;
- `internal/reconcile`;
- `internal/migration`;
- `internal/logging`;
- runtime schema loading and doctor;
- complete streaming redaction and evidence retention;
- lifecycle commands;
- MCP mode;
- template-authoring tooling; and
- release promotion and publication.

Deferral is checked for architectural compatibility during every universal
conformance review. Later-phase requirements may remain unimplemented, but the
Phase 1 design must not contradict or obstruct them.

## Implementation checkpoint — 2026-08-07

Implementation started from commit `9efba3f` on `v1.x-dev`. The working topic
branch is `feat/phase-01-go-foundation`, targeting reviewed integration into
`v1.x-dev`.

### Implementation-time selections

The Go toolchain is pinned to Go 1.26.5. This is both the toolchain supplied by
the aibox development environment and the current release reported by the
official Go version endpoint on 2026-08-07.

The initial CLI uses standard-library flag composition. The implemented
foundation commands do not require a third-party hierarchy or completion
framework, so an additional runtime dependency would add weight without
providing necessary behavior.

The following development tools are pinned through the `go.mod` tool
mechanism:

| Tool | Module version | License | Purpose |
|---|---|---|---|
| `goimports` | `golang.org/x/tools v0.48.0` | BSD-3-Clause | Import arrangement and formatting |
| `staticcheck` | `honnef.co/go/tools v0.7.0` | MIT | Static analysis |
| `golangci-lint` | `github.com/golangci/golangci-lint/v2 v2.12.2` | GPL-3.0 | Curated lint aggregation; development use only |
| `govulncheck` | `golang.org/x/vuln v1.6.0` | BSD-3-Clause | Reachable Go vulnerability analysis |
| `gosec` | `github.com/securego/gosec/v2 v2.28.0` | Apache-2.0 | Go security analysis |

These are development-only dependencies and are not linked into the ainfra
binary. Their module graph is intentionally recorded in `go.sum`. The direct
runtime implementation uses only the Go standard library. Versions were
resolved from their maintained upstream Go modules, license files were reviewed
from the pinned module content, and the complete graph is covered by the pinned
vulnerability scan. The GPL-licensed lint aggregator remains an unmodified
development tool and is not distributed in the ainfra binary.

### Implemented foundation

The first slice now contains:

- the root Go module;
- `cmd/ainfra` as the sole process exit boundary;
- `internal/command`, `internal/app`, `internal/diagnostic`, and
  `internal/output` with the required dependency direction;
- help and version commands without lifecycle placeholders;
- exit codes zero through six;
- rich, plain, and versioned JSON version rendering;
- schema-shaped successful and failed version envelopes;
- `internal/security` path, file, environment, executable-binding, and exact
  known-value redaction policies;
- `internal/exec` as the sole no-shell child-process runner with explicit
  argv, cwd, environment, sanitized IO, process groups, and cancellation;
- unit, component, fuzz-seed, and compiled-binary black-box tests;
- pinned Go quality and security tools;
- transparent validation, test, release-check, and four-target build scripts;
  and
- CGO-disabled Linux and macOS builds for amd64 and arm64.

### Specification deltas requiring resolution

The full `spec/doc/v1/` review found two contradictions in the machine-output
contract. They have not been resolved by weakening or extending the schema:

1. `07-build-and-quality.md` requires release binaries to report supported
   contract versions, but the closed version result in
   `spec/schemas/v1/machine-output.schema.json` has no field for them.
2. `12-cli-configuration-and-logging.md` requires every command and exit path
   to have a schema assertion, but an unknown command has no legal canonical
   command discriminator in the closed machine-output schema.

The implementation remains schema-valid: version reports only fields accepted
by the published schema, and failures associated with the recognized version
command produce a version envelope with a null result and diagnostic. Unknown
commands use stderr and exit code two. The affected requirements need an
explicit specification amendment before the implementation can claim complete
conformance for supported-contract reporting and unknown-command JSON.

### Security deferrals preserved by the foundation

The executable descriptor records and revalidates a content fingerprint before
execution. Child-tool version compatibility remains deferred until the phase
that selects and invokes each real tool. Exact known sensitive values are
redacted across output chunks now; heuristic credential-shape scanning and
complete evidence retention remain in the hardening phase. These are deferred
capabilities, not waived requirements.

### Checkpoint evidence

At this checkpoint:

- `go build ./cmd/ainfra` passes;
- unit, component, and black-box tests pass;
- the race-enabled suite passes;
- coverage reporting succeeds;
- `go vet`, `staticcheck`, and the curated `golangci-lint` configuration pass;
- `govulncheck` reports no reachable vulnerability;
- `gosec` findings at the intentional process and validated executable-open
  boundaries have narrow adjacent rationale suppressions;
- exact-value redaction is tested across write boundaries;
- argument-injection text remains one opaque argv element; and
- CGO-disabled builds pass for all four required target combinations.

This checkpoint was reviewed against every file in `spec/doc/v1/`. Later-phase
requirements remain explicitly deferred. The two unresolved contradictions
above prevent a final Phase 1 conformance claim until the specification is
clarified; all other identified implementation deltas were rectified in favor
of the specification.

### Per-file conformance review

The implemented slice was checked against each canonical documentation file.

| File | Phase 1 result |
|---|---|
| `README.md` | Conforms: the implementation is a new Go CLI and imports no Rust-era code or contracts. |
| `01-product-boundary.md` | Conforms: the shell exposes no invented infrastructure language, engine replacement, or false lifecycle capability. |
| `02-concepts-and-layouts.md` | Conforms for Phase 1: contained paths, regular-file policy, private directories, and immutable process inputs are established; deployment and run layouts remain deferred. |
| `03-contracts-and-sources.md` | Conforms for Phase 1: child invocation can only use structured argv and explicit paths; manifests, sources, native inputs, and locks remain deferred. |
| `04-cli-and-lifecycle.md` | Conforms for the implemented commands: global format options work before or after `version`, streams are separated, and exit codes zero through six are fixed; lifecycle commands remain deferred. |
| `05-security-and-trust.md` | Conforms for the foundation: executable, argv, environment, cwd, cancellation, permissions, and exact-value redaction policies fail closed; engine version compatibility and heuristic secret scanning remain deferred. |
| `06-software-architecture.md` | Conforms: `cmd/ainfra` composes `command`, which calls `app`; cross-cutting output, diagnostic, security, and exec packages do not import command or app; no generic utility package exists. |
| `07-build-and-quality.md` | Partially blocked: Go 1.26.5, native commands, pinned tools, scans, and four builds are implemented; supported-contract reporting conflicts with the closed version schema as recorded above. |
| `08-template-authoring.md` | Conforms by deferral: no template hooks, plugin system, arbitrary command surface, or premature authoring implementation was added. |
| `09-documentation-and-ai-agents.md` | Conforms: this append-only phase note records selections, results, deferrals, and deltas; no user documentation claims unimplemented behavior. |
| `10-release-engineering.md` | Conforms for Phase 1: work is on a `feat/*` branch from `v1.x-dev`, four target builds are scripted, and GitHub workflows remain prohibited; promotion and release packaging remain deferred. |
| `11-acceptance-migration.md` | Conforms for Phase 1: the standard-library CLI selection and tool versions are recorded, builds cover the required targets, and no Rust-era protocol was ported. |
| `12-cli-configuration-and-logging.md` | Partially blocked: rich, plain, and versioned JSON render from semantic results with stdout/stderr isolation; full configuration and logging are deferred, and unknown-command JSON conflicts with the schema as recorded above. |
| `13-testing-strategy.md` | Conforms for implemented behavior: package, component, adversarial, fuzz-seed, race, coverage, and compiled-binary black-box tests run offline with controlled inputs. |
| `14-go-source-conventions.md` | Conforms after formatting, imports, vet, static analysis, and lint: production code is internal, interfaces are focused, errors are wrapped, and only `cmd/ainfra` exits. |
| `15-mcp-server-mode.md` | Conforms by architecture and deferral: typed application results are reusable and no protocol type enters a domain package; MCP remains Phase 7. |
| `16-market-positioning.md` | Conforms: no portal, Kubernetes control plane, meta-platform, or multi-deployment coordination was introduced. |
| `architecture-overview.svg` | Conforms: the foundation preserves ainfra as the bounded orchestrator and does not collapse engine responsibilities into the CLI. |
| `package-dependencies.svg` | Conforms for created packages: dependency direction follows the diagram and has no cycle. |
| `roadmap.yaml` | Conforms for Phase 1 scope: module, command shell, typed results, process/filesystem boundaries, fixtures, tools, and four builds are present; later phases remain absent. |

The parser originally accepted output controls only after the `version`
command. This review identified the mismatch with their global status in
`12-cli-configuration-and-logging.md`; the parser and black-box test were
rectified to accept the controls before or after the command.
