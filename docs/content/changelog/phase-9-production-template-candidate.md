---
title: "Phase 9 candidate: Private Hetzner K3s"
date: 2026-08-21
description: >-
  The first provider-backed template candidate adds private management,
  tunneled ingress, temporary administration, and pinned K3s bootstrap.
badges:
  - label: Phase 9
  - label: certification pending
    variant: accent
tags: [roadmap, templates, hetzner, kubernetes]
toc: true
---

Phase 9's implementation is ready for offline evaluation. The template keeps
provider and host semantics in native OpenTofu and Ansible, emits the standard
non-secret inventory contract, and preserves ainfra's reviewed lifecycle.

Live certification is deliberately separate. It will be recorded only after a
cost-approved disposable deployment converges with zero change, removes its
temporary bastion without losing tunneled management, destroys cleanly, and
passes an independent provider inventory query.
