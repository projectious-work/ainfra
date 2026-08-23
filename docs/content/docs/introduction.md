---
title: Introduction
weight: 10
---

ainfra is the agent-native infrastructure execution boundary for reviewed,
reproducible OpenTofu and Ansible deployments. MCP is the primary agent
experience; the complete CLI remains available to humans, CI, recovery, and
break-glass operation. Both call the same application lifecycle.

Agents can interpret goals, select or author templates, explain plans, and
propose remediation. ainfra validates explicit contracts, locks template
content, verifies authorization, coordinates established tools, and retains
sanitized evidence. Read [Why ainfra exists](../why-ainfra/) for the complete
division of labour and product rationale.

Templates combine OpenTofu for infrastructure provisioning with Ansible for
system configuration. ainfra verifies the boundary between lifecycle steps
and retains sanitized evidence without hiding either underlying tool.

The Go-based v1 line is in alpha. Phases 0 through 9 are released as
`v1.0.0-alpha.9`, including the local template-authoring conformance contract,
clean-room proof, and the first provider-backed production candidate. The
Hetzner private K3s baseline passed its cost-approved disposable live
certification, tunnel-only management check, and independent teardown audit.
