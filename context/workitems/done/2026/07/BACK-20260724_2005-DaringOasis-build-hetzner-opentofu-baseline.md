---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-DaringOasis-build-hetzner-opentofu-baseline
  created: '2026-07-24T20:05:10+00:00'
  updated: '2026-07-25T20:38:16+00:00'
spec:
  title: Build secure Hetzner OpenTofu Kubernetes-ready baseline
  state: done
  type: story
  priority: high
  description: Milestone 3 / PR 4. Implement pinned Hetzner provider/network/compute/firewall/cloud-init/state
    configuration for Debian 13 Kubernetes-ready hosts. Default to no public IPv4
    allocation, private management, zero optional workers, operator-supplied public
    keys, remote-state safety, ownership metadata, and secret-free standardized outputs.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  started_at: '2026-07-25T10:23:10+00:00'
  completed_at: '2026-07-25T20:38:16+00:00'
---

## Transition note (2026-07-25T10:23:10+00:00)

Starting pinned Hetzner network, server, firewall, outputs, and cloud-init baseline.


## Transition note (2026-07-25T20:37:56+00:00)

OpenTofu baseline implemented with IPv4/IPv6 opt-in, firewall attachment at server creation, narrow direct-tool management CIDR validation, exact provider lock, and passing local validation.


## Transition note (2026-07-25T20:38:16+00:00)

Independent re-review found no remaining merge blocker. 59 tests, formatting, lint, type checks, and OpenTofu validate pass.
