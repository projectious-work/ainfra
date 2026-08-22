---
title: Introduction
weight: 10
---

ainfra turns reusable infrastructure templates into repeatable, reviewable
deployments. It validates explicit contracts, locks template content, and
coordinates established provisioning and configuration tools.

Templates combine OpenTofu for infrastructure provisioning with Ansible for
system configuration. ainfra verifies the boundary between lifecycle steps
and retains sanitized evidence without hiding either underlying tool.

The Go-based v1 line is in alpha. Phases 0 through 9 are released as
`v1.0.0-alpha.9`, including the local template-authoring conformance contract,
clean-room proof, and the first provider-backed production candidate. The
Hetzner private K3s baseline passed its cost-approved disposable live
certification, tunnel-only management check, and independent teardown audit.
