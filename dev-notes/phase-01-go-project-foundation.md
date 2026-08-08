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

## Accepted baseline change — 2026-08-07

PR #48 and PR #49 were accepted and merged into the retained specification
branch, then integrated into this topic branch as commits `35c43d9` and
`e5016cb`. The exact Phase 1 specification baseline is now `e5016cb`.

PR #48 adds `17-spec-driven-development-cycle.md`. Its process governs the
remainder of Phase 1 and supersedes the informal repetition of conformance
checks in the initial plan. PR #49 resolves the version-result contradiction by
adding a required closed `supportedContractVersions` object, a positive
fixture, and focused negative fixtures.

### Baseline inventory

The baseline consists of:

- all 17 numbered chapters and `README.md` under `spec/doc/v1/`;
- `roadmap.yaml` and both architecture diagrams;
- all eight schemas under `spec/schemas/v1/`;
- every positive example under `spec/examples/v1/`;
- every negative fixture under `spec/tests/v1/`;
- the governing decisions listed by `spec/doc/v1/README.md`; and
- the Phase 1 acceptance journey for four Linux/macOS builds.

The implementation-time selections remain Go 1.26.5, standard-library CLI
composition, and the pinned development tools recorded above.

### Phase 1 requirement inventory and disposition

The complete normative inventory is retained by disposition so no requirement
silently disappears from the cycle.

| Requirement set | Disposition for Phase 1 | Planned evidence or authority |
|---|---|---|
| `AINFRA-DEV-001`–`012` | Covered by the amended cycle | This baseline, matrix, task graph, independent reviews, gap closure, and retained phase record |
| `AINFRA-CODE-001`–`003`, `010`–`011` | Covered | Formatting, imports, vet, static analysis, lint, package tests, source review |
| `AINFRA-CLI-002`–`005` | Covered where a foundation command exists | CLI unit and black-box stream, format, error, and cancellation tests |
| `AINFRA-CLI-001`, `006`–`010` | Deferred by roadmap to deployment and engine phases | No deployment-bound or engine command ships in Phase 1 |
| `AINFRA-OUTPUT-001`–`008`, `010`–`013` | Covered for help/version; engine/raw-output clauses are accepted non-applicability | Renderer unit tests, golden-style assertions, black-box schema validation, amended version fixture |
| `AINFRA-SEC-010`–`015`, `024`–`027` | Covered at the generic process boundary or explicitly preserved for hardening | No-shell argv, environment, contained cwd, executable fingerprint, signal, permissions, and split-value redaction tests |
| `AINFRA-SEC-001`–`006`, `020`–`023`, `030`–`043` | Authorized roadmap deferral to source, lifecycle, and hardening phases | Phase 1 exposes no source, plan, state, evidence, SSH, cache, or Docker operation |
| `AINFRA-TEST-010`, `012`–`018` | Covered where the implemented slice applies | Offline tests, four cross-builds, race, coverage, fuzz seeds, schema formats, secret-safe fixtures |
| `AINFRA-TEST-001`–`003`, `011` | Covered as standing policy; no regression or required-suite skip exists | Reviewed tests and fail-closed scripts |
| `AINFRA-DOC-010`–`013` | Covered where Phase 1 applies | Append-only phase note and roadmap remains in progress; release review is deferred |
| `AINFRA-DOC-001`–`005` | Authorized roadmap deferral to template authoring | No AI template-authoring package ships in Phase 1 |
| `AINFRA-IMPL-001`–`003` | Covered for current selections; child-tool minimums are deferred | Selection ledger and no child-tool command |
| `AINFRA-REL-001`, `003`, `007`, `009` | Covered for topic-branch development and checks | Clean committed checkpoints, fail-closed scripts, branch identity |
| `AINFRA-REL-002`, `004`–`006`, `008` | Authorized deferral to release packaging | No release or tag is produced in Phase 1 implementation |
| `AINFRA-PROD-001`–`008`, `AINFRA-NONGOAL-001`–`009` | Architectural constraints covered; complete product journeys deferred | No meta-language, engine replacement, hidden mutation, portal, or coordination layer |
| All `AINFRA-LAYOUT-*`, `CONTRACT-*`, `SOURCE-*`, and `LOCK-*` | Authorized roadmap deferral except reusable path/file/process primitives | Contracts/doctor and immutable sources are Phases 2 and 3 |
| All `AINFRA-INIT-*`, `DOCTOR-*`, `RECON-*`, and `MIGRATE-*` | Authorized roadmap deferral | Phase 2 and later; no placeholder command ships |
| All `AINFRA-PLAN-*`, `APPLY-*`, `INV-*`, `ANS-*`, `DESTROY-*`, `LOGS-*`, and `LOG-*` | Authorized roadmap deferral | Infrastructure lifecycle and hardening are Phases 4–6 |
| All `AINFRA-MCP-*` | Authorized roadmap deferral | MCP is Phase 7 and remains an adapter over typed application results |
| All `AINFRA-TPL-*` | Authorized roadmap deferral | Template authoring and conformance is Phase 8 |

Schemas and examples remain executable contracts. The Phase 1 implementation
directly exercises the machine-output schema, the positive version example,
all negative machine-output fixtures through `validate-v1-spec`, and the
roadmap schema. Other schemas and examples remain tracked inputs for their
authorized phases and are validated without inventing runtime support.

### Amended task graph and ownership

