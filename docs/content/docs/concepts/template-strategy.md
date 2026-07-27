---
title: Template strategy
weight: 40
description: How ainfra turns versioned template source into reviewable infrastructure.
---

An ainfra template is a versioned, self-contained infrastructure
implementation behind a common contract. Templates share lifecycle and safety
rules, but keep provider-specific OpenTofu, host configuration, and operational
documentation together.

## Why templates exist

The template boundary separates three concerns:

1. The ainfra wrapper owns discovery, contract validation, reviewed-plan
   binding, sanitized outputs, and ownership-scoped destruction.
2. A template owns provider resources, host configuration, supported images,
   capabilities, and provider-specific defaults.
3. The operator owns the selected template, non-secret intent, credential
   references, reviewed plan, and explicit lifecycle approvals.

This lets a template evolve without hiding OpenTofu or Ansible. Every engine
working directory remains directly usable and inspectable.

## Source and discovery

All bundled templates are direct children of the repository's `templates/`
directory:

```text
templates/
└── hetzner-kubernetes-baseline/
    ├── ainfra-template.yaml
    ├── README.md
    ├── inputs/
    ├── tofu/
    ├── cloud-init/
    └── ansible/
```

The directory name is the template identifier. `ainfra` accepts that
identifier, resolves exactly `templates/<identifier>/`, and requires an
`ainfra-template.yaml` whose `metadata.name` matches the directory. It does
not search arbitrary paths or download templates at runtime.

See the [template catalog]({{< relref "/docs/reference/templates" >}}) for the
templates currently included in the repository.

## What “build” means

Templates are source, not compiled packages. Building a template means
assembling and validating a coherent directory:

```text
manifest + input contract + OpenTofu + Ansible + examples + tests
    → contract and path validation
    → static policy and engine validation
    → reviewed OpenTofu plan bound to template content
    → apply and host configuration
    → sanitized InfrastructureOutput
```

The manifest declares the template version, compatible wrapper versions,
engine working directories, input and output schemas, output location,
capabilities, and mandatory security properties. During planning, ainfra
hashes the complete template directory. Apply refuses to continue if the
template, input, environment, version, or reviewed plan bytes have changed.

Shared schemas live in `schemas/`. Template-specific implementation and
examples stay under the template directory. This prevents a template from
silently weakening the common consumer contract.

## Evolution rules

- Patch versions correct behavior without changing the declared contract.
- Minor versions may add compatible options or capabilities.
- Breaking contract changes require the repository's migration process; the
  `v1alpha1` API cannot be silently replaced.
- Provider, module, collection, role, image, and scanner dependencies remain
  pinned.
- A new capability must be supported by the wrapper before a manifest can
  declare it.
- Live-support claims require a disposable apply, idempotence check, complete
  teardown, and redacted evidence.
