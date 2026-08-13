# Phase 2: Contracts and doctor

## Planning entry — 2026-08-12

Phase 2 implements the local contracts-and-doctor boundary defined by the
canonical v1 specification under `spec/doc/v1/`. It starts from the shipped
Phase 1 Go foundation on `v1.x-dev` and is implemented on
`feat/phase-02-contracts-doctor`.

Phase 1 shipped as `v1.0.0-alpha.1`. Phase 2 is now `in_progress`.

### Outcome and boundary

At the Phase 2 completion gate, a user can point `ainfra doctor` at a local
deployment or already-resolved template and receive deterministic, actionable
text or JSON findings. The user may explicitly approve repairs only for
contained, ainfra-owned local artifacts.

Phase 2 does not acquire template sources, contact providers, mutate native
engine inputs, run infrastructure lifecycle operations, or implement MCP mode.
Source locking and materialization belong to Phase 3; OpenTofu and Ansible
lifecycle execution belongs to Phases 4 through 6.

### Requirement and disposition matrix

The disposition values are:

- **implement**: required for the Phase 2 completion gate;
- **supporting**: a cross-cutting contract exercised by Phase 2;
- **partial**: Phase 2 implements the local validation portion while later
  lifecycle phases own execution; and
- **deferred**: explicitly outside the Phase 2 boundary.

| Normative requirement | Disposition | Phase 2 implementation and evidence |
|---|---|---|
| `AINFRA-CLI-001`, `008`–`010` | implement | Deterministic explicit path, `--project`, `AINFRA_PROJECT`, and nearest-ancestor discovery; canonical directory/file equivalence; ambiguity, missing-manifest, and no-descendant-search black-box tests. |
| `AINFRA-CLI-002`–`004`, `007`, `011`–`013` | supporting | Shared text/JSON selection, stream separation, confirmation rules, side-effect-free static help, and the exact pre-dispatch invocation envelope and exit code. |
| `AINFRA-INIT-001`–`003` | implement | Conflict-first initialization, idempotent marked `.gitignore` block, and no credentials, keys, backends, or infrastructure; repeated-init and no-partial-write tests. |
| `AINFRA-CONTRACT-001`–`003` | implement | Strict ainfra-owned document parsing, unknown-field and inline-secret rejection, and deployment-schema validation. |
| `AINFRA-CONTRACT-010`–`015` | implement | Provider-neutral deployment/template validation, compatibility checks, contained regular paths, native dependency ownership, OpenTofu requirement, and Ansible applicability. |
| `AINFRA-CONTRACT-020`–`029` | partial | Preserve ordered native pointers and opaque bytes without translation or provider interpretation. Child-engine argv execution and plan bindings remain lifecycle-owned in Phases 4–6. |
| `AINFRA-CONTRACT-030`–`038` | partial | Validate declared standardized-output and inventory contracts locally. Output collection and inventory execution remain in Phases 4 and 5. |
| `AINFRA-DOCTOR-001`–`011` | implement | Registry of checks with stable scope, prerequisites, capabilities, applicability, status, code, severity, explanation, and next action; deterministic complete reports for `environment`, `deployment`, `template`, `run`, and `all`. |
| `AINFRA-RECON-001`–`004` | implement | Plan-before-write local reconciliation under a deployment lock, explicit confirmation, per-fix precondition checks, retained failure evidence, post-checks, containment, idempotency, and redaction. |
| `AINFRA-MIGRATE-001`–`003` | deferred | Template migration is outside Phase 2. Any shared analysis types must remain deterministic and machine-readable without exposing a migration command. |
| `AINFRA-CONFIG-001`–`010` | implement | Immutable typed effective configuration, exact precedence, project-layer restrictions, unknown/secret/version rejection, and redacted sorted provenance through doctor environment. |
| `AINFRA-OUTPUT-001`–`008`, `010`–`015` | supporting | Shared semantic results, one versioned JSON object, plain non-TTY output, central redaction, schema-valid fixtures, stable renderer goldens, and engine attribution where child checks apply. |
| `AINFRA-LOG-001`–`005` | partial | Phase 2 uses pre-redacted structured events and preserves correlation. Durable operational sinks and full concurrent lifecycle logging mature in Phase 6. |
| `AINFRA-SEC-010`–`015` | implement | Validate regular executables and versions; use argv without a shell, allowlisted environment, contained working directories, sanitized streams, and no secrets in arguments. |
| `AINFRA-SEC-020`–`027` | supporting | Reject inline secrets, classify/redact sensitive values including chunk boundaries, secret-scan examples and evidence, and enforce owner-only operational permissions. |
| `AINFRA-SEC-030`–`032` | supporting | Keep OpenTofu state opaque and engine-owned; retain native input bytes and ordering as future plan bindings. |
| `AINFRA-SEC-001`–`006`, `040`–`043` | deferred | Immutable acquisition/cache and SSH lifecycle controls belong to Phase 3 or later. Doctor must not implement provider, host, key-generation, or bastion behavior. |
| `AINFRA-TEST-001`–`003`, `010`–`018` | supporting | Offline deterministic unit, component, black-box, adversarial, race, fuzz, schema, secret, Linux, and macOS evidence with no unexpected skips. |
| `AINFRA-DOC-010`–`013` | implement | Maintain this append-only phase note with boundaries, decisions, evidence, deviations, security/compatibility effects, documentation changes, gaps, and linked follow-up work. |
| `AINFRA-DOC-001`–`005` | partial | Document user-visible discovery, configuration, diagnostics, output, exit, and security behavior as it ships. Full template-authoring documentation belongs to Phase 8. |

