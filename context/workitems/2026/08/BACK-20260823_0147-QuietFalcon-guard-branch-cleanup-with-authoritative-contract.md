---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260823_0147-QuietFalcon-guard-branch-cleanup-with-authoritative-contract
  created: '2026-08-23T01:47:27+00:00'
  labels:
    area: repository-governance
    lesson: session-2026-08-23
spec:
  title: Guard branch cleanup with the authoritative branching contract
  state: backlog
  type: task
  priority: high
  description: Update the repository cleanup workflow or guardrails so branch classification
    and deletion first load the git-branching version-line integration contract, inventory
    tags and consumers, and preserve every protected long-lived lane. Add coverage
    or an explicit checklist preventing v0.x-dev, v0.x-pre-release, v0.x-release,
    v0.x-maintenance, v1 lanes, main, or gh-pages from being labeled unused without
    contract evidence.
---
