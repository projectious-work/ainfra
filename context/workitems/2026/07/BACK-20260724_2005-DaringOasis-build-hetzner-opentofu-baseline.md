---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-DaringOasis-build-hetzner-opentofu-baseline
  created: '2026-07-24T20:05:10+00:00'
  updated: '2026-07-25T10:23:10+00:00'
spec:
  title: Build secure Hetzner OpenTofu Kubernetes-ready baseline
  state: in-progress
  type: story
  priority: high
  description: Milestone 3 / PR 4. Implement pinned Hetzner provider/network/compute/firewall/cloud-init/state
    configuration for Debian 13 Kubernetes-ready hosts. Default to no public IPv4
    allocation, private management, zero optional workers, operator-supplied public
    keys, remote-state safety, ownership metadata, and secret-free standardized outputs.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  started_at: '2026-07-25T10:23:10+00:00'
---

## Transition note (2026-07-25T10:23:10+00:00)

Starting pinned Hetzner network, server, firewall, outputs, and cloud-init baseline.