### Prioritized implementation work

The processkit epic is
`BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor`.
Its ordered children are:

| Order | Priority | WorkItem | Deliverable |
|---:|---|---|---|
| 1 | high | `BACK-20260812_0749-DeepLeaf-establish-config-project-discovery` | Effective configuration and deterministic project discovery. |
| 2 | high | `BACK-20260812_0749-BraveCedar-implement-deployment-native-contract-loader` | Strict deployment and opaque native-file loader. |
| 3 | high | `BACK-20260812_0749-CuriousBloom-validate-local-template-contracts` | Validation of already-local resolved templates. |
| 4 | high | `BACK-20260812_0749-ActiveShore-build-deterministic-doctor-check-registry` | Capability-aware, deterministic doctor checks and reports. |
| 5 | high | `BACK-20260812_0749-TallBeacon-expose-doctor-cli-stable-diagnostics` | Doctor CLI, stable text/JSON, streams, and exits. |
| 6 | medium | `BACK-20260812_0749-TrustyCharm-add-safe-local-doctor-reconciliation` | Narrow local repair plans and guarded execution. |
| 7 | medium | `BACK-20260812_0749-DaringTrail-add-phase-two-adversarial-coverage` | Negative, component, black-box, race, and fuzz coverage. |
| 8 | low | `BACK-20260812_0749-KindSpruce-complete-phase-two-conformance-docs` | Final conformance matrix, documentation, and release gate. |

### Implementation sequence

Implementation proceeds in dependency order: discovery and configuration,
contract loaders, local template validation, doctor registry, CLI rendering,
safe reconciliation, adversarial coverage, and conformance closure. Each slice
must be independently testable and may not introduce lifecycle placeholders.

### Parallel planning review

Three read-only agents reviewed the normative requirements, current Go
architecture, and implementation breakdown using the cost-efficient
`gpt-5.6-luna` model at low effort. Their independent conclusions agreed on
the boundary and dependency order captured above. All repository mutations
remain with the primary agent as required by repository policy.

### Evidence log

#### 2026-08-12 — Phase transition and planning

- Phase 1 marked `shipped` and linked to its developer note.
- Phase 2 marked `in_progress` and this developer note registered in the
  normative roadmap.
- The Phase 2 requirement/disposition matrix was derived from every applicable
  normative chapter, including contracts, CLI/lifecycle, security, output,
  configuration/logging, testing, quality, and documentation.
- One epic and eight ordered implementation WorkItems were created. The epic
  and first discovery/configuration slice entered `in-progress`.
- Decision `DEC-20260812_0746-PluckyRaven-ship-phase-1-and-begin-contracts`
  records the accepted phase transition and implementation boundary.

### Open evidence and gaps

Phase 2 local contract, doctor, initialization, and reconciliation behavior is
implemented and validated. Later-phase dependencies remain explicit: source
acquisition and locks are Phase 3; saved plans and detailed retained-run
semantics are Phase 4; standardized output/inventory execution is Phase 5; and
full operational logging is Phase 6. Their Phase 2 portions do not claim those
future execution outcomes.

#### 2026-08-12 — Discovery and deployment-loader foundation

- Added `internal/project` with canonical explicit target resolution and
  nearest-ancestor discovery. Explicit directory and manifest-file forms have
  one identity, and explicit directories never trigger descendant searches.
