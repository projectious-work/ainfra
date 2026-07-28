---
title: Authoring templates
weight: 30
description: Add a contract-compliant infrastructure template.
---

Read the [template strategy]({{< relref
"/docs/concepts/template-strategy" >}}) first. This guide covers the concrete
authoring workflow.

## Choose adaptation or a new template

Adapt an existing template when the provider, resource topology, lifecycle,
and security model remain the same. Typical adaptations include:

- changing sizes, counts, locations, or opt-in networking through inputs;
- adding another explicitly supported operating-system image;
- extending an existing Ansible role;
- adding a compatible output or capability already understood by ainfra.

Create a new template when the provider changes, the topology has different
ownership or teardown behavior, the engine layout changes materially, or the
existing template's security promises would become misleading.

Do not fork a template merely to change an environment value. Copy its example
input and keep the template source unchanged instead.

## Find and inspect templates

Bundled templates are direct children of `templates/`. The current catalog is
listed in [Template reference]({{< relref "/docs/reference/templates" >}}).
Inspect the manifest and template README before changing anything:

```sh
find templates -mindepth 2 -maxdepth 2 -name ainfra-template.yaml -print
sed -n '1,240p' \
  templates/hetzner-kubernetes-baseline/ainfra-template.yaml
sed -n '1,240p' templates/hetzner-kubernetes-baseline/README.md
```

## Adapt the existing template

For an environment-specific adaptation:

1. Copy `templates/hetzner-kubernetes-baseline/inputs/example.input.yaml`
   outside the template's tracked examples.
2. Keep `metadata.template: hetzner-kubernetes-baseline`.
3. Change only documented input fields.
4. Keep credentials as references or environment variables, never values.
5. Validate the template and input together.
6. Review the generated plan and retain immediate teardown readiness.

```sh
cp templates/hetzner-kubernetes-baseline/inputs/example.input.yaml \
  /tmp/my-environment.input.yaml
ainfra validate hetzner-kubernetes-baseline \
  --input /tmp/my-environment.input.yaml
```

If the implementation itself must change, edit the OpenTofu, cloud-init, or
Ansible source in the existing directory, increment `metadata.version`, update
the supported-value documentation, and add tests for the changed behavior.

## Create a new template

Start from the nearest existing implementation only when its security defaults
are appropriate:

```sh
cp -R templates/hetzner-kubernetes-baseline templates/my-template
```

Then complete every step:

1. Rename the directory using lowercase letters, numbers, and hyphens.
2. Set the same identifier in `ainfra-template.yaml` at `metadata.name`.
3. Reset `metadata.version` for the new template and describe its purpose.
4. Declare the supported ainfra wrapper range.
5. Point both engine working directories to paths inside the template.
6. Reference only shared schemas that are direct children of `schemas/`.
7. Replace example inputs, embedded template identifiers, output metadata,
   provider resources, and README content.
8. Declare only capabilities supported by `src/template.rs`.
9. Pin OpenTofu providers and Ansible collections and roles.
10. Add positive and negative contract, policy, engine, and lifecycle tests.

Discovery does not traverse arbitrary paths or accept absolute template
locations. A directory outside `templates/`, a nested catalog, or a manifest
whose name differs from its directory is rejected.

## Required directory contract

The exact implementation directories can vary when declared in the manifest,
but a complete template has this shape:

```text
templates/my-template/
├── ainfra-template.yaml
├── README.md
├── inputs/
│   └── example.input.yaml
├── tofu/
│   ├── .terraform.lock.hcl
│   ├── versions.tf
│   ├── providers.tf
│   ├── variables.tf
│   └── outputs.tf
└── ansible/
    ├── requirements.yml
    └── site.yml
```

Cloud-init, roles, modules, and other source belong under the same template
root. Runtime state and generated output belong under ignored `.ainfra/` or
OpenTofu state locations, never in Git.

## Validate the result

Run narrow checks while authoring, then the complete repository gates:

```sh
ainfra validate my-template \
  --input templates/my-template/inputs/example.input.yaml
tofu -chdir=templates/my-template/tofu fmt -check -recursive
tofu -chdir=templates/my-template/tofu init -backend=false
tofu -chdir=templates/my-template/tofu validate
scripts/validate-all
scripts/test-all
```

Also document the direct OpenTofu and Ansible commands so operators can inspect
and troubleshoot without the wrapper. Before claiming provider support, run a
cost-approved disposable lifecycle: plan, apply, host verification, Ansible
check, apply, idempotence check, destroy, and provider/state confirmation.

## Author checklist

- The manifest validates against `template-manifest.v1alpha1.json`.
- All manifest paths remain inside the allowed template or schema roots.
- Example inputs are useful and contain no secrets.
- Supported values and exclusions are explicit.
- Standard outputs validate as `InfrastructureOutput/v1alpha1` and contain no
  secret-shaped data.
- New behavior has positive and negative tests.
- Dependencies and images are pinned.
- Direct engine commands and teardown are documented.
- The full local validation suite is green.

## Image choices

Every user-selectable image must list and explain all supported values. The
initial template supports:

```yaml
# Supported images:
# - debian-13: official Hetzner Debian 13 image; the default and only
#   initially validated operating system.
image: debian-13
```

Adding an image is not merely an enum change. It requires hardening,
architecture, networking, update, idempotence, and disposable-lifecycle
verification.

## Keep the contract narrow

Template metadata is declarative and cannot add arbitrary shell hooks. If a
new capability is needed, define and test its contract semantics before
implementing provider behavior.
