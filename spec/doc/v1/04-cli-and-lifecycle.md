# 4. CLI and lifecycle

## Command surface

```text
ainfra init
ainfra validate
ainfra doctor
ainfra template lock
ainfra template update
ainfra plan [--destroy]
ainfra apply --plan RUN_ID
ainfra configure --run RUN_ID [--check]
ainfra deploy --plan RUN_ID
ainfra output --run RUN_ID
ainfra inventory --run RUN_ID
ainfra status
ainfra destroy --plan RUN_ID
ainfra version
```

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

## Validation and doctor

`validate` performs offline structure, containment, lock, manifest, and native
file checks, then invokes `tofu fmt -check` and `tofu validate` in an isolated
workspace where possible. `doctor` checks executable versions, Git/SSH
availability, credentials by presence only, backend readiness where safe, and
filesystem posture.

- **AINFRA-VALIDATE-001:** validation MUST not mutate provider infrastructure.
- **AINFRA-VALIDATE-002:** missing optional prerequisites MUST be distinguished
  from blockers for the selected command.
- **AINFRA-VALIDATE-003:** diagnostics MUST contain a stable code, severity,
  affected path or component, explanation, and next action.

## Planning

1. Load and validate deployment and lock.
2. Materialize the locked template into a new run workspace.
3. Hash deployment metadata, native input files, backend config, template, and
   relevant executable versions.
4. Run `tofu init` with an explicit backend configuration.
5. Run `tofu plan -out=plan.tfplan -var-file=<native tfvars>`; add `-destroy`
   for a destroy plan.
6. Create a sanitized structural plan summary and immutable plan record.

- **AINFRA-PLAN-001:** plan IDs MUST be unguessable or content-derived with
  collision resistance and safe as directory names.
- **AINFRA-PLAN-002:** saved binary plans and raw plan JSON MUST be treated as
  sensitive.
- **AINFRA-PLAN-003:** plan output MUST clearly state apply versus destroy,
  deployment, template source/digest, state backend identity where non-secret,
  and next command.
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
generated inventory, and native deployment `ansible-vars.yaml`. `--check`
enables check/diff behavior. `deploy` MUST run a normal configuration followed
by a separate check-mode verification.

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
- **AINFRA-DESTROY-002:** destroy MUST verify deployment, template, input,
  backend, plan, and operation bindings before execution.
- **AINFRA-DESTROY-003:** successful `tofu apply <destroy-plan>` MUST be
  followed by an empty-state verification for the same backend before
  reporting the deployment destroyed.
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
