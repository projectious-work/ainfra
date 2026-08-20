# template-clean-room

This provider-free template is the Phase 8 clean-room authoring proof. It was
derived from the public authoring contract and AI guide. It does not configure
hosts, contact a provider, prove provider security, or claim live support.

## Prerequisites and credentials

Use OpenTofu `>=1.10.0,<2.0.0` on a supported POSIX host. No account,
credential, private key, password, external module or provider is required.

## Architecture

```text
terraform.tfvars -> terraform_data.clean_room -> local OpenTofu state
```

## Variables

The complete native API is in [`docs/variables.md`](docs/variables.md).

## Network and access

The template opens no listener and performs no network or remote-host action.

## Cost

The built-in `terraform_data` resource is local and has no billable cost.

## Lifecycle and native state

OpenTofu owns its backend and state. From a disposable copy of the template:

```sh
tofu -chdir=tofu fmt -check
tofu -chdir=tofu init
tofu -chdir=tofu validate
tofu -chdir=tofu plan -var-file=../examples/minimal/terraform.tfvars -out=plan.tfplan
tofu -chdir=tofu apply plan.tfplan
tofu -chdir=tofu plan -destroy -var-file=../examples/minimal/terraform.tfvars -out=destroy.tfplan
tofu -chdir=tofu apply destroy.tfplan
```

## Failure and recovery

Correct format or validation failures before planning. After an interrupted
mutation, inspect OpenTofu directly before deciding whether to continue.

## Teardown and verification

Apply the reviewed destroy plan and verify OpenTofu reports the local object
destroyed. There is no independent provider API because no provider is used.

## Compatibility

Breaking variable or engine changes require a new template major version and
migration guidance. Compatible optional additions use a minor version.

## Validation status

The schema, doctor checks and provider-free native lifecycle were validated on
2026-08-20. This is local conformance evidence, not live-support evidence.

## Validation

Run `ainfra doctor template . --format json`, then `./tests/validate.sh`.
