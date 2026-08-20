---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260817_1405-GiftedRiver-adopt-immutable-release-integrity-checkpoints
  created: '2026-08-17T14:05:56+00:00'
spec:
  title: Adopt immutable release integrity checkpoints
  state: accepted
  decision: Release workflows must canonicalize and align repository lanes before
    freezing an exact commit, then enforce package, gate, sign, and resumable publication
    stages against immutable release state. Reusable controls should be expressed
    abstractly for adoption across unrelated internal repositories.
  context: The v1.0.0-alpha.7 release repeatedly invalidated human-approved evidence
    because signing occurred before the canonical release commit and because publication
    was not restart-safe.
  rationale: Mechanically enforced ordering moves expensive human actions after all
    mutable repository operations, preserves provenance, supports safe recovery, and
    creates a substrate suitable for organization-wide software delivery standards.
  alternatives:
  - option: Retain procedural documentation only
    reason: Rejected because sequencing mistakes remain possible and expensive.
  - option: Automate only this repository's exact branch names
    reason: Rejected as insufficiently reusable for diverse repositories.
  consequences: Release scripts and tests will gain explicit state validation and
    recovery behavior. Existing ad hoc release runs without frozen state will fail
    closed until prepared through the new workflow.
  decided_at: '2026-08-17T14:05:56+00:00'
---