```text
accepted baseline e5016cb
        |
        +-- impact analysis and amended plan
        |
        +-- independent plan-conformance review
        |
        +-- version result and renderer correction
        |
        +-- affected unit, black-box, and schema verification
        |
        +-- independent implementation-conformance review
        |
        +-- gap closure and repeated affected verification
```

The primary agent remains integration owner and owns all repository mutations.
The independent reviewer receives read-only access, does not rely on the
implementer summary, and must inspect code, tests, schemas, examples, generated
output, documentation, this phase note, roadmap status, and validation
evidence. Rollback is the removal of the uncommitted post-baseline correction;
the accepted specification commits themselves are not rewritten.

### Accepted change impact

The version correction affects `internal/app`, `internal/output`, renderer
tests, CLI black-box tests, generated JSON, schema validation, and the prior
phase-note conflict. It does not change security, process execution, exit
codes, project discovery, or any deferred lifecycle package.

The implementation now constructs the three required version arrays with the
current v1 values, renders them in human output, and emits them in the closed
JSON structure. Completed output work must be reverified against
`AINFRA-OUTPUT-001`–`013`, the updated schema, positive example, and every new
negative fixture before the baseline change is closed.

## Independent plan review and gap closure — 2026-08-07

An independent principal conformance reviewer using `gpt-5.6-sol` at medium
effort inspected the full baseline and withheld plan acceptance. The deeper
model class was selected because chapter 17 classifies conformance review as
high-risk synthesis rather than bounded mechanical work. No TeamMember binding
was returned by routing, so the reviewer used an ephemeral principal identity.

The review found incomplete dispositions, insufficient task traceability,
unresolved machine-output questions, incomplete acceptance journeys, missing
governing-source evidence, non-executable validation descriptions, incomplete
change impact, insufficient rollback, and missing routing evidence.

### Process deviation

The supported-contract code correction was applied after recording the amended
DAG but before its independent plan review completed. This violated the new
ordering in `AINFRA-DEV-006`. The change remains uncommitted and frozen. It is
not accepted as conformant. The gap is closed by recording the deviation,
amending the plan, obtaining an accepted independent plan review, then rerunning
the affected implementation and evidence from the accepted baseline.

### Governing sources and company standards

The nine governing decision bodies named by `spec/doc/v1/README.md` are not
present in this repository, its processkit index, or searchable repositories
available to this workspace. The exact `README.md` at baseline `e5016cb` is the
canonical incorporation of their effects and is therefore the review input:

- native OpenTofu and Ansible orchestration;
- a from-scratch simple Go CLI;
- Linux/macOS binaries and a user-buildable Dockerfile, without Windows or a
  published image;
- closed standardized output;
- layered CLI configuration;
- equivalent deployment paths and conditional Ansible;
- runtime JSON Schema format enforcement;
- company branching and release promotion; and
- read-only MCP after lifecycle and hardening.

The applicable company/project standards are `AGENTS.md`, the branching rules
in chapter 10, Effective Go, Go Code Review Comments, Go documentation
conventions, Semantic Versioning, and Keep a Changelog. If review of the full
external decision bodies is required beyond their canonical incorporated
effects, the phase remains blocked until immutable copies are supplied.

### Corrected requirement dispositions

Each row below has exactly one chapter-17 disposition.

