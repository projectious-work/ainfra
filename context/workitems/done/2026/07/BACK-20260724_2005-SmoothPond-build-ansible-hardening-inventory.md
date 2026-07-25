---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-SmoothPond-build-ansible-hardening-inventory
  created: '2026-07-24T20:05:10+00:00'
  updated: '2026-07-25T20:51:45+00:00'
spec:
  title: Build secret-free Ansible inventory and host-hardening roles
  state: done
  type: story
  priority: medium
  description: Milestone 4 / PR 5. Generate inventory from non-secret OpenTofu outputs
    and implement pinned, idempotent roles for base OS, administrator access, SSH,
    patching, reboot policy, firewalling, time, logging/audit, and observability prerequisites.
    Add lint, syntax, check-mode, and second-run idempotence coverage.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  started_at: '2026-07-25T20:44:29+00:00'
  completed_at: '2026-07-25T20:51:45+00:00'
---

## Transition note (2026-07-25T20:44:29+00:00)

Implementation started on feature/ansible-hardening after Hetzner baseline merged to v0.1-dev.


## Transition note (2026-07-25T20:51:31+00:00)

Implementation complete: deterministic secret-free inventory, Debian 13 hardening roles, pinned lint/core, local lint/syntax gates, and approved-live check/idempotence script. Independent review has no blockers.


## Transition note (2026-07-25T20:51:45+00:00)

67 tests, Ruff, mypy, ansible-lint production profile, and Ansible syntax validation pass. Live check/apply/idempotence execution remains correctly gated by separate cost approval.
