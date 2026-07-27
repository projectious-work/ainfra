---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1931-GentleSummit-approve-fifth-hetzner-full-verification-run
  created: '2026-07-26T19:31:01+00:00'
spec:
  title: Approve fifth Hetzner full verification run
  state: accepted
  decision: Apply fifth-test.tfplan for the six-resource first-test-20260726 environment,
    for at most one hour and EUR 0.0096 plus VAT estimated cost, run the full strict
    SSH, cloud-init, Ansible check/apply/idempotence sequence, and destroy all resources.
  context: The user explicitly approved the fifth exact test scope.
  rationale: The pristine-host check-mode service defect is fixed and locally validated.
  consequences: This is the final planned live deployment if all gates pass; destroy
    after success or failure.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T19:31:01+00:00'
---
