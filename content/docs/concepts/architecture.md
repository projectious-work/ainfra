---
title: Architecture
weight: 10
description: Tool ownership, public contracts, and lifecycle boundaries.
---

## Portfolio boundary

`ainfra` provisions and configures infrastructure targets. It hands a
non-secret output contract to downstream systems:

```text
InfrastructureTemplate + TemplateInput
                  │
                  ▼
        validation and policy
                  │
                  ▼
       OpenTofu plan and state
                  │
                  ▼
       Ansible host configuration
                  │
                  ▼
        InfrastructureOutput
                  │
                  ▼
        workload deployment
```

OpenTofu owns infrastructure desired state. Ansible owns host configuration.
The `ainfra` wrapper validates contracts and visibly orchestrates those tools.
Every wrapper operation has a documented direct-tool equivalent.

The first reference implementation is Hetzner-specific and produces
Kubernetes-ready hosts. It does not install or claim to operate Kubernetes.
Provider abstraction and workload deployment are outside the initial scope.

## Public contracts

Three strict JSON Schema draft 2020-12 contracts form the hand-off boundary:

1. `InfrastructureTemplate` declares engine locations, contract files,
   capabilities, and mandatory security invariants.
2. `TemplateInput` carries non-secret configuration and references to
   credentials or ignored local configuration.
3. `InfrastructureOutput` carries stable target identity, non-secret network
   facts, inventory metadata, and references to credentials.

Unsupported versions, kinds, capabilities, and unknown fields fail before
mutation. Template metadata is not an executable DSL and cannot declare
arbitrary shell hooks.

## Lifecycle

Planning creates an artifact that can be reviewed independently. Apply accepts
the exact plan identifier, not a general confirmation. Destroy accepts the
exact ownership scope, not a generic yes/no flag. These bindings prevent a
later command from silently acting on a different resource set.

Local state is permitted only for inputs explicitly marked disposable.
Long-lived environments require a capability-validated remote backend with
encryption, locking, version recovery, TLS, and access control.
