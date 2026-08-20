---
title: ainfra
eyebrow: Immutable infrastructure lifecycle
tagline: >-
  Deploy secured infrastructure for containerized AI through reviewed,
  reproducible OpenTofu and Ansible workflows.
cta:
  - label: Get started
    href: /docs/quick-start/
  - label: View the roadmap
    href: /docs/roadmap/
    variant: secondary
---

## Infrastructure you can inspect and reproduce

{{< cards cols="3" >}}
  {{< card
    title="Reviewed plans"
    subtitle="Create an immutable OpenTofu plan, review it, and apply exactly those verified bytes."
    link="/docs/reviewed-plans/"
    icon="circle-check"
  >}}
  {{< card
    title="Native tools"
    subtitle="Keep OpenTofu and Ansible inputs visible instead of hiding them behind another configuration language."
    link="/docs/introduction/"
    icon="file-code"
  >}}
  {{< card
    title="Durable evidence"
    subtitle="Retain sanitized lifecycle evidence, inventory, recovery guidance, and teardown verification."
    link="/docs/usage/"
    icon="list"
  >}}
{{< /cards >}}

## A guarded lifecycle from source to teardown

ainfra resolves immutable template sources, validates native deployment
contracts, creates reviewed plans, applies only their exact bound artifacts,
configures hosts through deterministic inventory, checks convergence, and
supports independently reviewed destruction.

{{< callout type="info" title="Current maturity" >}}
The current release is an alpha. It is suitable for controlled evaluation and
disposable environments while the first certified production template is being
prepared.
{{< /callout >}}

{{< cards cols="2" >}}
  {{< card
    title="Quick start"
    subtitle="Install the signed binary and exercise the guarded CLI lifecycle."
    link="/docs/quick-start/"
    icon="arrow-right"
  >}}
  {{< card
    title="Change log"
    subtitle="Follow shipped phases and release-level changes."
    link="/changelog/"
    icon="clock"
  >}}
{{< /cards >}}
