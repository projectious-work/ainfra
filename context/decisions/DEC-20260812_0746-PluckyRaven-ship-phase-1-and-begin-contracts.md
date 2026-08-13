---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260812_0746-PluckyRaven-ship-phase-1-and-begin-contracts
  created: '2026-08-12T07:46:38+00:00'
spec:
  title: Ship Phase 1 and begin Contracts and doctor
  state: accepted
  decision: Mark roadmap Phase 1 shipped and Phase 2 Contracts and doctor in progress,
    plan Phase 2 from the normative v1 specification, and implement it on a dedicated
    topic branch based on current v1.x-dev.
  context: Phase 1 was published and independently verified as v1.0.0-alpha.1, and
    its release-evidence work item is complete.
  rationale: The roadmap orders Contracts and doctor immediately after the Go foundation.
    A dedicated topic branch preserves the v1 promotion lanes while Phase 2 evolves.
  alternatives:
  - option: Continue Phase 1
    reason_not_chosen: All Phase 1 implementation and release gates are complete.
  - option: Start a later lifecycle phase
    reason_not_chosen: Later phases depend on Phase 2 discovery, configuration, parsing,
      validation, diagnostics, and reconciliation capabilities.
  consequences: Roadmap status and developer notes become authoritative for Phase
    2. Implementation work must trace to a Phase 2 disposition matrix and prioritized
    work items.
  decided_at: '2026-08-12T07:46:38+00:00'
---
