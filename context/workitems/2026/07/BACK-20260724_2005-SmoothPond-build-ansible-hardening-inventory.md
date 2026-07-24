---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-SmoothPond-build-ansible-hardening-inventory
  created: '2026-07-24T20:05:10+00:00'
spec:
  title: Build secret-free Ansible inventory and host-hardening roles
  state: backlog
  type: story
  priority: medium
  description: Milestone 4 / PR 5. Generate inventory from non-secret OpenTofu outputs
    and implement pinned, idempotent roles for base OS, administrator access, SSH,
    patching, reboot policy, firewalling, time, logging/audit, and observability prerequisites.
    Add lint, syntax, check-mode, and second-run idempotence coverage.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
---
