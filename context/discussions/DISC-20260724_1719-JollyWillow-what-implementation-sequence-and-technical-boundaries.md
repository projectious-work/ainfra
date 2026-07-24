---
apiVersion: processkit.projectious.work/v2
kind: Discussion
metadata:
  id: DISC-20260724_1719-JollyWillow-what-implementation-sequence-and-technical-boundaries
  created: '2026-07-24T17:19:48+00:00'
  updated: '2026-07-24T19:58:38+00:00'
spec:
  question: 'What implementation sequence and technical boundaries should guide delivery
    of GitHub issue #1 for ainfra-templates?'
  state: resolved
  opened_at: '2026-07-24T17:19:48+00:00'
  outcomes:
  - DEC-20260724_1913-LucidCharm-use-python-3-12-and-uv
  - DEC-20260724_1913-PromptHaven-make-public-ipv4-allocation-opt-in
  - DEC-20260724_1913-KindLantern-deliver-kubernetes-ready-hosts-without-installing
  - DEC-20260724_1941-EarnestMeadow-require-capability-validated-remote-state-for
  - DEC-20260724_1951-ToughLark-use-debian-13-with-explicit-documented
  - DEC-20260724_1954-PolishedGlade-use-pinned-checkov-and-gitleaks-as
  - DEC-20260724_1954-PromptGlade-run-all-repository-automation-locally-and
  - DEC-20260724_1958-DaringSpark-make-explicitly-cost-approved-live-verification
  closed_at: '2026-07-24T19:58:38+00:00'
---

# Implementation plan for GitHub issue #1

Source: https://github.com/projectious-work/ainfra-templates/issues/1

## 1. Planning context

The repository is effectively a scaffold: outside processkit and harness
configuration it currently contains only `AGENTS.md`, `aibox.toml`, and
`aibox.lock`. Issue #1 is therefore greenfield work, not an incremental
refactor. It is the authoritative architecture and scope statement.

The implementation must preserve three boundaries:

1. OpenTofu owns infrastructure desired state.
2. Ansible owns host configuration.
3. The `ainfra` wrapper validates and visibly orchestrates those tools
   without becoming a deployment system or hiding their behavior.

The first release should target one concrete provider and topology:
Hetzner Cloud with a Kubernetes-ready baseline. Installing Kubernetes,
integrating aibox, installing processkit, and introducing a generic
multi-cloud abstraction are out of scope.

## 2. Recommended implementation choices

Use Python 3.12 with `uv` for milestone one. The wrapper is primarily
schema validation, subprocess orchestration, output sanitization, and safety
guards; Python minimizes bootstrap cost and supports fast, isolated tests.
Keep the package boundary clean enough that a later compiled implementation
can preserve the command and contract surface.

Use JSON Schema draft 2020-12 for all three public contracts. Keep
`ainfra-template.yaml` declarative and bounded. Do not add arbitrary
hooks or executable lifecycle fields.

Treat every safety invariant as both validation logic and a negative test.
Security requirements should fail before OpenTofu or Ansible mutation.

## 3. Delivery sequence

### Milestone 0: repository foundation and executable specification

Create the top-level structure from the issue:

- `README.md`, `LICENSE`, and security-focused `.gitignore`
- `docs/architecture.md`, `docs/security-model.md`,
  `docs/state-and-secrets.md`, `docs/authoring-templates.md`, and
  `docs/operator-guide.md`
- `schemas/template-manifest.v1alpha1.json`
- `schemas/template-input.v1alpha1.json`
- `schemas/template-output.v1alpha1.json`
- Python package and CLI under `tools/ainfra/`
- `scripts/validate-all` and `scripts/test-all`
- non-secret positive and negative fixtures

Define exit-code classes before implementing commands: invalid usage,
contract validation, missing/incompatible dependency, unsafe configuration,
underlying tool failure, and approval refusal. Document direct
`tofu`/`ansible-playbook` equivalents.

Completion gate:

- schemas compile under draft 2020-12;
- sample manifest, input, and output validate;
- unsupported versions and unknown fields fail;
- local test scripts run from a clean checkout;
- no GitHub Actions are added.

### Milestone 1: contracts and pure validation core

Implement typed internal models around the schemas without duplicating
schema rules inconsistently. Add:

- template discovery restricted to direct descendants of `templates/`;
- path normalization and traversal/symlink escape rejection;
- strict `apiVersion`, `kind`, wrapper-version, capability, and engine
  checks;