| Requirement | Result | Rationale, task, and evidence |
|---|---|---|
| `AINFRA-DEV-001` | covered | T0 records exact baseline and inventory |
| `AINFRA-DEV-002` | covered | Affected work is frozen on material ambiguity |
| `AINFRA-DEV-003` | covered | This matrix retains schemas, examples, journeys, and requirements |
| `AINFRA-DEV-004` | covered | T0 freezes ownership; primary agent integrates |
| `AINFRA-DEV-005` | covered | Fast/low for tooling, Terra/medium for CLI/security, Sol/medium for review |
| `AINFRA-DEV-006` | blocked by an unresolved decision | Independent amended-plan review must accept this correction |
| `AINFRA-DEV-007` | covered | Specification deltas are explicit and implementation is frozen |
| `AINFRA-DEV-008` | covered | PR #49 impact and affected reverification are recorded below |
| `AINFRA-DEV-009` | deferred with explicit authority | It is the post-integration gate later in this same mandated cycle |
| `AINFRA-DEV-010` | deferred with explicit authority | It governs gap closure after implementation review |
| `AINFRA-DEV-011` | deferred with explicit authority | Roadmap remains `in_progress` until completion approval |
| `AINFRA-DEV-012` | deferred with explicit authority | Final retained record and retrospective occur at phase completion |
| `AINFRA-CONFIG-001` | deferred with explicit authority | Roadmap Phase 2 implements typed effective configuration |
| `AINFRA-CONFIG-002` | deferred with explicit authority | Roadmap Phase 2 implements parsing and deterministic merging |
| `AINFRA-CONFIG-003` | deferred with explicit authority | Roadmap Phase 2 implements configuration schema enforcement |
| `AINFRA-CONFIG-004` | deferred with explicit authority | Roadmap Phase 2 implements project-layer restrictions |
| `AINFRA-CONFIG-005` | deferred with explicit authority | Roadmap Phase 2 implements provenance diagnostics |
| `AINFRA-CONFIG-006` | deferred with explicit authority | Roadmap Phase 2 implements configuration file behavior |
| `AINFRA-CONFIG-007` | deferred with explicit authority | Roadmap Phase 2 implements project configuration restrictions |
| `AINFRA-CONFIG-008` | deferred with explicit authority | Roadmap Phase 2 implements discovery behavior |
| `AINFRA-CONFIG-009` | deferred with explicit authority | Roadmap Phase 2 implements doctor environment output |
| `AINFRA-CONFIG-010` | deferred with explicit authority | Roadmap Phase 2 implements deterministic effective output |
| `AINFRA-OUTPUT-001` | requires a specification change | Help is a command but has no machine-result discriminator; T8 |
| `AINFRA-OUTPUT-002` | covered | T2/T5 assert plain output for non-TTY stdout |
| `AINFRA-OUTPUT-003` | covered | T2/T5 assert one closed JSON object |
| `AINFRA-OUTPUT-004` | covered | T3 redacts before output; no plan/state content exists |
| `AINFRA-OUTPUT-005` | partially satisfied | T5 has reviewed literal fixtures; committed golden files remain to add |
| `AINFRA-OUTPUT-006` | not applicable, with rationale | No progress rendering exists in the Phase 1 command slice |
| `AINFRA-OUTPUT-007` | covered | Diagnostics and result streams are independently asserted |
| `AINFRA-OUTPUT-008` | not applicable, with rationale | No child command is exposed through the CLI and no raw mode exists |
| `AINFRA-OUTPUT-010` | requires a specification change | Help and unknown invocation cannot satisfy the current discriminator schema; T8 |
| `AINFRA-OUTPUT-011` | covered | PR #49 changed the draft baseline before v1 conformance completion |
| `AINFRA-OUTPUT-012` | not applicable, with rationale | No engine report exists in Phase 1 output |
| `AINFRA-OUTPUT-013` | covered | T2/T5 emit and schema-check all three required arrays |
| `AINFRA-SEC-010` | covered | T3/T5 prove structured argv and no shell |
| `AINFRA-SEC-011` | deferred with explicit authority | Generic fingerprint exists; tool version/binding begins with first real-tool phase |
| `AINFRA-SEC-012` | covered | T3/T5 prove explicit allowlisted environment and non-nil empty environment |
| `AINFRA-SEC-013` | covered | T3/T5 canonicalize and contain cwd before execution |
| `AINFRA-SEC-014` | covered | T3/T5 stream through exact-value redaction and separate streams |
| `AINFRA-SEC-015` | deferred with explicit authority | Enforcement applies when engine credential mechanisms ship |
| `AINFRA-SEC-024` | deferred with explicit authority | Exact values work; common credential-shape scanning is roadmap Phase 6 |
| `AINFRA-SEC-025` | covered | T3/T5 verify owner-only directories and files |
| `AINFRA-SEC-026` | deferred with explicit authority | Structured engine evidence begins in Phases 4–6 |
| `AINFRA-SEC-027` | deferred with explicit authority | Engine evidence profiles begin in Phases 4–6 |
| `AINFRA-TEST-010` | covered | T5 default tests use no network, cloud, OpenTofu, or Ansible |
| `AINFRA-TEST-011` | not applicable, with rationale | No required external integration suite can skip in Phase 1 |
| `AINFRA-TEST-012` | partially satisfied | T5/T7 cross-build all targets; native macOS smoke remains blocked on environment |
| `AINFRA-TEST-013` | partially satisfied | Fixtures are synthetic; final gitleaks artifact scan remains T7 |
| `AINFRA-TEST-014` | deferred with explicit authority | Document/source/archive/engine fuzzing belongs to owning Phases 2–6 |
| `AINFRA-TEST-015` | partially satisfied | T5 will run the exact bounded fuzz command below and retain failures |
| `AINFRA-TEST-016` | covered | T5 runs race tests on every supported Go package on Linux |
| `AINFRA-TEST-017` | covered | T5 directly tests positive and negative security/process behavior |
| `AINFRA-TEST-018` | covered | T5 validates positive and malformed `date-time`/`uri` fixtures |
| `AINFRA-DOC-012` | not applicable, with rationale | Phase 1 is not marked shipped |
| `AINFRA-DOC-013` | deferred with explicit authority | Release review occurs only after phase completion |
| `AINFRA-REL-001` | not applicable, with rationale | No release is being built by this implementation cycle |
| `AINFRA-REL-007` | not applicable, with rationale | No promotion branch is advanced |
| `AINFRA-IMPL-003` | deferred with explicit authority | Child-tool minimums are selected at first real-tool integration |

All remaining source, lock, lifecycle, engine, evidence, logging, MCP, and
template-authoring requirements retain the individual roadmap deferrals listed
by their exact ID families in the earlier inventory. They are not claimed as
covered by reusable primitives. Product and non-goal requirements are design
constraints only; the Phase 1 evidence proves absence of forbidden layers and
does not claim the later product journeys.

### Acceptance-journey inventory

Phase 1 must provide evidence for:

- chapter 11 journey 13: version machine JSON and exit-code compatibility;
- the build portion of journey 18: all four supported binaries;
- the Dockerfile portion of journey 18: a maintained root Dockerfile, lint,
  local build, and scan, now assigned to T4 because no later roadmap phase
  authorizes its omission;
- applicable security negatives: argument injection, symlink/traversal,
  special files, permissions, child executable changes, environment leakage,
  split known values, and cancellation; and
- native-or-documented build-and-smoke evidence for each supported target.

The Dockerfile and native macOS smoke evidence are open plan gaps, not silent
skips. Phase 1 cannot become shipped until they are implemented or explicitly
deferred through an accepted baseline change.

### Requirement-to-task-to-evidence map

