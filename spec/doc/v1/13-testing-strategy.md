## Objectives

The test strategy provides fast local feedback, realistic process-boundary
coverage, permanent regression protection, and explicit release evidence.
Tests prove observable contracts and safety properties; they MUST NOT merely
mirror implementation structure.

- Put behavior at the lowest test level that proves it reliably.
- Add a black-box test when behavior is visible through the CLI contract.
- Keep default tests deterministic, parallel-safe, credential-free, and
  offline.
- Test failure and interruption paths as deliberately as successful paths.
- Use real OpenTofu and Ansible only where fakes cannot prove integration.
- Never make live provider tests an incidental developer command.

## Test layers

| Layer | Scope | Dependencies | Purpose |
|---|---|---|---|
| Unit | One package or pure function | None | Prove parsing, transformations, state transitions, and policy edge cases quickly. |
| Component | Packages behind one boundary | Temporary files, fake time and tools | Prove collaboration, persistence, reconciliation, cancellation, and redaction. |
| Black box | Compiled `ainfra` binary | Fake tools and isolated environment | Prove commands, output, exit codes, configuration, and filesystem effects. |
| Real-tool integration | CLI adapters | Pinned local OpenTofu and Ansible Runner | Prove supported child-tool contracts without cloud infrastructure. |
| Disposable end to end | Complete certified template | Explicit credentials and cost approval | Prove the complete create, configure, verify, and destroy lifecycle. |

Each higher layer is smaller than the layer below it. End-to-end tests do not
replace focused tests: they are slower, harder to diagnose, and cannot cover
all negative paths economically.

## Unit tests

Unit tests live beside packages as `*_test.go`. External test packages are
preferred for a package contract; same-package tests are reserved for important
unexported invariants that cannot otherwise be exercised.

Direct unit coverage is required for:

- YAML/JSON parsing, unknown fields, canonical serialization, and versions;
- source parsing, path containment, digesting, and lock comparison;
- configuration precedence, layer restrictions, and value provenance;
- diagnostic applicability, ordering, and reconciliation planning;
- plan bindings, run transitions, interruption, and recovery decisions;
- output validation and deterministic inventory conversion;
- redaction across chunks and secret-shaped output detection; and
- rich, plain, JSON, and structured-log event construction before IO.
- engine evidence framing, protocol-version negotiation, declarative profile
  validation, unknown-event behavior, and error-only classification;
- every command-discriminated machine-result shape, including partial failure
  and interrupted recovery results.

Table-driven tests SHOULD cover equivalent valid and invalid cases. Time, IDs,
environment, filesystem-sensitive decisions, and randomness use injected
values. Tests MUST NOT depend on the developer's home, locale, terminal, Git
configuration, SSH agent, or ambient environment.

## Component and process-contract tests

Component tests use temporary directories and fake `tofu`, `ansible-runner`,
`git`, and `ssh` executables. Fakes record argv, cwd, selected environment
names, stdin, signals, output chunks, and exit behavior. Fixtures simulate slow
output, split secrets, malformed events, non-zero exits, hangs, and
cancellation.

These tests prove that ainfra:

- never invokes a shell or changes argument boundaries;
- forwards only allowed environment variables;
- preserves native input bytes and expected paths;
- records sanitized correlated events under concurrent output;
- refuses stale bindings before invoking a mutating child;
- handles partial writes and interruption without false success; and
- applies only reconcilers whose preconditions still hold.

Fake executables are fixtures, not alternate production adapters.

## Black-box CLI tests

Black-box tests build the real executable once and invoke it from outside
`internal/` packages. They use an empty controlled environment, isolated
config/cache/run directories, fake tools, fixed terminal capabilities, and
bounded timeouts.

The suite covers:

- help, version, invalid syntax, and every documented exit code;
- project discovery and complete config/env/flag precedence permutations;
- explicit deployment paths, nested paths, missing manifests, and fail-closed
  parent directories containing one or more deployment children;
- `-v`, `-vv`, `-vvv`, log formats, rotation, and a fake local syslog endpoint;
- rich, plain, no-color, narrow-width, non-TTY, and JSON output;
- default combined, source-filtered, `--errors`, and guarded `--raw` run-log
  views, including ainfra lifecycle events, refusal to search operational
  sinks, and refusal to guess from unstructured prose;
- doctor scopes, `doctor all`, reconciliation confirmation, partial failure,
  and post-reconciliation checks;
- lock, plan, apply, configure, deploy, status, and destroy behavior using fake
  tools; and
- signal forwarding, cancellation, corrupt evidence, and recovery guidance.

Tests assert stdout, stderr, sinks, exit code, child invocation, and allowed
filesystem changes separately. JSON assertions decode the schema. Human output
uses reviewed golden files with volatile values normalized before comparison.

