---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260820_0745-SmoothRaven-apply-approved-reconciliation-remediations
  created: '2026-08-20T07:45:18+00:00'
spec:
  title: Apply approved reconciliation remediations
  state: accepted
  decision: 'Apply all local, previously presented project-reconciliation remediations:
    the pending runtime migration, migration-briefing archival, and semantic-index
    refresh. Treat the flagged project email addresses as intentional/public and leave
    them unchanged.'
  rationale: The user explicitly approved every confirmation-required remediation
    identified during the reconciliation.
  consequences: The pending migration will be transitioned through its lifecycle;
    historical briefing files may be archived; the semantic index will be refreshed.
    GitHub issue and pull-request inventory remains blocked until credentials are
    supplied.
  decided_at: '2026-08-20T07:45:18+00:00'
---