| Task | Owner/reviewer | Requirement IDs | Exact evidence |
|---|---|---|---|
| T0 baseline and frozen interfaces | Primary / independent plan reviewer | `AINFRA-DEV-001`–`008`, `IMPL-001`–`002`, `REL-009` | Baseline commit, matrix, plan review, branch log |
| T1 module, tools, and builds | Cora with Terra/low / conformance reviewer | `CODE-001`–`003`, `TEST-016`, chapter 7 toolchain | Native commands, pinned module, four build artifacts |
| T2 CLI, app, diagnostic, output | Ephemeral CLI engineer with Terra/medium / conformance reviewer | `CLI-002`–`005`, `OUTPUT-001`–`013` by dispositions above | Unit/black-box tests and machine fixtures |
| T3 security and process boundaries | Ephemeral security engineer with Terra/medium / Sol reviewer | `SEC-010`–`015`, `024`–`027` by dispositions above | Adversarial unit/component tests, race, gosec |
| T4 optional Dockerfile delivery | Primary / Sol reviewer | Product delivery, journey 18, Docker security threats | Dockerfile, hadolint, local build, syft/grype scan |
| T5 tests and contract validation | Primary / Sol reviewer | `TEST-010`–`018`, `OUTPUT-005`, `OUTPUT-010`, journey 13 | Exact commands below, golden files, schema checks |
| T6 documentation and phase record | Primary / Sol reviewer | `DOC-010`–`013`, `DEV-008`, `DEV-012` | Append-only note, observed user docs, roadmap/changelog review |
| T7 platform and supply-chain evidence | Primary / independent platform reviewer | `TEST-012`–`013`, chapter 7 checks | Native smoke records, OSV/gitleaks/SBOM/grype/cosign evidence |
| T8 machine-interface clarification | Specification owner / independent reviewer | `OUTPUT-001`, `OUTPUT-010` | Accepted spec change or explicit authoritative interpretation |
| T9 independent implementation review | Independent Sol reviewer | `DEV-009`–`011` | Requirement-complete conformance report and gap closure |

### Exact validation plan

```text
go build ./cmd/ainfra
go fmt ./...
git ls-files -z '*.go' | xargs -0 go tool goimports -w
git diff --exit-code -- '*.go'
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go test ./internal/security -run '^$' -fuzz FuzzBuildEnvironment -fuzztime=10s
scripts/validate-v1-spec
scripts/validate-cli-contracts
scripts/build-targets.sh /tmp/ainfra-phase-1-targets
go tool govulncheck ./...
go tool gosec ./...
gitleaks git --redact
scripts/release-checks.sh run 0.1.0-alpha.1
```

The controlled offline test runs `go test ./...` with an empty inherited
environment plus explicit `PATH`, temporary `HOME`, `GOCACHE`, and
`GOMODCACHE`. Reviewed human-output evidence currently lives as literal expected
output in `internal/output/render_test.go`; T5 converts it to committed golden
files before completion.

OSV, shellcheck, hadolint, syft, grype, cosign, Docker build/scan, and native
macOS smoke are required T4/T7 evidence. If unavailable in the aibox
environment, they are explicit environment blockers rather than successful or
skipped checks.

### PR #49 impact and compatibility

The required field changes the not-yet-finalized draft machine-result contract.
No released compatibility promise is broken and no migration is needed. The
change affects chapter 7 build metadata, chapter 12 output, the schema, positive
and negative fixtures, semantic result types, both human renderers, JSON,
black-box tests, schema-validation scripts, phase documentation, and eventual
CLI reference material. It does not change roadmap phase order, changelog
release entries, current user documentation, release artifacts, security
boundaries, exit codes, or deferred lifecycle packages.

`AINFRA-DEV-008` reverification is T5: validate the positive version fixture,
reject every focused negative fixture, validate successful and failed generated
version envelopes, review human output, rerun unit/black-box tests, and rerun
all affected output requirements.

### Rollback and recovery

Each implementation wave is one focused commit after its gate. On failure,
leave the failing evidence visible, remove only generated ignored artifacts at
their explicit paths (`coverage.out`, the selected `/tmp/ainfra-*` directory,
or `dist/<exact-version>`), and restore the affected uncommitted patch to the
last accepted implementation commit without rewriting accepted specification
commits. Interface drift returns to T0/T8 for plan amendment. A failed
integration never advances `v1.x-dev`, promotion branches, roadmap status, or a
tag. Security or schema failure blocks the wave rather than being suppressed.

## Second plan correction — 2026-08-07

This section supersedes every earlier requirement-disposition matrix and task
graph. Earlier sections remain unchanged as audit history. Only the five
chapter-17 dispositions are used below.

### Remaining governing-source blockers

- The complete bodies of the nine decisions incorporated by
  `spec/doc/v1/README.md` are unavailable locally and were not found through
  the available GitHub organization search. Because the README says decisions
  prevail over summaries, plan acceptance is `blocked by unresolved decision`
  until immutable decision bodies are supplied or the specification owner
  authorizes the README summaries as the complete Phase 1 baseline.
- `help` and unknown-command results are commands under
  `AINFRA-OUTPUT-001`, but the closed result schema has no discriminator for
  them. `AINFRA-OUTPUT-001`, `AINFRA-CLI-002`, `AINFRA-CLI-003`, and
  `AINFRA-OUTPUT-010` therefore `require a specification change`. T8 freezes
  the affected CLI/output work.

### Complete Phase 1 disposition matrix

