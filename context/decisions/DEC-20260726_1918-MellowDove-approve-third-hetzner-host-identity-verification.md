---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1918-MellowDove-approve-third-hetzner-host-identity-verification
  created: '2026-07-26T19:18:27+00:00'
spec:
  title: Approve third Hetzner host-identity verification run
  state: accepted
  decision: Apply third-test.tfplan for the six-resource first-test-20260726 environment,
    for at most one hour and EUR 0.0096 plus VAT estimated cost, run strict SSH and
    full Ansible verification, then destroy all resources.
  context: The user explicitly approved the exact third test scope.
  rationale: Cloud-init host-key regeneration is now disabled so the injected identity
    should remain authoritative.
  consequences: Destroy after success or any failure.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T19:18:27+00:00'
---
