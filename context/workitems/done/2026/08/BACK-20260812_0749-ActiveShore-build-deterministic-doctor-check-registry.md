---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-ActiveShore-build-deterministic-doctor-check-registry
  created: '2026-08-12T07:49:16+00:00'
  updated: '2026-08-12T14:10:46+00:00'
spec:
  title: Build deterministic doctor check registry
  state: done
  type: story
  priority: high
  description: Define checks with stable ID, scope, prerequisites, capabilities, applicability,
    status, severity, remediation, and deterministic ordering for environment, deployment,
    template, run, and all scopes.
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
  started_at: '2026-08-12T09:48:56+00:00'
  completed_at: '2026-08-12T14:10:46+00:00'
---

## Transition note (2026-08-12T09:48:56+00:00)

Capability-aware registry and initial environment platform/executable checks are now implemented.


## Transition note (2026-08-12T14:10:39+00:00)

Implementation and acceptance evidence complete on feat/phase-02-contracts-doctor; authoritative validation, security, race, schema, black-box, and relevant fuzz checks pass.


## Transition note (2026-08-12T14:10:46+00:00)

Accepted on the recorded Phase 2 conformance evidence and clean processkit release audit.
