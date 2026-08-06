## Command surface

| Command | Purpose |
|---|---|
| `ainfra init` | Create a minimal deployment definition without overwriting existing files or creating infrastructure. |
| `ainfra doctor [--reconcile]` | Shorthand for `ainfra doctor all` against the nearest deployment. |
| `ainfra doctor all [TARGET] [--reconcile]` | Diagnose the applicable environment, deployment, resolved template, and latest-run contracts together. |
| `ainfra doctor deployment [TARGET] [--reconcile]` | Diagnose a selected deployment definition, native inputs, template lock, and referenced paths. |
| `ainfra doctor template [TARGET] [--reconcile]` | Diagnose a template layout, contract compatibility, documentation, fixtures, and available local child-tool checks. |
| `ainfra doctor run [TARGET] [--reconcile]` | Diagnose run evidence, saved-plan bindings, interruption state, and safe resumability. |
| `ainfra doctor environment [--reconcile]` | Diagnose operating-system, architecture, executable, version, and capability prerequisites required by ainfra. |
| `ainfra template lock` | Resolve the selected template source and record its immutable revision and content digest. |
| `ainfra template update` | Resolve an explicitly requested newer template revision and update the lock after compatibility checks. |
| `ainfra template migrate SOURCE --to VERSION [--write]` | Analyze a template-contract migration and optionally apply safe deterministic changes to a local working copy. |
| `ainfra plan [--destroy]` | Create a saved apply or destroy plan bound to the deployment, opaque native inputs, template, and tool versions. |
| `ainfra apply --plan RUN_ID` | Verify all bindings and apply the exact reviewed OpenTofu apply plan. |
| `ainfra configure --run RUN_ID [--check]` | Run Ansible for an applied run, optionally in check mode, using generated inventory and native variables. |
| `ainfra deploy --plan RUN_ID` | Apply a reviewed plan, collect output, generate inventory, configure hosts, and verify convergence. |
| `ainfra output --run RUN_ID` | Show the sanitized standardized infrastructure output recorded for a run. |
| `ainfra inventory --run RUN_ID` | Show or regenerate deterministic Ansible inventory from a run's validated standardized output. |
| `ainfra status` | Summarize deployment and run state, including failures, cancellations, and recovery guidance. |
| `ainfra destroy --plan RUN_ID` | Apply the exact reviewed OpenTofu destroy plan and record the engine result. |
| `ainfra version` | Print the ainfra version and machine-readable build information. |

`deploy` composes apply, output collection, inventory generation, Ansible
configuration, and Ansible check-mode verification. It MUST NOT create a plan
or infer approval.

## Common behavior

- **AINFRA-CLI-001:** commands MUST locate the nearest deployment by searching
  ancestors for `ainfra.yaml`, unless `--project` selects an exact root.
- **AINFRA-CLI-002:** every command MUST support `--format text|json` where its
  result is meaningful to automation.
- **AINFRA-CLI-003:** JSON output MUST use a versioned envelope and stderr MUST
  contain diagnostics only.
- **AINFRA-CLI-004:** `--non-interactive` MUST prevent prompts and fail when
  required confirmation or input is absent.
- **AINFRA-CLI-007:** `--yes` MAY satisfy a documented confirmation only when
  `--non-interactive` is also set; it MUST NOT bypass security refusals, plan
  bindings, or destructive approval requirements.
- **AINFRA-CLI-005:** cancellation signals MUST be forwarded to the active
  child process and recorded without claiming rollback.
- **AINFRA-CLI-006:** ainfra MUST show the equivalent direct engine command in
  verbose or diagnostic output, with secrets redacted.

## Initialization

`init` creates a minimal deployment without overwriting existing files. It MAY
select a source and create examples copied from the resolved template.

- **AINFRA-INIT-001:** conflicts MUST be detected before the first write.
- **AINFRA-INIT-002:** `.gitignore` updates MUST be idempotent and append only a
  clearly marked ainfra block.
- **AINFRA-INIT-003:** init MUST NOT create credentials, private keys, backend
  resources, or infrastructure.

## Doctor

`doctor` is the single command for validation, troubleshooting, and
compatibility. With no subcommand it behaves as `doctor all` against the nearest
deployment. `doctor all [TARGET]` explicitly composes applicable environment,
deployment, resolved-template, and latest-run checks. The other subcommands
narrow the target when authoring a template, inspecting a run, or testing a
development environment:

- `environment`: operating system, architecture, executable discovery,
  supported versions, and capabilities required by ainfra;
- `deployment`: manifest/schema version, native variable files, permissions,
  ignored sensitive paths, native input pointers, and referenced file
  containment;
- `template`: layout and manifest, version compatibility, dependency pins,
  documentation, OpenTofu and Ansible syntax, standardized outputs, fixtures,
  and deprecated contracts;
- `run`: event-log integrity, saved-plan bindings, executable drift, interrupted
  stages, output/inventory provenance, and safe resumability advice;