## Real-tool integration tests

Real-tool tests cover every minimum supported OpenTofu and Ansible Runner minor
version plus the current preferred version. They use local fixtures, disabled
OpenTofu backends where supported, Ansible localhost/check mode, and temporary
homes and caches.

They verify argv, initialization assumptions, OpenTofu JSON UI framing and
version records, native variables, output JSON, Ansible Runner events,
check-mode interpretation, and unsupported protocol versions. A
missing selected tool produces an explicit skip locally; release gates treat a
required-version skip as failure.

## Disposable end-to-end tests

Every certified template has an opt-in test that performs:

1. doctor and lock verification;
2. reviewed plan creation and approval capture;
3. exact-plan apply;
4. standardized output and inventory generation;
5. Ansible configuration and zero-change verification;
6. repeat-plan/idempotency inspection;
7. reviewed destroy-plan execution and engine-result recording; and
8. independent provider-side teardown verification.

Live tests require an explicit enable flag, isolated credentials, budget and
region limits, unique ownership labels, a maximum lifetime, and an emergency
cleanup procedure. Evidence is sanitized. A zero destroy exit code alone MUST
NOT be reported as independently verified cleanup.

## Regression policy

Every confirmed defect receives a test that fails for the original behavior
before or with the fix. The regression belongs at the lowest effective layer;
a user-visible defect also receives a black-box assertion. Security defects add
a negative fixture proving refusal or redaction.

Regression tests reference the issue or advisory in test or fixture metadata.
A regression or golden case may be removed only when behavior is intentionally
retired, with migration documentation and review rationale. Golden and schema
updates require a readable diff; unreviewed bulk regeneration is prohibited.

- **AINFRA-TEST-001:** every defect fix MUST add or identify permanent
  regression coverage.
- **AINFRA-TEST-002:** user-visible regressions MUST be asserted through the
  compiled CLI unless that cannot reproduce the contract.
- **AINFRA-TEST-003:** changed golden files MUST be reviewed as product output,
  not accepted as mechanical data.

## Fuzzing and adversarial testing

Fuzz targets cover manifests, locks, source syntax, paths and archives,
configuration merging, structured output, inventory, Ansible events, and
redaction chunk boundaries. Seed corpora include past parser/security
regressions and compact valid examples.

Adversarial fixtures cover traversal, symlink races, special files, oversized
documents, malformed encodings, argument injection, poisoned caches, reordered
events, split secrets, log rotation races, reconciliation races, and corrupt
plans or run records. Minimized failures are committed only after secret scans.

## Concurrency and race testing

`go test -race ./...` is mandatory on Linux for release candidates and changes
to streaming, logging, cancellation, cache, run storage, or parallel checks.
Stress tests exercise concurrent child streams, rotation, cancellation, and
operation-lock contention.

Tests use deadlines and observable synchronization, not arbitrary sleeps. A
flaky test is a defect: fix it or quarantine it with an owner, issue, scope, and
expiry. Silently retrying until green is prohibited.

## Fixtures and isolation

Shared fixtures live under top-level `testdata/`; package fixtures use adjacent
`testdata/`. Harnesses live under `test/blackbox/`, `test/integration/`, and
`test/e2e/`.

Fixtures use reserved domains, synthetic addresses, placeholder credentials,
and minimum content. Every test owns a temporary root. No test writes to real
user configuration, caches, run directories, known-hosts files, or system logs.

## Coverage and performance

Coverage highlights risk; it is not a release percentage. Critical parsers,
security policies, bindings, reconciliation, run state, inventory, and process
execution require direct positive and negative tests.

Benchmarks cover large inventories, event streams, redaction, hashing, and
template materialization. Thresholds follow a representative baseline.
Correctness, containment, and redaction MUST NOT be weakened for a benchmark.

## Developer and release gates

| Gate | Required suites |
|---|---|
| Local fast | Formatting, vet/static analysis, unit, component, and black-box tests. |
| Change-risk | Local fast plus race, fuzz smoke, and affected real-tool tests. |
| Release candidate | Supported unit/black-box targets, race, security, compatibility matrix, integration, and schema/golden validation. |
| Certified template | Release candidate plus the template's approved disposable end-to-end lifecycle. |

The build chapter defines native commands and tools. Release reports record
versions, suites, skips, failures, reruns, live environment, and sanitized
evidence.

- **AINFRA-TEST-010:** default tests MUST run without network, cloud
  credentials, OpenTofu, or Ansible.
- **AINFRA-TEST-011:** required release checks MUST fail on an unexpected skip.
- **AINFRA-TEST-012:** Linux and macOS black-box suites MUST cover supported
  architectures natively or through a documented build and smoke combination.
- **AINFRA-TEST-013:** test output and fixtures MUST pass the production
  artifact secret-scanning policy.
