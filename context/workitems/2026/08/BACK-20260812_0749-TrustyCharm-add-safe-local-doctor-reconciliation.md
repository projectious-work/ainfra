---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-TrustyCharm-add-safe-local-doctor-reconciliation
  created: '2026-08-12T07:49:16+00:00'
spec:
  title: Add safe local doctor reconciliation
  state: backlog
  type: story
  priority: medium
  description: Plan and apply only explicitly approved ainfra-owned local repairs
    under a deployment lock, with containment, precondition rechecks, interruption
    evidence, post-checks, idempotency, confirmation, and redaction.
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
---
