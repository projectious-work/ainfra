---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260820_1014-CrispFjord-phase-8-local-conformance-and-hosting
  created: '2026-08-20T10:14:36+00:00'
spec:
  title: Phase 8 local conformance and hosting boundary
  state: accepted
  decision: Phase 8 delivers local template-authoring conformance, clean-room documentation,
    and GitHub Pages publication. Any billable or live provider lifecycle remains
    separately approval-gated.
  rationale: This enables independently reproducible authoring and public guidance
    without treating a local check as authorization to create provider resources.
  consequences: The reference template can demonstrate local conformance. A future
    production template and any hosted provider lifecycle must capture explicit immediate
    approval and sanitized evidence.
  related_workitems:
  - BACK-20260820_1007-HopefulTulip-implement-phase-eight-template-authoring-conformance
  decided_at: '2026-08-20T10:14:36+00:00'
---
