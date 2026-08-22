# AI template-authoring guide


This is the self-contained execution entry point for an AI agent authoring an
ainfra v1 template. It is a conformance workflow, not permission to provision
a live environment.

## Objective

Produce a portable template whose ainfra documents, native OpenTofu and
optional Ansible content can be validated from a clean room. Do not invent an
ainfra variable language, generate credentials, contact a provider during local
validation, or claim live support without disposable lifecycle evidence.

## Required deliverables

Create this minimum tree. Omit `tofu/outputs.tf`, `ansible/`,
`ansible-vars.yaml` and the output fixture when the manifest declares
`inventory: none` and no other output is needed:

```text
ainfra-template.yaml
README.md
docs/variables.md
tofu/versions.tf
tofu/variables.tf
tofu/outputs.tf
ansible/
examples/minimal/ainfra.yaml
examples/minimal/terraform.tfvars
examples/minimal/ansible-vars.yaml
tests/README.md
tests/validate.sh
tests/fixtures/output.json
```

The complete provider-free
[reference template](https://github.com/projectious-work/ainfra/tree/v1.x-dev/spec/examples/v1/template-example)
is an example, not hidden context required to use this guide.

## Manifest

Use this infrastructure-only starting point and change the name, version and
engine constraint deliberately. The containing directory must have the same
name as `metadata.name`.

```yaml
apiVersion: ainfra.projectious.work/v1
kind: Template
metadata:
  name: example-template
  version: 1.0.0
spec:
  engines:
    tofu:
      directory: tofu
      version: ">=1.10.0 <2.0.0"
  outputs:
    inventory: none
```

For configurable hosts, add an Ansible engine with `directory`, `playbook` and
an explicit version range, then set `inventory` to the exact OpenTofu output
name. The accepted manifest fields are defined by the
[published schema](https://github.com/projectious-work/ainfra/blob/v1.x-dev/spec/schemas/v1/template-manifest.schema.json);
unknown fields are invalid.

## Native variables and dependencies

Declare infrastructure inputs in `tofu/variables.tf` and host-configuration
inputs in native Ansible variable files. ainfra passes deployment-selected
native files through unchanged. Pin OpenTofu providers and modules; commit
`.terraform.lock.hcl` whenever provider selections exist. Pin every external
Ansible collection and role in `requirements.yml`.

`docs/variables.md` contains these H2 sections in this exact order:

1. `OpenTofu variables`
2. `Ansible variables`
3. `Cross-variable rules`
4. `Examples`

Each engine section uses:

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `example_name` | `string` | yes | — | Non-empty | no | Operational effect. |

Document every deployment-facing native variable. Reproduce types, defaults,
validation and sensitivity accurately, and explain replacement, access, cost
or teardown consequences. Examples must be minimal, valid and non-secret.

## Standard output

An inventory-producing template emits one non-sensitive OpenTofu output whose
name matches the manifest. Its JSON shape is:

```json
{
  "schema_version": "1",
  "hosts": {
    "node": {
      "groups": ["all"],
      "connection": {"type": "local"}
    }
  }
}
```

Use only the connection types and fields allowed by the
[standard-output schema](https://github.com/projectious-work/ainfra/blob/v1.x-dev/spec/schemas/v1/standard-output.schema.json).
Outputs must contain stable routing facts, never credentials or private keys.

## README and security

The template README documents purpose and unsupported outcomes, provider and
account prerequisites, least-privilege credentials, an architecture diagram,
variables, network/access exposure, costs and billable opt-ins, backend/state
ownership, the full lifecycle, equivalent native commands, failure/recovery,
teardown with independent verification, compatibility and dated validation
limitations.

Default public ingress off. Keep management addresses private where supported,
make any temporary administrative ingress explicit and narrow, verify SSH host
keys independently, label owned resources, keep credentials external, and make
destroy cover every resource the template owns. Fixtures and evidence contain
no secret value.

## Local validation

`tests/validate.sh` is executable and runs, as applicable:

```sh
tofu -chdir=tofu fmt -check
tofu -chdir=tofu init -input=false
tofu -chdir=tofu validate
ainfra doctor template . --format json
./tests/validate.sh
```

The script also checks dependency pins, documented-variable drift, secret
policy, standard output fixtures, Ansible syntax, and a zero-change check-mode
pass. The doctor reports native checks as delegated; its `skip` is an
instruction to run the script, not a conformance success.

## Disposable live acceptance

Stop before this section until a human gives explicit, immediate approval for
the named provider account, region, cost ceiling and teardown window. After
approval: create and review a plan; apply exactly that plan; configure; verify
zero-change convergence; reapply only if the template requires that proof;
create and review a destroy plan; destroy; and independently query the provider
to prove owned resources are absent.

Retain sanitized evidence containing template revision/digest, date, operator,
provider and tool versions, approved scope/cost ceiling, plan summaries,
configure/check results, destroy result, independent teardown confirmation,
limitations and final pass/fail. Never retain state, raw credentials or secret
values in published evidence.

## Completion checklist

- [ ] Manifest identity, version, supported tools and input/output contracts
      are accurate.
- [ ] Provider/module and Ansible dependencies are pinned and applicable native
      lockfiles are committed.
- [ ] The README explains prerequisites, architecture, variables, network,
      cost, failure behavior, teardown, compatibility and validation.
- [ ] `docs/variables.md` has the required headings and complete tables.
- [ ] A clean-room fixture exists and `tests/validate.sh` validates it.
- [ ] Variable drift, secret policy and standard output fixtures are tested.
- [ ] `ainfra doctor template . --format json` has no `fail` findings.
- [ ] `./tests/validate.sh` passes using compatible native tools.
- [ ] Live support is claimed only after the approved disposable lifecycle,
      independent teardown verification and sanitized evidence are complete.


---
Source: https://projectious-work.github.io/ainfra/docs/template-authoring-ai/index.md
