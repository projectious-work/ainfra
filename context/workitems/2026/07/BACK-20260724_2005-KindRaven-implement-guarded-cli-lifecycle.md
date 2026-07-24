---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-KindRaven-implement-guarded-cli-lifecycle
  created: '2026-07-24T20:05:10+00:00'
spec:
  title: Implement thin execution adapters and guarded CLI lifecycle commands
  state: backlog
  type: story
  priority: high
  description: Milestone 2 / PR 3. Implement plan, apply, destroy, and outputs around
    visible OpenTofu and Ansible subprocesses. Preserve output and exit behavior,
    resolve secrets only into child environments, require explicit approvals, apply
    reviewed plans, prevent hidden migration or repair, and test argv, redaction,
    signals, and failure ordering.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
---
