---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-CuriousBloom-validate-local-template-contracts
  created: '2026-08-12T07:49:15+00:00'
  updated: '2026-08-12T14:10:46+00:00'
spec:
  title: Validate local resolved template contracts
  state: done
  type: story
  priority: high
  description: 'Validate only already-local resolved template layout and manifest
    contracts, compatibility, native ownership rules, and prohibited runtime/state
    content. Do not acquire sources. Acceptance: positive and hostile fixture suites
    with capability-aware non-mutating child checks.'
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
  started_at: '2026-08-12T10:18:09+00:00'
  completed_at: '2026-08-12T14:10:46+00:00'
---

## Transition note (2026-08-12T10:18:09+00:00)

Strict local template loader and doctor template CLI are implemented without source acquisition.


## Transition note (2026-08-12T14:10:39+00:00)

Implementation and acceptance evidence complete on feat/phase-02-contracts-doctor; authoritative validation, security, race, schema, black-box, and relevant fuzz checks pass.


## Transition note (2026-08-12T14:10:46+00:00)

Accepted on the recorded Phase 2 conformance evidence and clean processkit release audit.
