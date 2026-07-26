---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  created: '2026-07-24T20:05:10+00:00'
  updated: '2026-07-25T20:52:48+00:00'
spec:
  title: Integrate local security gates and final live verification
  state: in-progress
  type: story
  priority: medium
  description: Milestone 5 / PR 6. Integrate pinned Checkov and Gitleaks with all
    local schema, wrapper, OpenTofu, cloud-init, and Ansible gates. Add complete negative
    tests and sanitized evidence handling. Run billable Hetzner verification only
    as the final step after explicit user approval, then require separate explicit
    approval for ownership-scoped destroy.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  started_at: '2026-07-25T20:52:48+00:00'
---

## Transition note (2026-07-25T20:52:48+00:00)

Local security gate implementation started on feature/security-gates. Billable Hetzner live verification remains excluded pending separate explicit approval.