- Implemented the normative target precedence as immutable inputs: positional
  explicit path, `--project`, `AINFRA_PROJECT`, then nearest ancestor.
- Added a strict YAML deployment loader with unknown-field, API version, kind,
  name, required template source, duplicate path, traversal, symlink, special
  file, and containment enforcement.
- Native input paths retain declaration order. Their contents are deliberately
  not read or interpreted by the loader.
- Added unit coverage for equivalence, precedence, nearest-ancestor selection,
  no descendant search, unknown fields, traversal, duplicate inputs, symlinks,
  and ordered opaque native inputs.
- `scripts/validate-all` passed, including formatting, imports, vet,
  staticcheck, golangci-lint, unit tests, schema/spec validation, container-gate
  tests, govulncheck, and gosec.
- `scripts/test-all` passed, including unit, race, coverage, and container-gate
  suites. The new `internal/project` package reports 77.2% statement coverage
  in this initial slice.

This slice establishes internal contracts only. Wiring the resolver into CLI
options, effective configuration provenance, and doctor results remains open
under the first and subsequent WorkItems.

#### 2026-08-12 — Effective-configuration core

- Added `internal/config` with immutable typed UI, logging, storage-path, and
  executable settings plus sparse typed merge layers.
- Implemented the normative precedence mechanics for compiled defaults,
  system, user, project, explicit, supported environment, and command-line
  layers. Scalars replace lower values and destination lists replace rather
  than append.
- Added deterministic, unique, key-sorted effective-value provenance with the
  winning source and ordered overridden sources.
- Added the exact Linux and macOS system/user paths, optional project path,
  and required explicit-file behavior. No generic current-directory or
  `.ainfra/` configuration discovery exists.
- Configuration YAML uses strict known-field decoding and rejects unsupported
  API versions, kinds, and enum values. Environment booleans accept only
  `true` or `false`; environment path overrides must be absolute.
- Repository-controlled project configuration is limited to `ui` and
  `logging.level`. Prohibited executable, storage, destination, and syslog
  selections fail normal resolution. Diagnostic resolution records them and
  computes the safe result without applying them.
- Added precedence, provenance, path-resolution, project-hostility,
  unknown-field, environment-validation, optional-layer, and Linux/macOS
  location tests.

The remaining part of the first WorkItem is CLI composition: selecting
`--config`, `--project`, and their environment counterparts without violating
static-help side-effect rules, then exposing the redacted result through
`doctor environment`. Log-file, log-format, and local-syslog environment
overrides will be completed with that composition because their semantics
operate on the effective destination list rather than an isolated scalar.

#### 2026-08-12 — CLI composition and environment doctor

- Added the first executable Phase 2 command, `ainfra doctor environment`.
  Its JSON discriminator is `doctor.environment`; its result contains scope,
  summary, findings, and the complete effective-configuration view required by
  the machine-result schema.
- Added command parsing for `--config` and `--project`, including `--option`
  and `--option=value` forms, with typed application requests. Explicit
  `--project` conflicts with a different `AINFRA_PROJECT` instead of silently
  winning.
- Configuration and project discovery occur only after a canonical command is
  dispatched. Static root, group, and leaf help remain independent of host
  configuration and discovery.
- The application composition loads the project layer only when a deployment
  is explicitly selected or found as the nearest ancestor. Commands without
  an applicable deployment have no project layer.
- Added closed environment handling for log format, file, and local-syslog
  destinations. File paths must be absolute; environment destination order is
  deterministic; rotation defaults are 10 MiB, five backups, seven days, and
  compression enabled.
- Extended semantic diagnostics with doctor check, scope, status, and
  reconciliation fields while preserving the existing invocation contract.
- Added deterministic plain-text rendering and one-object JSON rendering for
  doctor results.
- Extended the CLI schema-validation gate to execute and validate a real
  `doctor environment` result. Unit and compiled black-box tests cover command
  dispatch, static-help isolation, project restrictions, conflicting
  selectors, clean streams, and the complete 12-key configuration result.

This first doctor slice deliberately performs only local configuration and
supported-platform checks. Executable discovery and version checks will enter
through the capability-aware doctor registry; no child executable is invoked
by this command yet.

#### 2026-08-12 — Doctor registry and executable capabilities

- Added `internal/doctor` with validated check definitions, stable IDs,
  explicit scopes, declared prerequisites and child-tool needs, applicability,
  and injected capabilities. Checks receive no implicit filesystem, process,
  environment, or network authority.
