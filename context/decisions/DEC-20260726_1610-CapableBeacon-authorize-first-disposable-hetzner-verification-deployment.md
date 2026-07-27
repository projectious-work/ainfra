---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260726_1610-CapableBeacon-authorize-first-disposable-hetzner-verification-deployment
  created: '2026-07-26T16:10:55+00:00'
spec:
  title: Authorize first disposable Hetzner verification deployment
  state: accepted
  decision: Proceed with the first disposable Hetzner test deployment after repository
    health and non-mutating readiness checks pass, while preparing and validating
    the exact ownership-scoped destroy path before apply.
  context: The user explicitly requested continuation to the first Hetzner test deployment
    and required readiness to tear it down immediately.
  rationale: This preserves the cost and destruction boundaries while allowing live
    verification of the secured infrastructure template.
  consequences: Provision only the minimal single-node disposable environment; keep
    local state and evidence ignored; do not expose credentials; verify teardown targeting
    before creation and destroy promptly on failure or user request.
  related_workitems:
  - BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  - BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  decided_at: '2026-07-26T16:10:55+00:00'
---
