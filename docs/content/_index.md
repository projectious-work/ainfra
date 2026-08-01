---
title: ainfra
description: >-
  Security-first, contract-driven infrastructure templates built with
  OpenTofu and Ansible.
layout: hextra-home
---

{{< hextra/hero-badge
  link="docs/concepts/security-model/"
  class="ainfra-eyebrow"
>}}
  Cloud · Security · Automation
{{< /hextra/hero-badge >}}

<div class="ainfra-hero">
{{< hextra/hero-headline >}}
  Security-first infrastructure templates for containerized AI agents.
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  `ainfra` turns explicit, versioned contracts into reviewable OpenTofu
  plans, hardened hosts, and sanitized outputs. The tools remain visible.
  The safety boundaries remain enforceable.
{{< /hextra/hero-subtitle >}}

<div class="ainfra-hero__actions">
  <a class="ainfra-button ainfra-button--primary"
    href="{{< relref "/docs/getting-started/quickstart" >}}">
    Get started
  </a>
  <a class="ainfra-button ainfra-button--secondary"
    href="{{< relref "/docs/concepts/security-model" >}}">
    Read the security model
  </a>
</div>
</div>

<div class="ainfra-introduction">
Infrastructure automation should make every important boundary easier to see.
`ainfra` validates inputs before mutation, binds apply to an exact reviewed
plan, keeps secrets out of ordinary outputs, and makes teardown a first-class
operation.
</div>

<section class="ainfra-workflow" aria-labelledby="workflow-title">
  <div class="ainfra-section-heading">
    <p class="ainfra-kicker">A reviewable path to production</p>
    <h2 id="workflow-title">From contract to teardown</h2>
    <p>
      Every consequential step stays explicit, inspectable, and repeatable.
      No hidden control plane stands between you and the underlying tools.
    </p>
  </div>
  <ol class="ainfra-workflow__steps">
    <li>
      <span aria-hidden="true">01</span>
      <strong>Validate</strong>
      Reject unknown inputs before infrastructure can change.
    </li>
    <li>
      <span aria-hidden="true">02</span>
      <strong>Plan</strong>
      Review the exact OpenTofu plan and its security boundaries.
    </li>
    <li>
      <span aria-hidden="true">03</span>
      <strong>Apply</strong>
      Bind mutation to the reviewed plan, then configure with Ansible.
    </li>
    <li>
      <span aria-hidden="true">04</span>
      <strong>Destroy</strong>
      Keep teardown tested and available from the first deployment.
    </li>
  </ol>
</section>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    title="Contract driven"
    icon="document-text"
    subtitle="Strict contracts reject unknown fields and unsupported versions."
    link="docs/reference/contracts/"
    class="ainfra-feature-card"
  >}}
  {{< hextra/feature-card
    title="Security first"
    icon="shield-check"
    subtitle="Private networking and verified trust are secure defaults."
    link="docs/concepts/security-model/"
    class="ainfra-feature-card"
  >}}
  {{< hextra/feature-card
    title="Direct tools stay visible"
    icon="cube"
    subtitle="OpenTofu and Ansible remain visible, inspectable, direct tools."
    link="docs/concepts/architecture/"
    class="ainfra-feature-card"
  >}}
{{< /hextra/feature-grid >}}

<section class="ainfra-teardown">
  <div>
    <h2>Start disposable. Keep teardown ready.</h2>
    <p>
      The first reference template provisions a Kubernetes-ready Debian
      baseline on Hetzner Cloud. Plan, verify, apply, configure, and destroy
      it through explicit local commands.
    </p>
  </div>
  <a class="ainfra-button ainfra-button--light"
    href="{{< relref "/docs/getting-started/quickstart" >}}">
    Read the quickstart
  </a>
</section>
