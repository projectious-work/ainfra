---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260809_1418-AbleFrog-support-macos-and-linux-in-the
  created: '2026-08-09T14:18:02+00:00'
spec:
  title: Support macOS and Linux in the host container gate
  state: accepted
  decision: Extend the reviewed two-stage container-gate launcher to support both
    macOS and Linux while preserving the same offline uv bootstrap, owner-approved
    exact Python digest, fixed executable paths, clean environment, and unchanged
    Python container-validation boundary.
  context: 'The initial PR #55-aligned launcher targeted macOS. Linux support is needed
    for equivalent owner-operated release evidence on Docker-capable Linux hosts.'
  rationale: The owner explicitly requested the previously proposed Linux portability
    work. Platform-specific account lookup, cache roots, checksum utilities, trusted
    tool paths, and installation guidance can be isolated in the launcher without
    broadening container authority.
  alternatives:
  - option: Keep macOS-only support
    reason: Rejected because it does not meet the owner's portability requirement.
  - option: Maintain separate launchers
    reason: Rejected because duplicated security-sensitive logic would increase review
      and drift risk.
  consequences: The controlled launcher requires renewed owner review. Acceptance
    must cover simulated platform branches locally and actual execution on each supported
    host family before relying on that platform's release evidence.
  related_workitems:
  - BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  decided_at: '2026-08-09T14:18:02+00:00'
---
