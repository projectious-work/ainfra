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

The Go-based v1 line is in alpha. Phases 0 through 7 are released. The
`v1.x-dev` branch contains the Phase 8 local template-authoring conformance
candidate; it is not part of `v1.0.0-alpha.7`. The first certified production
template remains the next planned milestone.