| Requirement | Disposition and Phase 1 evidence |
|---|---|
| `AINFRA-DEV-001`–`005`, `007`, `008` | covered by baseline, inventory, ownership, routing, deltas, and PR #49 impact analysis |
| `AINFRA-DEV-006` | blocked by unresolved decision until independent plan acceptance |
| `AINFRA-DEV-009`–`012` | deferred with explicit authority to the named post-integration, gap-closure, approval, and retrospective gates in this cycle |
| `AINFRA-CODE-001`, `002`, `010`, `011` | covered by T1/T5 formatting, lint, documentation, and comment-behavior checks |
| `AINFRA-CODE-003` | not applicable, with rationale: Phase 1 commits no generated Go source |
| `AINFRA-CLI-001`, `008`–`010` | deferred with explicit authority to roadmap Phase 2 deployment discovery |
| `AINFRA-CLI-002`, `003` | requires a specification change for help/unknown results; version work stays frozen in T2/T5 |
| `AINFRA-CLI-004`, `007` | not applicable, with rationale: Phase 1 exposes no prompts or confirmation |
| `AINFRA-CLI-005` | covered for the internal process boundary by T3; public use is deferred to the first engine phase |
| `AINFRA-CLI-006` | deferred with explicit authority to the first engine phase |
| `AINFRA-CONFIG-001`–`010` | deferred with explicit authority to roadmap Phase 2 |
| `AINFRA-OUTPUT-001`, `010` | requires a specification change through T8 |
| `AINFRA-OUTPUT-002`–`005`, `007`, `011`, `013` | covered by T2/T5 stream assertions, closed JSON, pre-render redaction, committed goldens, schema validation, and PR #49 arrays |
| `AINFRA-OUTPUT-006` | not applicable, with rationale: Phase 1 defines no verbosity or logging surface |
| `AINFRA-OUTPUT-008` | not applicable, with rationale: no CLI child command or raw mode exists |
| `AINFRA-OUTPUT-012` | not applicable, with rationale: Phase 1 emits no engine report |
| `AINFRA-SEC-010`, `012`–`014`, `025` | covered by T3/T5 structured argv, allowlisted environment, contained cwd, split streams, redaction, and mode tests |
| `AINFRA-SEC-011`, `015`, `024`, `026`, `027` | deferred with explicit authority to the first real-tool, credential, or engine-evidence phase |
| `AINFRA-TEST-010`–`018` | covered by T5/T7: required suites fail closed; four targets, native evidence, secret scans, fuzz retention, race, security, and format tests are required |
| `AINFRA-DOC-001`–`005`, `010`, `011` | covered by executable references, enumerated values, validated examples, and this append-only observable record |
| `AINFRA-DOC-012` | not applicable, with rationale: Phase 1 is not shipped |
| `AINFRA-DOC-013` | deferred with explicit authority to release review |
| `AINFRA-REL-001`, `003`, `009` | covered by the clean-worktree fail-closed release check, non-skipping checks, and recorded line/source/baseline |
| `AINFRA-REL-002`, `004`–`008` | deferred with explicit authority to release packaging and promotion; Phase 1 creates no release/tag/promotion |
| `AINFRA-IMPL-003` | deferred with explicit authority to first real-tool integration |
| `AINFRA-PROD-001`–`008` | deferred with explicit authority to their roadmap lifecycle, configuration, engine, evidence, and template phases; Phase 1 supplies only their Go foundation |
| `AINFRA-NONGOAL-001`–`009` | covered as absence constraints: T9 verifies no compiler, transformation, Go automation replacement, hosted plane, implicit install, published image, Windows support, Rust compatibility, or embedded engines |

Every exact later-phase requirement family in the prior inventory is deferred
only by its owning roadmap phase. T9 checks the implementation against every
file in `spec/doc/v1/`, not only the rows above.

### Corrected task graph and ownership

```text
T0 baseline/inventory -> independent plan review
T8 authoritative output clarification -> T2 CLI/output -> T5 validation
accepted review -> T1 module/build, T3 process/security, T4 Docker
T1 + T2 + T3 + T4 -> T5 -> T7 platform/supply-chain -> T6 phase record
T6 -> T9 independent implementation review -> gap closure -> owner approval
```

- T0 (primary): baseline `e5016cb`, every spec file, decision availability,
  standards, journeys, interfaces, owners, and rollback.
- T1 (Cora, Terra/low): `go.mod`, `go.sum`, `tools.go`, build scripts, Go
  packages, source comments, and four target binaries.
- T2 (ephemeral CLI engineer, Terra/medium): `cmd/ainfra`, `internal/app`,
  `internal/cli`, `internal/output`, and golden fixtures. Frozen behind T8.
- T3 (ephemeral security engineer, Terra/medium): `internal/security`,
  `internal/exec`, adversarial tests, environment/cwd/cancellation boundaries.
  Independent Sol/medium review supplies deep security synthesis.
- T4 (primary, Sol/medium review): root `Dockerfile`, `.dockerignore`, lint,
  local-only image build, SBOM and vulnerability evidence. No image publish.
- T5 (primary, Sol/medium review): `test/blackbox`, schema/example/golden
  validation, exact Phase 1 checks, and full-spec delta correction.
- T6 (primary): append-only note plus roadmap/changelog/user-doc review.
- T7 (ephemeral platform reviewer, Sol/medium): native platform and
  supply-chain evidence.
- T8 (specification owner, independent reviewer): accepted result-kind
  clarification for help and unknown invocation.
- T9 (independent Sol/medium reviewer): check code, tests, artifacts, docs, and
  behavior against every `spec/doc/v1/` file; correct every delta before
  approval.