- Registry construction rejects incomplete and duplicate checks. Execution
  does not stop after failures and sorts complete findings failures-first,
  followed by warnings, skips, and passes with stable check/path/code order.
- Added environment checks for the supported OS/architecture matrix and the
  `tofu`, `ansible-runner`, `git`, and `ssh` executables. OpenTofu is a required
  v1 dependency; an unavailable OpenTofu check fails. Tools whose applicability
  depends on later deployment/template facts are reported as skipped, never
  passed, when unavailable.
- Host executable discovery validates absolute, regular, executable files and
  fingerprints them through the Phase 1 security boundary. Version queries run
  with structured argv, an empty allowlisted environment, a contained working
  directory, redacted streams, and a five-second timeout through the sole
  subprocess runner.
- Successful executable findings record canonical path and reported version.
  Auto-discovered paths appear in effective configuration with `discovered`
  provenance; explicitly configured paths retain their actual winning layer.
- Doctor reports with failed checks now use a partial-failure envelope: the
  complete schema-valid result and failed diagnostics are retained while the
  command returns exit code 3 for a missing or incompatible dependency.
- Added registry tests for deterministic ordering, pass/skip/fail summaries,
  duplicate refusal, non-short-circuit behavior, remediation on skips, and
  discovered executable provenance. CLI schema and black-box tests accept both
  healthy exit 0 and missing-dependency exit 3 while always validating the
  complete machine result.

#### 2026-08-12 — Deployment doctor

- Added `ainfra doctor deployment [TARGET]` with canonical
  `doctor.deployment` text and JSON results. Omitted targets use the normative
  project flag, environment, or nearest-ancestor resolution path.
- A positional target conflicts with `--project` or `AINFRA_PROJECT`; the CLI
  refuses ambiguity before loading a manifest.
- Deployment diagnosis is built only from the strict loader's validated facts:
  one canonical identity/root, a structurally valid ainfra-owned manifest, and
  ordered native pointers that resolve to contained non-symlink regular files.
  Native bytes remain unread and uninterpreted.
- Added stable deployment checks for identity, manifest, and native inputs.
  The registry produces three deterministic pass findings after successful
  loading; contract loading failures retain the `doctor.deployment`
  discriminator rather than falling back to an invocation error.
- Missing and malformed input contracts return exit code 2 with
  `AINFRA-E2300`. Typed security refusals such as traversal and symlink paths
  return exit code 4 with `AINFRA-E2304`.
- Deployment-bound configuration uses the same resolver and precedence as the
  environment doctor, including the deployment's project layer. The deployment
  result does not duplicate the environment-only effective-configuration
  payload.
- Added unit, registry, CLI-dispatch, compiled black-box, and machine-schema
  validation for valid deployments, target conflicts, missing targets, and
  unsafe native pointers.

#### 2026-08-12 — Local template doctor

- Added `ainfra doctor template [TARGET]` with canonical `doctor.template`
  text and JSON results. The target is an already-resolved local template
  directory; omitted target means the current directory. `--project` is
  rejected because this command does not resolve a deployment source.
- Added `internal/template`, a strict known-field loader for
  `ainfra-template.yaml`. It validates API version, kind, normalized name and
  directory-basename identity, semantic version syntax, mandatory OpenTofu,
  and closed inventory/Ansible applicability.
- Declared engine directories, playbooks, documentation, native dependency
  files, and the minimal example must be contained non-symlink paths of the
  required type. Absolute and traversing declarations are refused.
- Ansible templates require native `requirements.yml`, the declared playbook,
  and a minimal native Ansible variables file through the deployment example.
  Infrastructure-only templates must omit Ansible when inventory is `none`.
- The minimal example is loaded through the same strict deployment loader, so
  its ainfra-owned contract and native input pointers receive identical
  validation without interpreting native bytes.
- Template README and variables documentation are mandatory. The README must
  expose equivalent direct OpenTofu commands and, when applicable, direct
  Ansible commands.
- Template source containing `.ainfra`, `.terraform`, `*.tfstate`, or
  `*.tfstate.*` runtime/infrastructure state is refused before checks pass.
- Added stable checks for template identity, layout, native engine ownership,
  and inventory applicability, plus loader, application, registry, command,
  compiled black-box, and machine-schema evidence.
- Missing or malformed template contracts return exit 2 with `AINFRA-E2400`;
  containment, symlink, file-type, and state-policy refusals return exit 4 with
  `AINFRA-E2405`.

No source URL is parsed, cloned, updated, locked, or cached in this slice.
Immutable acquisition remains Phase 3.