- input loading from JSON or YAML;
- secret-reference validation with no secret resolution during structural
  validation;
- environment classification for disposable versus non-disposable targets;
- security-policy validation for state, SSH, CIDRs, pins, and public ingress;
- output-contract validation and recursive secret-key/value detection.

Reject unknown manifest fields, broad management networks, inline secret
values, generated-key configuration, root/password SSH, unsafe production
state, and unpinned dependencies.

Completion gate:

- `ainfra validate` is deterministic and side-effect free;
- every issue-listed negative case has a named test;
- validation errors include the file, field path, invariant, and remediation;
- fixtures and logs contain placeholders only.

### Milestone 2: thin CLI and execution adapters

Implement the six commands with one shared execution pipeline:

1. discover and load the template;
2. validate manifest and input;
3. verify required executables and minimum versions;
4. validate state/backend and security prerequisites;
5. resolve secret references only into child-process environments;
6. invoke the underlying tool with streamed stdout/stderr;
7. preserve non-zero exits;
8. validate and emit only the standardized sanitized output.

Command behavior:

- `validate`: boundary checks only, with optional tool/static checks.
- `plan`: `tofu init` without hidden migration, `tofu plan` to an
  explicit plan file, plus Ansible syntax/check mode where meaningful.
- `apply`: require `--approve`; apply the reviewed plan rather than
  silently replanning; run Ansible explicitly after infrastructure success.
- `destroy`: require `--approve` and an exact confirmation containing
  template, environment, and target identity; enforce ownership scope.
- `outputs`: transform declared OpenTofu outputs into
  `InfrastructureOutput/v1alpha1`, redact, validate, then render JSON/YAML.
- `doctor`: report wrapper, OpenTofu, Ansible, scanner, schema, backend,
  lock, and dependency health without mutation.

Put subprocess construction behind adapters and test argv/environment/log
redaction. Never place secret values in arguments. Preserve raw tool output
unless a redaction rule applies, and make redaction visible.

Completion gate:

- all commands have unit and subprocess-contract tests;
- approval cannot be bypassed by non-interactive execution;
- signals and non-zero exits propagate correctly;
- direct-tool equivalents are documented and match wrapper behavior;
- no automatic repair or state migration occurs.

### Milestone 3: Hetzner OpenTofu baseline

Add `templates/hetzner-kubernetes-baseline/` with pinned OpenTofu and
Hetzner provider constraints and a committed provider lock. Separate files
by responsibility: versions/providers, variables, locals/invariants,
network, servers, firewalls, outputs, and backend example.

Implement:

- one control-plane-capable node and configurable workers;
- a configurable RFC1918 private network no broader than `/16`;
- explicit public IPv4/IPv6 flags, disabled where feasible by default;
- role-specific, deny-by-default firewalls;
- operator-supplied SSH public keys only;
- pinned OS image by immutable identifier or documented stable selector;
- cloud-init that creates a non-root administrator and disables root and
  password authentication;
- no private-key, password, token, or kubeconfig resources in state;
- remote encrypted and locked state requirement for non-disposable inputs;
- non-secret outputs sufficient to generate inventory and the standard
  hand-off contract.

Do not call the result a Kubernetes cluster. Use
`target.type: kubernetes-ready` until a tested Kubernetes distribution is
installed in separately scoped work.

Completion gate:

- `tofu fmt -check -recursive`, `tofu validate`, and `tofu test` pass;
- plan assertions prove no public SSH default, no broad CIDR, no generated
  key, and role-correct firewall rules;
- ownership tags/labels allow destroy-scope verification;
- state and recovery documentation covers backup, locking, rotation, and
  disposable local-state exceptions.

### Milestone 4: Ansible configuration and inventory hand-off

Generate inventory from an explicit, non-secret OpenTofu output mapping.
Keep transport metadata separate from credentials. Pin collections and
roles in `requirements.yml`.

Implement idempotent roles for:

- base packages, locale, time synchronization, and hostname;
- non-root administrator and SSH hardening;
- automatic security updates and explicit reboot policy;
- host firewall consistent with cloud firewall intent;
- logging/audit and observability prerequisites;
- conservative kernel/network hardening appropriate to the target.

Add syntax, lint, check-mode, and idempotence tests. Treat inaccessible
private-only nodes as an operator configuration error with actionable tunnel
or private-network guidance; do not silently open public SSH.

Completion gate:

