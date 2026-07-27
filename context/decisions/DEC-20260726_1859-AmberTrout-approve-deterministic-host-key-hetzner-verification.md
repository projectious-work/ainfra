---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1859-AmberTrout-approve-deterministic-host-key-hetzner-verification
  created: '2026-07-26T18:59:59+00:00'
spec:
  title: Approve deterministic-host-key Hetzner verification run
  state: accepted
  decision: Apply the saved second-test plan for the six-resource first-test-20260726
    environment, run deterministic-host-key SSH and full Ansible verification, and
    destroy the environment afterward within one hour.
  context: The user accepted the second deployment at an estimated EUR 0.0096 plus
    VAT.
  rationale: The deterministic host key closes the trusted SSH identity gap found
    by the first live run.
  consequences: Use only the saved plan; prepare destroy immediately; destroy after
    success or failure.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  decided_at: '2026-07-26T18:59:59+00:00'
---
