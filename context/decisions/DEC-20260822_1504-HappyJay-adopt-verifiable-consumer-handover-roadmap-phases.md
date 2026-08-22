---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260822_1504-HappyJay-adopt-verifiable-consumer-handover-roadmap-phases
  created: '2026-08-22T15:04:59+00:00'
spec:
  title: Adopt verifiable consumer handover roadmap phases
  state: accepted
  decision: 'Accept PR 106''s roadmap update: add Phase 10 for a versioned non-secret
    consumer target handover with signed deployment provenance, renumber the existing
    later phases, and add Phase 22 for confidential computing and live attestation.'
  context: The user reviewed and accepted PR 106 and explicitly authorized its merge
    while preparing the Phase 9 alpha release.
  rationale: The roadmap separates deterministic deployment provenance from current-state
    and hardware attestation, keeps secret references symbolic, preserves one standardized
    infrastructure-result family, and establishes a bounded ainfra-to-aibox handover
    direction.
  consequences: Phase 10 becomes the next product-contract direction after Phase 9;
    prior phases 10 through 20 become 11 through 21; confidential computing becomes
    Phase 22. Phase 9 remains release-scoped and must be marked shipped for alpha.9.
  decided_at: '2026-08-22T15:04:59+00:00'
---
