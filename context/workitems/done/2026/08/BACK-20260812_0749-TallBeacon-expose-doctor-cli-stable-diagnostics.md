---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-TallBeacon-expose-doctor-cli-stable-diagnostics
  created: '2026-08-12T07:49:16+00:00'
  updated: '2026-08-12T14:10:46+00:00'
spec:
  title: Expose doctor CLI and stable diagnostics
  state: done
  type: story
  priority: high
  description: Implement doctor scopes, bare-doctor aliasing, complete text/JSON reports,
    stdout/stderr and exit contracts, failure-first text, and schema/golden coverage.
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
  started_at: '2026-08-12T10:44:15+00:00'
  completed_at: '2026-08-12T14:10:46+00:00'
---

## Transition note (2026-08-12T10:44:15+00:00)

Wired doctor run, doctor all, and the exact bare-doctor alias with schema-valid aggregate evidence.


## Transition note (2026-08-12T14:10:39+00:00)

Implementation and acceptance evidence complete on feat/phase-02-contracts-doctor; authoritative validation, security, race, schema, black-box, and relevant fuzz checks pass.


## Transition note (2026-08-12T14:10:46+00:00)

Accepted on the recorded Phase 2 conformance evidence and clean processkit release audit.
