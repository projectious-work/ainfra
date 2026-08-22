---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_1551-SoftBison-authorize-phase-9-hetzner-live-certification
  created: '2026-08-21T15:51:23+00:00'
spec:
  title: Authorize Phase 9 Hetzner live certification budget and teardown window
  state: accepted
  decision: Authorize the Phase 9 Hetzner live certification lifecycle in NBG1 using
    three CX23 control-plane nodes and an optional temporary CX23 bastion, with a
    maximum total Hetzner cost of EUR 1 including VAT, a normal teardown deadline
    of 8 hours after provisioning starts, and an absolute emergency teardown deadline
    of 24 hours.
  context: The Phase 9 production candidate passed all offline validation. Live certification
    requires a representative three-node provider lifecycle and explicit cost approval.
  rationale: Current Hetzner hourly pricing makes the expected six-hour lifecycle
    approximately EUR 0.24 including VAT and the 24-hour incident envelope approximately
    EUR 0.91, keeping the approved ceiling conservative.
  alternatives:
  - option: Do not run live certification
    reason: Leaves Phase 9 as an uncertified production candidate.
  - option: Use one control-plane node
    reason: Cheaper but not representative of the intended three-node topology.
  consequences: Provisioning may proceed only when required credentials, protected
    state, SSH trust inputs, Cloudflare fixture, and teardown controls are ready.
    All owned resources must be destroyed and independently verified absent before
    certification is claimed.
  related_workitems:
  - BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
  decided_at: '2026-08-21T15:51:23+00:00'
---
