---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1935-PolishedPond-approve-sixth-hetzner-verification-run
  created: '2026-07-26T19:35:18+00:00'
spec:
  title: Approve sixth Hetzner verification run
  state: accepted
  decision: Apply sixth-test.tfplan for the six-resource first-test-20260726 environment,
    for at most one hour and EUR 0.0096 plus VAT estimated cost, run full strict SSH,
    cloud-init, Ansible check/apply/idempotence verification, and destroy.
  context: The user explicitly approved the sixth exact test scope.
  rationale: All package-backed service tasks were swept for pristine-host check-mode
    behavior and locally validated.
  consequences: Destroy after success or failure.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T19:35:18+00:00'
---
