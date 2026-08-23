---
title: ainfra
eyebrow: Agent-native infrastructure execution
tagline: >-
  Let agents create and operate standard OpenTofu and Ansible templates while
  plans, approvals, credentials, execution, and evidence remain governed.
cta:
  - label: Get started
    href: /docs/quick-start/
  - label: Why ainfra
    href: /docs/why-ainfra/
    variant: secondary
---

## Agent speed with an explicit execution boundary

{{< cards cols="3" >}}
  {{< card
    title="MCP-first"
    subtitle="Give agents typed inspection and guarded lifecycle operations without arbitrary shell execution."
    link="/docs/mcp/"
    icon="circle-check"
  >}}
  {{< card
    title="Native tools"
    subtitle="Keep OpenTofu and Ansible inputs visible instead of hiding them behind another configuration language."
    link="/docs/introduction/"
    icon="file-code"
  >}}
  {{< card
    title="Reviewed execution"
    subtitle="Bind approval to an exact plan, then retain sanitized evidence and recovery state."
    link="/docs/reviewed-plans/"
    icon="list"
  >}}
{{< /cards >}}

ainfra separates probabilistic interpretation and template creation from
governed validation, authorization, credential delivery, execution, and
evidence. Read [why ainfra exists](/docs/why-ainfra/) for the reasoning and
division of labour.

## A guarded lifecycle from source to teardown

ainfra resolves immutable template sources, validates native deployment
contracts, creates reviewed plans, applies only their exact bound artifacts,
configures hosts through deterministic inventory, checks convergence, and
supports independently reviewed destruction.

{{< callout type="info" title="Current maturity" >}}
The current release is an alpha. It is suitable for controlled evaluation and
disposable environments. `v1.0.0-alpha.9` includes the first provider-backed
production candidate. Its cost-approved live certification covered a
three-node private K3s lifecycle, tunnel-only administration, convergence,
exact destruction, and an independent provider leak check.
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
