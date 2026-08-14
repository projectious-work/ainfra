# template-example

`template-example` is the minimal provider-free reference template for the
ainfra v1 contract. It exercises native OpenTofu variables, the standardized
inventory output, deterministic inventory generation, native Ansible variables,
Runner outcomes, check mode, and destroy-plan handling entirely on localhost.

It is intended for template authors, CLI implementers, and AI agents testing
contract conformance. It does not provision cloud infrastructure, configure a
remote host, prove provider security, or qualify as a certified provider
template.

## Quick start

After building ainfra and installing compatible OpenTofu and Ansible Runner:

```sh
cd examples/minimal
ainfra doctor
ainfra template lock
ainfra plan
```

Review the saved plan and follow the exact `ainfra apply --plan <run-id>` and
`ainfra configure --run <run-id>` commands printed by the CLI.

## Prerequisites and credentials

- OpenTofu `>=1.10.0,<2.0.0`;
- Ansible Runner backed by Ansible Core `>=2.18.0,<3.0.0`; and
- a POSIX-like local execution environment supported by ainfra.

No provider account, cloud credential, SSH credential, secret, external role,
or Ansible collection is required. The template uses OpenTofu's built-in
`terraform_data` resource and Ansible's local connection plugin.

## Architecture

```text
native tfvars -> OpenTofu terraform_data -> ainfra_inventory output
                                                |
                                                v
native Ansible vars -> generated localhost inventory -> Ansible Runner
```

OpenTofu owns its resource lifecycle. Ansible owns task execution and reported
outcomes. ainfra validates and transports the standardized non-secret handoff
and retains execution evidence; it does not model either engine's state.

## Variables

The complete deployment-facing API is in
[`docs/variables.md`](docs/variables.md). The ready-to-run native inputs are in
[`examples/minimal/`](examples/minimal/).

## Network and access model

All execution targets the local machine through `ansible_connection: local`.
The example opens no listener, creates no network resource, uses no SSH
connection, and therefore needs no `known_hosts` entry.

## Cost

The example creates no billable resource. OpenTofu records only a built-in
local `terraform_data` resource in the backend selected by OpenTofu.

## ainfra lifecycle

From `examples/minimal/`, run `doctor`, lock the local template, create and
review a plan, apply that exact plan, generate inventory, and configure. A
second configuration run in check mode must report zero changed, failed, and
unreachable hosts. Teardown requires a separately reviewed destroy plan.

The CLI prints the concrete run ID required by later commands. The
specification intentionally does not prescribe a hand-written run-ID value.

## Equivalent native commands

These commands, run from a disposable copy of the template root, provide a
diagnosis and recovery path. ainfra may use a private materialized workspace
and different contained paths:

```sh
tofu -chdir=tofu fmt -check
tofu -chdir=tofu init
tofu -chdir=tofu validate
tofu -chdir=tofu plan -var-file=../examples/minimal/terraform.tfvars -out=plan.tfplan
tofu -chdir=tofu apply plan.tfplan
tofu -chdir=tofu output -json ainfra_inventory
ansible-playbook -i tests/fixtures/inventory.yaml -e @examples/minimal/ansible-vars.yaml ansible/site.yml
tofu -chdir=tofu plan -destroy -var-file=../examples/minimal/terraform.tfvars -out=destroy.tfplan
tofu -chdir=tofu apply destroy.tfplan
```

ainfra uses Ansible Runner for structured execution evidence; `ansible-playbook`
is shown because it is the simplest direct Ansible diagnosis path. The ainfra
command and test fixtures are authoritative for the standardized handoff.

## Failure and recovery

- Schema or path failures must be corrected before engine execution.
- An OpenTofu validation or plan failure is diagnosed with the equivalent
  native command above.
- Ansible assertion failures indicate a missing inventory or variable handoff;
  inspect sanitized Runner events and rerun only after correcting the input.
- An interrupted apply, configure, or destroy is not automatically repeated;
  inspect `ainfra status` and the owning engine before creating a new plan.

Deleting `.ainfra/` removes local execution evidence and saved plans, not
OpenTofu-managed resources. Recovery then requires explicit engine inspection
and a new reviewed plan.

## Teardown and verification

Apply only the exact reviewed destroy plan. Success means OpenTofu reports that
the built-in resource was destroyed. There is no independent provider API to
inspect; verify the expected local artifact is absent without reading OpenTofu
state. A provider-backed certified template MUST instead document an
authenticated provider inventory/API check scoped by its ownership labels.

## Compatibility and versioning

The manifest declares the supported engine ranges. Breaking changes to native
variables, standardized outputs, required tools, or directory contracts require
a new template major version and migration guidance. Additive optional inputs
may be released in a minor version.

## Validation status

The schemas, documentation structure, native inputs, and deterministic fixtures
were reviewed on 2026-08-06. The provider-free native lifecycle passed with
OpenTofu 1.12.5 and Ansible Core 2.18.7, including plan, apply, output matching,
normal configuration, zero-change check mode, destroy plan, and destroy apply.
Full ainfra black-box validation remains pending implementation of the v1 Go
CLI. This is a conforming reference template, not evidence of a released CLI or
certified provider deployment.

See [`tests/`](tests/) for the executable lifecycle acceptance sequence and
expected artifacts.