- source-lock completeness, immutable revision, digest, and cache containment
  are checked as part of the deployment or template that owns the reference.

Doctor's knowledge stops at ainfra's contracts. It MUST NOT contain
provider-specific, resource-topology, network, DNS, bastion, tunnel, host, or
application-health logic. It MUST NOT test provider credentials or backend and
target connectivity. Those concerns belong to the template and to OpenTofu or
Ansible.

Doctor MAY invoke documented, non-mutating child-tool checks such as
`tofu fmt -check`, `tofu validate`, and Ansible syntax checking. It passes
native files through unchanged and reports the command, exit status, and
sanitized diagnostics. It MUST NOT reproduce OpenTofu or Ansible validation,
interpret provider/resource semantics, or convert localized console output
into a stronger conclusion than the child tool provides. A check that requires
plugins, credentials, backend access, inventory connectivity, or provider
operations is skipped with an explanation and a direct command the operator
can run deliberately.

Each check has a stable ID and emits exactly one of `pass`, `skip`, `warning`,
or `fail`, plus concise evidence, remediation, and whether a safe automatic
reconciler exists. `--format json` returns all findings even when failures
exist, allowing editors, CI, and support tooling to consume the same contract.
Text output SHOULD group failures first and include a reproducible rerun
command. Doctor MUST distinguish an unavailable check from a passed check.

### Reconciliation

`--reconcile` asks doctor to correct findings with registered reconcilers. It
does not broaden doctor's ownership boundary. A reconciler MUST be
deterministic, idempotent, separately testable, and limited to ainfra-managed
local derived or operational artifacts. Suitable fixes include creating
ainfra runtime directories with safe permissions, tightening permissions on
ainfra-owned operational files, removing contained stale temporary files, and
regenerating inventory from already validated standardized output.

Reconciliation MUST NOT:

- edit `ainfra.yaml`, native tfvars, Ansible variables, backend configuration,
  template source, or engine-native lock files;
- install or upgrade executables, providers, modules, roles, or collections;
- run mutating OpenTofu, Ansible, Git, SSH, provider, backend, host, or network
  operations;
- alter plans, state, immutable run events, credentials, firewall rules, or
  remote infrastructure; or
- guess a fix for an ambiguous finding.

Before writing, doctor builds an ordered reconciliation plan containing check
IDs, paths, intended changes, and rollback limitations. Interactive execution
requires confirmation. Non-interactive execution requires both
`--non-interactive` and `--yes`. Doctor rechecks preconditions immediately
before each fix, records every attempted change, reruns affected checks, and
reports `applied`, `failed`, or `still_failing`. It MUST NOT claim that the
whole reconciliation was atomic when only individual file writes were atomic.

`ainfra template migrate SOURCE --to VERSION` identifies the current ainfra
template contract version, reports incompatible or deprecated constructs, and
produces an ordered migration plan. `--write` MAY apply documented
deterministic transformations to a local working copy only. It MUST produce a
patch, create a backup or require a clean version-controlled worktree, never
migrate remote or cached content in place, never rewrite native deployment
variable values, and rerun `ainfra doctor template` afterward. Provider and
resource migration remain explicit template-specific operations. Backend and
infrastructure-state migration remain OpenTofu-owned operations. None of them
may be performed by this command.

- **AINFRA-DOCTOR-008:** doctor MUST not mutate provider infrastructure.
- **AINFRA-DOCTOR-009:** missing optional prerequisites MUST be distinguished
  from blockers for the selected command.
- **AINFRA-DOCTOR-010:** diagnostics MUST contain a stable code, severity,
  affected path or component, explanation, and next action.
- **AINFRA-DOCTOR-001:** every doctor command MUST be read-only unless the user
  explicitly supplies `--reconcile` and confirms the reconciliation plan.
- **AINFRA-DOCTOR-002:** every check MUST declare its scope, child-process
  needs, prerequisites, and applicability rule.
- **AINFRA-DOCTOR-003:** skipped and unavailable checks MUST include a reason
  and MUST NOT be counted as passes.
- **AINFRA-DOCTOR-004:** diagnostic evidence MUST pass through the same
  redaction and output-scanning policy as child-process output.
- **AINFRA-DOCTOR-005:** doctor SHOULD suggest a direct child-tool command or
  documentation URL; it MUST NOT offer automatic credential, state, backend,
  firewall, connectivity, or remote infrastructure repair.
- **AINFRA-DOCTOR-006:** every doctor JSON finding MUST populate `check`,
  `scope`, `status`, `code`, `severity`, and `message`; failed or skipped checks
  MUST also populate `nextAction`.
- **AINFRA-DOCTOR-007:** adding a certified template MUST NOT require adding
  provider, topology, connectivity, or application-specific checks to ainfra.
- **AINFRA-DOCTOR-011:** bare `doctor` and `doctor all` MUST select and execute
  the same checks for the same deployment target.