- generated inventory contains no secrets;
- `ansible-lint`, syntax check, and check mode pass;
- a second Ansible run reports no unexpected changes;
- recovery access is documented without a permanent insecure path;
- standardized output references credentials rather than embedding them.

### Milestone 5: integrated security gates and disposable verification

Wire `scripts/validate-all` to run, in a stable documented order:

1. schema and fixture validation;
2. wrapper formatting, lint, type, unit, and integration tests;
3. OpenTofu formatting, validation, tests, and lock checks;
4. cloud-init schema validation;
5. Ansible lint, syntax, check mode, and idempotence harness;
6. IaC scanning;
7. secret scanning;
8. documentation/link/example checks.

Use pinned scanner/tool versions or clearly declared minimum versions.
Ensure missing tools fail with installation guidance rather than silently
skipping required gates.

In a disposable Hetzner project, capture evidence for reviewed plan, apply,
standard output validation, secret scans of plan/logs/inventory/output,
second-run idempotence, recovery procedure, and ownership-scoped destroy.
Do not commit state, secrets, raw credentials, or sensitive evidence.

Completion gate:

- every acceptance criterion in issue #1 maps to an automated check or a
  documented manual verification with retained non-secret evidence;
- destroy removes only resources owned by the selected target;
- limitations and deferred features are explicit;
- the repository is ready for review without claiming aibox integration.

## 4. Suggested pull-request decomposition

1. Foundation, schemas, fixtures, documentation skeleton, and test harness.
2. Validation core and `validate`/`doctor`.
3. Execution adapters and guarded `plan`/`apply`/`destroy`/`outputs`.
4. Hetzner network, compute, firewall, state, and cloud-init baseline.
5. Ansible inventory, hardening roles, and operator runbooks.
6. Security gates, negative tests, and disposable-environment evidence.

Each pull request should be independently testable, keep contracts stable,
and update the acceptance matrix. Avoid mixing infrastructure behavior with
large wrapper refactors.

## 5. Test architecture

Use three layers:

- pure tests for schemas, policy rules, redaction, output transformation,
  discovery, version checks, and approval parsing;
- adapter tests with fake executables that assert argv, environment,
  streaming, signals, exit codes, and failure ordering;
- tool-backed tests for OpenTofu, cloud-init, Ansible, scanners, and the
  disposable Hetzner environment.

Maintain a machine-readable acceptance matrix linking issue criteria to test
IDs and manual evidence. Negative tests are first-class: each unsafe
configuration should fail before any fake mutation marker is reached.

## 6. Risks and controls

- Contract drift with future aibox consumption: version schemas, keep output
  minimal, and coordinate breaking output changes outside this issue.
- Wrapper growth: expose direct commands, centralize orchestration, and reject
  template-specific branches in the CLI.
- Secret leakage through subprocesses: environment-only resolution,
  structured redaction, sanitized outputs, and adversarial tests.
- False safety from static validation: assert generated plans and perform a
  disposable apply/idempotence/destroy review.
- Provider and OS drift: pin dependencies and image selection; make upgrades
  explicit reviewed changes.
- Private-access bootstrap complexity: document supported operator network or
  tunnel setup and fail closed instead of enabling public SSH.
- CI ambiguity: keep all gates locally executable; do not add hosted
  workflows without a separate decision.

## 7. Open questions for review

1. Confirm Python/uv for v0.1 versus accepting the extra bootstrap cost of
   Rust now.
2. Choose the exact remote-state backend pattern to document and validate for
   non-disposable environments.
3. Decide whether public IPv4 should be entirely disabled by default or
   allowed for workload traffic while management ingress remains closed.
4. Select the pinned Debian release/image identifier and upgrade policy.
5. Select the local IaC and secret scanners and their pinning mechanism.
6. Define the disposable Hetzner test budget, credentials owner, and evidence
   retention location.
7. Confirm whether the initial baseline stops at Kubernetes-ready hosts, as
   recommended, or separately scopes a minimal distribution after the
   baseline is proven.

## 8. Definition of done

Issue #1 is complete only when the six commands exist, the contracts are
versioned and strictly validated, the Hetzner template and Ansible roles pass
all local gates, unsafe defaults are demonstrably rejected, the standardized
output is secret-free, a disposable lifecycle proves idempotence and scoped
destroy, and documentation makes project boundaries and direct-tool
equivalents unambiguous.

This discussion remains active until the open questions are reviewed. Once
the implementation choices are accepted, record the consequential choices
as DecisionRecords, link them as outcomes, and resolve this Discussion.
