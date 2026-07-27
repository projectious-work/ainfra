---
title: Authoring templates
weight: 30
description: Add a contract-compliant infrastructure template.
---

Every template is a direct child of `templates/` and contains
`ainfra-template.yaml`. Discovery does not traverse arbitrary paths or accept
absolute template locations.

Template authors must:

- validate against `template-manifest.v1alpha1.json`;
- keep OpenTofu and Ansible working directories beneath the template;
- use only declared capabilities;
- provide non-secret example inputs;
- pin every provider, collection, role, image, and scanner choice;
- document direct OpenTofu and Ansible equivalents;
- produce an `InfrastructureOutput/v1alpha1`;
- add positive and negative contract and policy fixtures;
- prove disposable teardown before claiming live support.

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