- **AINFRA-RECON-001:** a check without a registered safe reconciler MUST
  remain a manual finding when `--reconcile` is set.
- **AINFRA-RECON-002:** reconciliation MUST acquire the same deployment-local
  operation lock used to prevent conflicting ainfra writes.
- **AINFRA-RECON-003:** a failed reconciliation MUST preserve its evidence and
  continue only with fixes whose preconditions remain valid.
- **AINFRA-RECON-004:** reconcilers MUST have fixtures proving idempotency,
  containment, safe interruption, and redaction.
- **AINFRA-MIGRATE-001:** migrations MUST be explicit source-version to target-
  version transformations and MUST be idempotent where automated.
- **AINFRA-MIGRATE-002:** migration analysis and resulting patches MUST be
  available in machine-readable output.
- **AINFRA-MIGRATE-003:** unsupported or ambiguous transformations MUST stop
  with a manual action, never a guessed rewrite.

## Planning

1. Load and validate deployment and lock.
2. Materialize the locked template into a new run workspace.
3. Hash deployment metadata, all declared native input files, template, and
   relevant executable versions.
4. Run `tofu init`, adding each declared backend configuration file as a
   repeated `-backend-config=<path>` argument in declaration order.
5. Run `tofu plan -out=plan.tfplan`, adding each declared variable file as a
   repeated `-var-file=<path>` argument in declaration order; add `-destroy`
   for a destroy plan.
6. Create a sanitized structural plan summary and immutable plan record.

- **AINFRA-PLAN-001:** plan IDs MUST be unguessable or content-derived with
  collision resistance and safe as directory names.
- **AINFRA-PLAN-002:** saved binary plans and raw plan JSON MUST be treated as
  sensitive.
- **AINFRA-PLAN-003:** plan output MUST clearly state apply versus destroy,
  deployment, template source/digest, native input digests, and next command.
- **AINFRA-PLAN-004:** a destroy plan MUST never authorize apply, and an apply
  plan MUST never authorize destroy.

## Apply

- **AINFRA-APPLY-001:** apply MUST require an exact reviewed plan/run ID.
- **AINFRA-APPLY-002:** immediately before child execution, ainfra MUST verify
  all plan bindings and plan bytes.
- **AINFRA-APPLY-003:** apply MUST execute `tofu apply <saved-plan>` rather than
  generate a new implicit plan.
- **AINFRA-APPLY-004:** ainfra MUST record `started` before process execution
  and a terminal `succeeded`, `failed`, or `cancelled` event afterward.
- **AINFRA-APPLY-005:** ambiguous interruption MUST route to manual inspection.

## Output and inventory

After successful apply, ainfra runs `tofu output -json`, extracts only the
manifest-declared output, validates it, scans it for secret-shaped material,
and persists standardized `output.json`. Inventory is derived from that output
without reading provider state directly.

- **AINFRA-INV-001:** inventory generation MUST be a pure deterministic
  function of validated output.
- **AINFRA-INV-002:** output and inventory MUST be written atomically.
- **AINFRA-INV-003:** generated inventory MUST not contain deployment secrets.

## Configure and verify

`configure` invokes Ansible Runner with the materialized template project,
generated inventory, and every declared native Ansible variable file in
declaration order. `--check` enables check/diff behavior. `deploy` MUST run a
normal configuration followed by a separate check-mode verification.

- **AINFRA-ANS-001:** SSH host-key checking MUST be enabled.
- **AINFRA-ANS-002:** an independently populated `known_hosts` file MUST be
  required for SSH deployments unless the connection plugin provides an
  equivalent authenticated trust mechanism.
- **AINFRA-ANS-003:** successful verification requires zero unreachable and
  failed hosts and zero changes for every expected inventory host.
- **AINFRA-ANS-004:** Ansible Runner event data, not localized console prose,
  SHOULD be used for outcome verification.
- **AINFRA-ANS-005:** successful prior stages MAY be resumed; interrupted or
  failed mutating stages MUST NOT be automatically repeated.

## Destroy

Destroy uses the same reviewed-plan boundary as apply.

- **AINFRA-DESTROY-001:** destroy MUST require a saved destroy plan ID.
- **AINFRA-DESTROY-002:** destroy MUST verify deployment, template, opaque
  native-input, plan, and operation bindings before execution.
- **AINFRA-DESTROY-003:** ainfra MUST report the result of `tofu apply
  <destroy-plan>` without independently reading or interpreting OpenTofu state.
- **AINFRA-DESTROY-004:** provider-side independent verification MUST be
  documented for every certified template.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Successful requested operation |
| 1 | Operation or child engine failed |
| 2 | Invalid command or input contract |
| 3 | Missing/incompatible dependency |
| 4 | Security or trust refusal |
| 5 | Stale binding or reviewed-plan mismatch |
| 6 | Ambiguous interrupted state requiring recovery |

Specific diagnostics use stable `AINFRA-E####` codes inside the machine
envelope; callers MUST rely on the code, not English text.
