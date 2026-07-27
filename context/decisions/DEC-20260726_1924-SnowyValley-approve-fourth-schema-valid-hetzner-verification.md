---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1924-SnowyValley-approve-fourth-schema-valid-hetzner-verification
  created: '2026-07-26T19:24:46+00:00'
spec:
  title: Approve fourth schema-valid Hetzner verification run
  state: accepted
  decision: Apply fourth-test.tfplan for the six-resource first-test-20260726 environment,
    for at most one hour and EUR 0.0096 plus VAT estimated cost, run strict SSH, cloud-init
    schema, and full Ansible verification, then destroy all resources.
  context: The user explicitly approved the fourth exact test scope.
  rationale: The configuration now passes local cloud-init schema validation and preserves
    deterministic host trust.
  consequences: Destroy after success or failure.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T19:24:46+00:00'
---
