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