Routing evidence: Agent A returned TeamMember `cora` and class `fast`, used as
Terra/low. Agents B and C returned no TeamMember or class and used ephemeral
Terra/medium. Independent reviewers returned no binding and use Sol/medium.

### Executable evidence and cleanup

```text
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go test ./internal/security -run '^$' -fuzz FuzzBuildEnvironment -fuzztime=10s
shellcheck scripts/validate-all scripts/test-all scripts/build-targets.sh scripts/validate-cli-contracts scripts/release-checks.sh
go tool govulncheck ./...
go tool gosec ./...
osv-scanner scan source -r .
gitleaks git --redact --report-path dist/phase1-evidence/gitleaks.json
hadolint Dockerfile
docker build --pull=false -t ainfra:phase1 .
syft ainfra:phase1 -o spdx-json=dist/phase1-evidence/sbom/container.spdx.json
grype ainfra:phase1 --fail-on high
scripts/validate-v1-spec
scripts/validate-cli-contracts
scripts/release-checks.sh run 0.1.0-alpha.1
```

Missing required executables fail T5/T7; they are never skipped. `go mod
download` first preloads the pinned graph. The offline test then uses a
temporary controlled home/cache and `GOPROXY=off`, retaining output at
`dist/phase1-evidence/offline/go-test.log`.

`scripts/build-targets.sh dist/phase1-evidence/platform` cross-builds all four
targets. Native smoke is required on Linux and both macOS architectures before
shipment. Each writes
`platform/<os>-<arch>/{version.json,sha256.txt,smoke.log}` and validates JSON
against the closed schema. Cross-builds do not replace native execution;
unavailable macOS runners block shipment.

Binary evidence uses `syft file:<binary>` and `grype <binary> --fail-on high`.
Blob signing is deferred to release packaging because no signing backend is
selected; `cosign sign-blob --yes --bundle <bundle> <binary>` becomes mandatory
after that governed selection.

Generated evidence is isolated below `dist/phase1-evidence/` and uncommitted.
After reporting, remove it and local image `ainfra:phase1`; `coverage.out` and
explicit `/tmp/ainfra-phase1-*` paths are likewise removable. The unrelated
owner-edited `aibox.toml` is excluded from staging, cleanup, rollback, and
clean-worktree claims. Release checks remain blocked while it is modified.

## Third plan correction — 2026-08-07

This section further supersedes conflicting statements in the second
correction while preserving them as audit history.

Additional dispositions are:

| Requirement | Disposition and authority |
|---|---|
| `AINFRA-TEST-001`–`003` | covered by permanent regression tests, black-box assertions for visible behavior, and Sol review of committed golden changes in T5/T9 |
| `AINFRA-TEST-014` | deferred with explicit authority: document/config fuzzing to Phase 2; source/archive/path fuzzing to Phase 3; standardized output, event, and redaction fuzzing to Phases 5–6 |
| `AINFRA-IMPL-001`, `002` | covered by T0/T9 contract-preservation review and the recorded standard-library-first dependency decision |
| `AINFRA-DOC-001`–`005` | deferred with explicit authority to roadmap Phase 8 template-authoring guidance; they are not claimed by the Phase 1 developer note |

For reconstructability, all other numbered requirement families are explicitly
deferred with roadmap authority as follows: `CONTRACT`, `LAYOUT`, `DOCTOR`,
`RECON`, and remaining `CLI`/`CONFIG` behavior to Phase 2; `SOURCE` and `LOCK`
to Phase 3; `INIT`, `PLAN`, `APPLY`, and plan-binding portions of `SEC` to
Phase 4; `OUTPUT`, `INV`, `ANS`, and engine `LOG` evidence to Phase 5;
`DESTROY` and recovery/hardening portions of `SEC` to Phase 6; `MCP` to Phase
7; `TPL` and authoring `DOC` to Phase 8; initial template/provider coverage to
Phases 9–11; `MIGRATE` to Phase 12; and later catalog/export/editor/bootstrap,
policy, coordination, remote-execution, and engine-evolution requirements to
their roadmap Phases 13–20. Phase 1 dispositions stated in the corrected
tables take precedence over these family deferrals.

T1 is now an ephemeral build/tooling engineer assignment at Terra/low because
`cora` is not resolvable in the project index. It owns only `go.mod`, `go.sum`,
the existing Go tool directives in `go.mod`, and `scripts/build-targets.sh`;
no `tools.go` is planned. T2 owns `internal/command` (not `internal/cli`),
`cmd/ainfra`, `internal/app`, and `internal/output`. T3 exclusively owns
`internal/security` and `internal/exec`. For B/C, routing returned exact fields
`recommended_team_member_slug: null` and `recommended_model_class: null`;
fallbacks are ephemeral `(CLI engineer, specialist)` and `(security engineer,
senior)`. Terra/medium is chosen for bounded implementation, with Sol/medium
independent review for deep synthesis. The retained router responses in the
session transcript are review evidence; their exact fields are repeated here.

The offline evidence command is executable as follows after the pinned module
graph is downloaded:

```sh
go mod download
mkdir -p /tmp/ainfra-phase1-offline/home /tmp/ainfra-phase1-offline/go-build
cp -a "$(go env GOMODCACHE)" /tmp/ainfra-phase1-offline/go-mod
env -i PATH="$(go env GOROOT)/bin:/usr/bin:/bin" \
  HOME=/tmp/ainfra-phase1-offline/home \
  GOCACHE=/tmp/ainfra-phase1-offline/go-build \
  GOMODCACHE=/tmp/ainfra-phase1-offline/go-mod GOPROXY=off \
  go test ./... 2>&1 | tee dist/phase1-evidence/offline/go-test.log
```

