---
title: Documentation
linkTitle: Documentation
description: >-
  Guides and reference material for secure, contract-driven infrastructure
  templates.
---

`ainfra` is the infrastructure provisioning layer for projectious.work. It
combines strict contracts with a thin, local wrapper around OpenTofu and
Ansible:

<div class="ainfra-flow" role="img"
  aria-label="ainfra provisions a target, aibox deploys workloads, and
  processkit reconciles workspace content">
ainfra apply → provisioned target + non-secret output contract<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;→ aibox deploys workloads to that target<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;→ processkit reconciles workspace content
</div>

The project provisions targets. It does not build workload images, install
Kubernetes or processkit, deploy aibox fleets, or hide OpenTofu and Ansible
behavior.

## Find your path

| If you want to… | Start here |
|---|---|
| Validate and provision a disposable target | [Quickstart]({{< relref "/docs/getting-started/quickstart" >}}) |
| Understand ownership and tool boundaries | [Architecture]({{< relref "/docs/concepts/architecture" >}}) |
| Review threat assumptions and invariants | [Security model]({{< relref "/docs/concepts/security-model" >}}) |
| Operate plan, apply, output, and destroy | [Lifecycle operations]({{< relref "/docs/guides/lifecycle" >}}) |
| Author another infrastructure template | [Template authoring]({{< relref "/docs/guides/authoring-templates" >}}) |
| Look up commands or schemas | [Reference]({{< relref "/docs/reference" >}}) |

## Core promises

- Inputs are validated before infrastructure mutation.
- Apply identifies the exact reviewed plan.
- Destruction identifies the exact ownership scope.
- Credentials enter through references and child-process environments.
- Ordinary output contracts contain references, never secret values.
- Provider, image, automation, and scanner versions are pinned.
- Every automated gate is available locally.
