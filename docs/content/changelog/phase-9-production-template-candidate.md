---
title: "Phase 9 shipped: Private Hetzner K3s"
date: 2026-08-22
description: >-
  The first provider-backed template candidate adds private management,
  tunneled ingress, temporary administration, and pinned K3s bootstrap.
badges:
  - label: Phase 9
  - label: v1.0.0-alpha.9
    variant: accent
tags: [roadmap, templates, hetzner, kubernetes]
toc: true
---

Phase 9's implementation is live-certified. The template keeps
provider and host semantics in native OpenTofu and Ansible, emits the standard
non-secret inventory contract, and preserves ainfra's reviewed lifecycle.

The cost-approved disposable deployment converged with zero change, completed
Ansible check mode without drift, removed its temporary bastion without losing
Cloudflare Service Auth SSH access, and retained a healthy three-server K3s
control plane. Its exact destroy plan completed successfully, and independent
Hetzner API queries confirmed that no Phase 9 resources remained.
