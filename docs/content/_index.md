---
title: ainfra
description: >-
  Security-first, contract-driven infrastructure templates built with
  OpenTofu and Ansible.
params:
  body_class: td-navbar-links-all-active
---

{{% blocks/cover
  title="Infrastructure you can inspect, approve, and remove"
  image_anchor="top"
  height="full td-below-navbar"
%}}

<p class="ainfra-eyebrow">Cloud · Security · Automation</p>

`ainfra` turns explicit, versioned contracts into reviewable OpenTofu plans,
hardened hosts, and sanitized outputs. The tools remain visible. The safety
boundaries remain enforceable.
{.lead .display-6}

<div class="td-cta-buttons my-5">
  <a class="btn btn-lg btn-primary me-3"
    href="{{< relref "/docs/getting-started/quickstart" >}}">
    Get started
  </a>
  <a class="btn btn-lg btn-secondary"
    href="{{< relref "/docs/concepts/security-model" >}}">
    Read the security model
  </a>
</div>

{{% /blocks/cover %}}

{{% blocks/lead color="white" %}}

Infrastructure automation should make every important boundary easier to see.
`ainfra` validates inputs before mutation, binds apply to an exact reviewed
plan, keeps secrets out of ordinary outputs, and makes teardown a first-class
operation.

{{% /blocks/lead %}}

{{% blocks/section color="light" type="row" %}}

{{% blocks/feature icon="fa-file-contract" title="Contract driven" %}}

Strict JSON Schema contracts define template capabilities, inputs, and
non-secret outputs. Unknown fields and unsupported versions fail closed.

[Explore the contracts]({{< relref "/docs/reference/contracts" >}})

{{% /blocks/feature %}}

{{% blocks/feature icon="fa-shield-halved" title="Security first" %}}

Private networking, deliberate public-address allocation, verified SSH trust,
pinned dependencies, and local security gates are defaults rather than
afterthoughts.

[Understand the security model]({{< relref
"/docs/concepts/security-model" >}})

{{% /blocks/feature %}}

{{% blocks/feature icon="fa-layer-group" title="Direct tools stay visible" %}}

OpenTofu owns infrastructure state. Ansible owns host configuration. `ainfra`
orchestrates them without inventing a hidden infrastructure language.

[Read the architecture]({{< relref "/docs/concepts/architecture" >}})

{{% /blocks/feature %}}

{{% /blocks/section %}}

{{% blocks/section color="primary" type="row" %}}

<div class="col-lg-8">
<h2>Start disposable. Keep teardown ready.</h2>
<p class="lead">
The first reference template provisions a Kubernetes-ready Debian baseline on
Hetzner Cloud. Plan, verify, apply, configure, and destroy it through explicit
local commands.
</p>
</div>
<div class="col-lg-4 d-flex align-items-center justify-content-lg-end">
<a class="btn btn-lg btn-light"
  href="{{< relref "/docs/getting-started/quickstart" >}}">
  Read the quickstart
</a>
</div>

{{% /blocks/section %}}
