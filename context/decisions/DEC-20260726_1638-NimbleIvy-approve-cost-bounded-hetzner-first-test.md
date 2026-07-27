---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1638-NimbleIvy-approve-cost-bounded-hetzner-first-test
  created: '2026-07-26T16:38:06+00:00'
spec:
  title: Approve cost-bounded Hetzner first test apply
  state: accepted
  decision: Apply the saved first-test-20260726 OpenTofu plan for one CX23 server
    and its six planned resources, with an intended lifetime of at most one hour and
    an estimated maximum one-hour cost of EUR 0.0096 plus VAT, while retaining immediate
    same-state destroy readiness.
  context: The user explicitly accepted the exact project, resource, lifetime, and
    cost scope before the billable apply.
  rationale: This satisfies the repository live-verification approval gate.
  consequences: Apply only the saved reviewed plan. On safety failure, destroy through
    the same state immediately. Do not retain the environment beyond one hour.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T16:38:06+00:00'
---