#### 2026-08-12 — Retained-run and aggregate doctor

- Added `ainfra doctor run [TARGET]`. It locates the latest retained local run
  deterministically and validates the required `run.json` and `events.jsonl`
  regular files through a deployment-rooted filesystem handle.
- A deployment with no retained runs reports an explained `skip`, never a
  pass. An incomplete latest run reports a failed finding while preserving the
  complete machine result for troubleshooting.
- Added `ainfra doctor all [TARGET]` as the aggregate of the environment,
  deployment, resolved-template, and latest-run scopes. Findings are sorted by
  the same stable registry ordering and receive a recomputed aggregate summary.
- Until Phase 3 has resolved and locked a deployment template, the aggregate
  emits an explicit template-scope skip with remediation. It does not acquire
  source content or falsely claim that an unresolved template passed.
- Bare `ainfra doctor [TARGET]` dispatches the exact same application function
  and canonical `doctor.all` result as `ainfra doctor all [TARGET]`. A command
  test compares their complete JSON bytes for the same target.
- Extended the compiled CLI contract gate to validate schema-conforming
  `doctor.run` and `doctor.all` envelopes, including the allowed partial
  dependency-failure result from aggregate environment checks.

This retained-run slice establishes local layout and availability evidence.
Saved-plan bindings, executable drift, interruption semantics, output and
inventory provenance, and deeper event integrity depend on the Phase 4 run
lifecycle artifacts and remain explicitly deferred rather than inferred.

#### 2026-08-12 — Guarded local runtime reconciliation

- Added `internal/reconcile` with an ordered, reviewable action model and a
  single registered repair for the ainfra-owned deployment runtime directory.
  The repair creates `.ainfra` with mode `0700` or restricts an existing
  non-symlink directory to that mode; it cannot edit manifests or native files.
- Planning is read-only and records the check ID, exact path, intended mode,
  action kind, and rollback limitation. Unsafe file types and symlinks are
  security refusals rather than repair candidates.
- Applying a plan takes an exclusive advisory deployment-operation lock on
  the existing `ainfra.yaml`, recomputes the complete plan under that lock,
  and refuses stale preconditions before writing.
- Doctor deployment and doctor all expose the runtime-permission finding as
  `warning` with reconciliation `available`. After successful application they
  rerun the check and report `pass` with reconciliation `applied`.
- Interactive reconciliation prints the full ordered plan and requires a
  terminal confirmation. Automation must provide all three of `--reconcile`,
  `--non-interactive`, and `--yes`; `--yes` cannot be used on its own.
- Tests cover plan purity, creation, restrictive permissions, idempotency,
  stale-plan refusal, symlink refusal, missing automation approval, approved
  two-phase execution, post-application checks, and schema-valid doctor output.
- Failed individual repairs now remain in the complete doctor report as
  reconciliation `still_failing`, with retained structured evidence and the
  still-applicable repair plan. The executor continues to later actions whose
  preconditions remain valid; lock or stale-plan failures still stop all writes.
- Retained-run diagnostics require the selected run directory and its mandatory
  `run.json` and `events.jsonl` evidence to be owner-only. Group- or
  world-accessible operational evidence fails closed with a direct remediation.

This slice intentionally registers no repair for executable installation,
configuration, deployment manifests, native inputs, templates, retained run
evidence, state, credentials, or remote infrastructure.

#### 2026-08-12 — Adversarial coverage expansion

- Added native Go fuzz targets for the strict deployment and template YAML
  decoders, seeded with valid contracts, malformed bytes, multiple-document
  structures, unknown fields, and recursive-alias-shaped input.
- Executed each new fuzz target for two seconds with eight workers. The
  deployment decoder completed 136,925 executions and the template decoder
  completed 127,360 executions without a crash or invariant violation.
- Extended the compiled-binary black-box suite to exercise a real approved
  reconciliation. It verifies that JSON stdout remains one clean protocol
  object, plan evidence is isolated to stderr, `.ainfra` is created with mode
  `0700`, and the resulting finding reports reconciliation `applied`.
- Corrected the deployment black-box expectation to include the new runtime
  permission finding and forced an uncached execution while developing the
  fixture, avoiding false confidence from the external-binary test cache.

#### 2026-08-12 — Minimal deployment initialization

- Added `ainfra init [DEPLOYMENT]` with canonical text and JSON results. An
  omitted target selects the current directory; the directory basename must be
  a valid deployment name.
