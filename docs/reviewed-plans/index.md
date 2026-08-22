# Reviewed OpenTofu Plans


ainfra separates OpenTofu planning from infrastructure mutation. `plan`
creates a private saved plan and immutable review record. `apply` requires the
exact run ID and executes only those saved bytes.

## Prerequisites

The deployment must have a valid `ainfra.yaml`, a current `ainfra.lock`, and a
verified template cache entry. OpenTofu is discovered only when `plan` or
`apply` runs. Pin an explicit executable in trusted user or explicit
configuration when required:

```yaml
apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
executables:
  tofu: /usr/local/bin/tofu
```

Repository-controlled configuration cannot select executable, cache, or run
paths.

## Create and review a plan

```sh
ainfra plan path/to/deployment
```

Planning creates a collision-resistant run ID and an owner-only directory
under the configured runs root. It snapshots declared backend and variable
files, materializes the locked template, runs `tofu init`, creates
`plan.tfplan`, and reduces `tofu show -json` to structural action counts.
Raw plan values are not persisted in the summary.

For automation, request the stable v1 result:

```sh
ainfra plan path/to/deployment --format json
```

The result identifies apply versus destroy intent, deployment and template
digests, the aggregate input digest, saved-plan digest, structural engine
report, evidence paths, and exact next command.

Destroy planning is explicit:

```sh
ainfra plan path/to/deployment --destroy
```

It produces a destroy-intent record and a `ainfra destroy ... --plan RUN_ID`
next command. `ainfra apply` always refuses that record. After reviewing the
structural delete counts, execute only that saved plan:

```sh
ainfra destroy path/to/deployment --plan RUN_ID
```

ainfra repeats every binding check described below and invokes `tofu apply
<saved-destroy-plan>` as an argument array. It never runs `tofu destroy`,
creates an implicit replacement plan, or reads OpenTofu state.

## Apply the exact reviewed plan

```sh
ainfra apply path/to/deployment --plan RUN_ID
```

ainfra acquires the deployment operation lock and then reverifies every plan
binding: deployment manifest, ordered native inputs, template lock and cache,
template-controlled workspace bytes, OpenTofu executable and version, summary,
and saved plan. Any drift returns the stale-binding exit code without starting
OpenTofu.

After recording durable `started` evidence, ainfra invokes only:

```text
tofu apply -input=false -no-color <contained-relative-plan.tfplan>
```

There is no implicit replanning. A second invocation with the same run ID is
refused, including when another apply process is already using it.

## Evidence and recovery

`events.jsonl` records the started event before execution and a succeeded,
failed, or cancelled terminal event afterward. Cancellation or an ambiguous
process interruption additionally records `inspection-required` and disables
automatic retry.

Destroy uses the same evidence and interruption boundary. A zero OpenTofu exit
code records the engine result; it is not independent proof that the provider
contains no owned resources. Follow the certified template's provider-side
teardown procedure after every destroy.

A certified provider template's teardown procedure must identify:

- the authenticated, read-only provider API or inventory command;
- the deployment ownership labels, account, project, and region boundaries;
- the paginated query and the exact empty-result condition;
- a bounded wait for provider eventual consistency;
- the sanitized evidence retained for certification; and
- an emergency cleanup and escalation path when resources remain.

The verification must query the provider directly. It must not infer absence
from the destroy exit code, a saved plan, ainfra events, or direct inspection
of OpenTofu state.

When inspection is required, do not create or apply another plan merely to
clear the error. Preserve the run directory, inspect it with:

```sh
ainfra doctor run path/to/deployment
```

Then inspect the backend and infrastructure using the provider's native,
read-only facilities before deciding whether a new plan is safe.

Saved plans are sensitive even though the structural summary is sanitized.
Do not copy run directories into source control, attach them to public issues,
or parse `plan.tfplan` outside the trusted local environment.

## Retained status and logs

`ainfra status DEPLOYMENT` derives lifecycle state only from private run
records and events. It does not inspect OpenTofu state. Interrupted mutations
are inspection-required and are never marked safe for automatic retry.

`ainfra logs DEPLOYMENT --run RUN_ID` shows the sanitized ainfra event
timeline. Add `--errors` to select typed failed, cancelled, and
inspection-required states; it never searches localized prose. Raw child
evidence requires an explicitly retained stream, child source, stream, and
interactive confirmation:

```sh
ainfra logs DEPLOYMENT --run RUN_ID --source opentofu \
  --raw --stream stderr
```

Automation must add both `--non-interactive` and `--yes`. Raw access is
incompatible with `--errors` and JSON output. Selected bytes go directly to
stdout and the warning goes to stderr; raw bytes never enter the result
envelope or normal renderer. Stream files must be private regular files inside
the selected run.

For Ansible-backed runs, `--source ansible-runner` reads the retained native
Runner v2 job-event artifacts. `--errors` selects only native failure,
unreachable, asynchronous-failure, and error event types. Words such as
`ERROR` in localized `stdout` never determine classification, and event data
is not copied into the sanitized display. Unsupported Runner protocol majors,
symlinks, public files, excessive trees, and oversized artifacts fail closed.


---
Source: https://projectious-work.github.io/ainfra/docs/reviewed-plans/index.md