T7's platform owner runs these commands natively on each named runner
(`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`), substituting
only the literal `<os>-<arch>` directory name:

```sh
dist/phase1-evidence/platform/<os>-<arch>/ainfra version --format json \
  > dist/phase1-evidence/platform/<os>-<arch>/version.json
sha256sum dist/phase1-evidence/platform/<os>-<arch>/ainfra \
  > dist/phase1-evidence/platform/<os>-<arch>/sha256.txt
uv run --frozen --offline check-jsonschema \
  --schemafile spec/schemas/v1/machine-output.schema.json \
  dist/phase1-evidence/platform/<os>-<arch>/version.json \
  > dist/phase1-evidence/platform/<os>-<arch>/smoke.log 2>&1
```

On macOS, `shasum -a 256` replaces `sha256sum`. The platform owner signs the
three files with the runner identity in the implementation-review record.
Native runner absence blocks T7.

The four binary evidence commands use these explicit paths:

```text
dist/phase1-evidence/platform/linux-amd64/ainfra
dist/phase1-evidence/platform/linux-arm64/ainfra
dist/phase1-evidence/platform/darwin-amd64/ainfra
dist/phase1-evidence/platform/darwin-arm64/ainfra
```

For each path, T7 runs `syft file:<path> -o
spdx-json=dist/phase1-evidence/sbom/<os>-<arch>.spdx.json` and `grype <path>
--fail-on high`, recording output at
`dist/phase1-evidence/scans/<os>-<arch>.log`.

## Fourth plan correction — 2026-08-07

The family mapping is refined as follows. `LAYOUT-001`–`016` and
`CONTRACT-001`–`029` are deferred to Phases 2–4 according to discovery,
source-lock, and engine ownership; `LAYOUT-020`–`025` are deferred to Phases
4–6; `CONTRACT-030`–`038` to Phase 5. Operational `LOG-001`–`005` is deferred
to Phases 2 and 5, while retained-evidence `LOGS-001`–`007` is deferred to
Phases 4–6.

The remaining security IDs have exact authority: `SEC-001`–`006` to Phase 3;
the deferred portions of `SEC-011`, `015`, `020`–`022` to the first owning
contract/engine phase (2 or 4); `SEC-023`, `026`, `027`, and `040`–`042` to
Phase 5; and `SEC-024`, `030`–`034`, and `043` to Phase 6. The Phase 1
dispositions for `SEC-010`, `012`–`014`, and `025` still take precedence.

T1's routing fallback is explicitly ephemeral `(build/tooling engineer,
specialist)` at Terra/low.

The fail-closed offline command replaces the prior version:

```bash
set -o pipefail
mkdir -p dist/phase1-evidence/offline \
  /tmp/ainfra-phase1-offline/home \
  /tmp/ainfra-phase1-offline/go-build
go mod download
cp -a "$(go env GOMODCACHE)" /tmp/ainfra-phase1-offline/go-mod
env -i PATH="$(go env GOROOT)/bin:/usr/bin:/bin" \
  HOME=/tmp/ainfra-phase1-offline/home \
  GOCACHE=/tmp/ainfra-phase1-offline/go-build \
  GOMODCACHE=/tmp/ainfra-phase1-offline/go-mod GOPROXY=off \
  go test ./... 2>&1 | tee dist/phase1-evidence/offline/go-test.log