- Initialization checks the target type, existing `ainfra.yaml`, and malformed
  ainfra ownership markers in `.gitignore` before its first write. Existing
  manifests are never overwritten.
- The generated deployment contains only the v1 deployment identity and a
  placeholder local template reference. It creates no credentials, keys,
  backend configuration, engine state, or infrastructure.
- `.gitignore` receives only the clearly marked ainfra block for `.ainfra/`.
  Existing unrelated content is preserved and complete existing ainfra blocks
  are not duplicated.
- Added package, command-dispatch, compiled CLI schema, no-overwrite, and
  conflict-before-write coverage for the initializer.

#### 2026-08-13 — Session-start conformance verification

- Verified Hugo `v0.164.0` is available and ran the complete documentation,
  v1 specification, CLI schema, static-analysis, security, and container-gate
  validation through `scripts/validate-all`.
- Ran `scripts/test-all`, including ordinary, race, coverage, compiled
  black-box, and container-gate suites. All configured tests passed.
- Corrected the compiled black-box harness to resolve Go's effective module
  cache when `GOMODCACHE` is not explicitly exported, preserving its offline
  nested build without depending on an empty cache path.
- Replaced ambiguous shell `A && B || C` guards in container and release
  scripts with explicit conditionals. This retains the fail-closed path and
  symlink checks while satisfying the configured ShellCheck gate.
- Rechecked the Phase 2 disposition matrix against the retained evidence. No
  new Phase 2 implementation gap remains; later-phase acquisition, lifecycle,
  output collection, and durable logging boundaries remain explicitly
  deferred as recorded above.

#### 2026-08-13 — Independent completion review

The post-integration review used merged commit `3c9d8df` as its immutable
baseline and inspected implementation, tests, schemas, examples, user
documentation, phase evidence, and release material against every row of the
Phase 2 requirement/disposition matrix.

| Requirement group | Completion result | Evidence disposition |
|---|---|---|
| CLI, initialization, configuration | satisfied | Unit, compiled black-box, schema, and static-help checks pass. |
| Deployment and template contracts | satisfied | Strict loaders, containment tests, negative fixtures, and fuzz targets pass. |
| Doctor scopes and output | satisfied | All five scopes retain deterministic complete text and JSON results. |
| Local reconciliation | satisfied | Plan, lock, approval, precondition, failure-retention, and post-check evidence pass. |
| Security and testing | satisfied | Offline, race, fuzz, vulnerability, secret, static-security, and container gates pass. |
| Documentation and phase record | satisfied | User docs and this append-only record describe the observed Phase 2 boundary. |
| Later lifecycle behavior | accepted deferral | Phase 3 owns acquisition and locks; Phases 4–6 own execution, outputs, and durable logging. |

The review found no unexplained skip, unverifiable Phase 2 claim,
specification inconsistency, or unowned gap. The conformance matrix is accepted
for release preparation. Decision
`DEC-20260813_0853-CoolFern-ship-phase-2-only-after-verified` keeps the roadmap
state at `in_progress` until the release is published and independently
verified.

#### 2026-08-13 — Completion approval and release identity

Phase 2 is approved as shipped after the ordered completion gates succeeded:

- the independent requirement-complete review found no unexplained Phase 2
  gap;
- `scripts/release-checks.sh run 1.0.0-alpha.2` passed the full local,
  security, race, coverage, archive, and cross-platform validation sweep;
- the owner-operated macOS host gate passed native smoke, offline Docker
  build, hardened runtime, Syft SBOM, Grype vulnerability, and cleanup checks;
- the checksum manifest was signed and verified with Sigstore identity
  `info@projectious.work` and issuer `https://github.com/login/oauth`; and
- the publish workflow downloaded the public assets and independently
  reverified their checksums and signature.

The release identity is:

- release: `v1.0.0-alpha.2`;
- release commit: `34eab1a617d1c8bd133fda2e475c4fbce560727c`;
- GitHub release ID: `369842902`;
- published at: `2026-08-13T10:49:09Z`;
- public release:
  <https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.2>;
- published payload: four platform archives, four SPDX JSON SBOMs,
  `checksums.sha256`, and `checksums.sha256.sigstore.json`.

GitHub's public release API independently reported the release as non-draft,
prerelease, and fully uploaded. This evidence satisfies
`AINFRA-DEV-009`–`012` and Decision
`DEC-20260813_0853-CoolFern-ship-phase-2-only-after-verified`. The roadmap may
therefore transition Phase 2 from `in_progress` to `shipped`.
