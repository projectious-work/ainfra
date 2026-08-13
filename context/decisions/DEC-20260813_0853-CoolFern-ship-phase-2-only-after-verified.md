---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260813_0853-CoolFern-ship-phase-2-only-after-verified
  created: '2026-08-13T08:53:03+00:00'
spec:
  title: Ship Phase 2 only after verified release publication
  state: accepted
  decision: Phase 2 will be marked shipped in the roadmap only after its independent
    conformance review is accepted, gaps are closed, the Phase 2 release is prepared,
    promoted, published, and verified, and the release identity is recorded in the
    phase evidence.
  context: 'The Phase 2 implementation merged in PR #63, but the roadmap remains in_progress.
    AINFRA-DEV-011 requires code, tests, specification, documentation, phase evidence,
    release material, and release identity to agree before the shipped transition.'
  rationale: This ordering prevents the roadmap from claiming shipped status before
    an externally verifiable release exists and keeps phase state aligned with the
    normative completion gate.
  alternatives:
  - option: Mark Phase 2 shipped before publishing
    reason_rejected: Would violate the release-identity agreement required by AINFRA-DEV-011.
  decided_at: '2026-08-13T08:53:03+00:00'
---