```

The primary/platform operator, not a sub-agent, owns T7 execution and all
artifact writes. The ephemeral Sol/medium platform reviewer is read-only.
Build output uses the script's actual hierarchy:

```text
dist/phase1-evidence/platform/linux/amd64/ainfra
dist/phase1-evidence/platform/linux/arm64/ainfra
dist/phase1-evidence/platform/darwin/amd64/ainfra
dist/phase1-evidence/platform/darwin/arm64/ainfra
```

Before redirection the operator creates each evidence directory with `mkdir
-p`. Binaries are built into those exact paths by
`scripts/build-targets.sh dist/phase1-evidence/platform`, transferred to the
matching native runner, made executable, and invoked there. The native command
is `<path>/ainfra --format json version`; its result is validated with
`uv run --frozen --offline check-jsonschema --schemafile
spec/schemas/v1/machine-output.schema.json <version.json>`, followed by
`sha256sum` on Linux or `shasum -a 256` on macOS. Output and exit status are
retained in the same `<os>/<arch>/smoke.log` directory. `syft file:<path>` and
`grype <path> --fail-on high` use each of the four literal paths above, with
SBOMs in `dist/phase1-evidence/sbom/<os>-<arch>.spdx.json` and scan logs in
`dist/phase1-evidence/scans/<os>-<arch>.log`.

## PR #51 resolution and rebaseline — 2026-08-07

Discussion #50 supplied the specification-owner resolution, and PR #51 was
squash-merged and integrated at baseline `cf8fa99`. The checked-in
specification is now explicitly the complete normative authority; external
decision IDs are provenance only. The governing-source blocker is closed.

PR #51 also resolves T8. Static help is a successful side-effect-free command
with discriminator `help`; pre-dispatch syntax and selection failures use
`invocation`, exit 2, and have a null result. Exact valid `--format json` or
`--format=json` before `--` selects the JSON error envelope even when other
syntax is invalid. Invalid format values fail in text mode. Post-dispatch
semantic failures retain the dispatched command discriminator.

The following requirements supersede their former blocked dispositions:

| Requirement | Disposition and evidence |
|---|---|
| `AINFRA-SPEC-001`, `002` | covered by baseline `cf8fa99` and T9 whole-spec review |
| `AINFRA-CLI-002`, `003`, `011`–`013` | covered by T2 static metadata/parser implementation and T5 unit/black-box tests |
| `AINFRA-OUTPUT-001`, `010`, `014`, `015` | covered by T2 closed semantic result types/rendering and T5 schema/golden tests |

T2 now includes root, group, and leaf static help metadata; normalization of
`help`, `--help`, and command `--help`; exact global format recognition before
`--`; and stable invocation diagnostics. T5 adds the PR #51 positive examples,
negative fixtures, schema checks, text/JSON stream separation, exit status,
side-effect assertions, and post-dispatch discriminator checks. No project
discovery, configuration, network, or child execution is added.

PR #51 affects `internal/command`, `internal/output`, their tests,
`test/blackbox`, `scripts/validate-cli-contracts`, the phase note, and the
eventual CLI reference. It does not alter process/security primitives, version
contract arrays, engine phases, roadmap status, release branches, or
`aibox.toml`. Affected work remains frozen until independent acceptance of this
rebaselined plan; after acceptance it is implemented and rechecked against
every file in `spec/doc/v1/`.

## Implementation and first conformance review — 2026-08-07

The independent plan review accepted baseline `cf8fa99`. Implementation added
the PR #49 supported-contract arrays and PR #51 static `help` and pre-dispatch
`invocation` results. The first T9 review rejected Phase 1 because help
metadata and delimiter handling were incomplete, mutable package state
remained, black-box/golden evidence was incomplete, Docker and supply-chain
delivery were absent, and user/phase documentation did not match behavior.

Gap closure made help metadata canonical across root, groups, and leaf topics;
made text render arguments; corrected `--`; removed mutable production
help/redaction globals; enforced envelope constructor invariants; expanded
parser, help, and schema tests; added a non-root scratch Dockerfile and narrow
build context; and updated installation and usage documentation.

Observed passing evidence after correction:

```text
go test ./...
go test -race ./...
go test -coverprofile=/tmp/ainfra-phase1-coverage.out ./...
go vet ./...
go tool staticcheck ./...
go tool golangci-lint run
go tool govulncheck ./...
go tool gosec ./...
scripts/validate-v1-spec
scripts/validate-cli-contracts
git diff --check
```

The host required isolated `GOCACHE`, `GOMODCACHE`, and `XDG_CACHE_HOME`
because its default cache locations are non-writable. Docker, hadolint, syft,
grype, and native macOS runners are unavailable here. Their mandatory T4/T7
execution evidence remains a Phase 1 completion blocker; it is not recorded as
a successful or skipped check. Roadmap Phase 1 remains `in_progress`.

## Second implementation review and focused gap closure — 2026-08-07

The second T9 review accepted PR #49 but rejected the focused PR #51 work.
Remaining defects were non-canonical leaf signatures, insufficient compiled
CLI coverage, absent reviewed help output, and documentation that claimed
canonical help before that was true.

The correction aligns every help usage, argument, and command option with the
chapter 4 command table, adds schema checks for every help topic, adds
compiled-CLI root/group/leaf and both-spelling coverage, exercises unknown
topics/options, malformed format, extra arguments, and `--`, bounds tested
processes, isolates state directories, and commits reviewed plan-help golden
output. Documentation now corresponds to the corrected metadata.

The wider T4/T7 environment blockers remain. This commit is an implementation
checkpoint, not Phase 1 shipment; it does not change the roadmap status.

## Third closing attempt — 2026-08-07

The closing review was repeated against specification baseline `772fba4` in
an isolated clean checkout. The complete local validation suite passed,
including formatting, `goimports`, vet, static analysis, curated linting,
unit and black-box tests, schema and example validation, CLI contract
validation, `govulncheck`, and `gosec`. Race and coverage runs passed. All four
Linux/macOS target binaries cross-built, the Dockerfile passed `hadolint`, and
the native Linux arm64 binary returned schema-conforming version output.

The retry also exposed one correctable release-tooling gap: `shellcheck`
reported SC1007 for the five repository scripts that establish their project
root. Those assignments now use explicit `CDPATH=''` syntax, and the complete
shell-script set passes `shellcheck`.

Phase 1 remains `in_progress` and no release or promotion was performed. The
following mandatory closing evidence is still unresolved:

- the full-history `gitleaks` finding for the synthetic API-key example was
  resolved with exact fingerprint entries in `.gitleaksignore`; the follow-up
  scan covered 131 commits and reported no leaks;
- this harness cannot execute the optional Dockerfile through Docker or
  rootless Podman, so container build, runtime smoke, SBOM, and image scan
  evidence is absent;
- native macOS amd64 and arm64 archive smoke evidence is absent;
- processkit doctor and release-audit each report two errors because the
  installed `changelog` and `git-branching` skills reference a missing
  `release-semver` skill;
- the primary worktree contains unrelated pre-existing changes and therefore
  cannot satisfy `AINFRA-REL-001`; and
- no intended Semantic Version release identity has been selected.

Per `AINFRA-DEV-009`–`011`, `AINFRA-REL-001`–`004`, and the fail-closed test
policy, these are blockers rather than skips. The release-evidence WorkItem
records the same state and remains blocked pending remediation and independent
requirement-complete conformance acceptance.
