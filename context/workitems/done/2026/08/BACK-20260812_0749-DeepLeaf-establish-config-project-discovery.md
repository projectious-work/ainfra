---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-DeepLeaf-establish-config-project-discovery
  created: '2026-08-12T07:49:15+00:00'
  updated: '2026-08-12T14:10:45+00:00'
spec:
  title: Establish effective configuration and project discovery
  state: done
  type: story
  priority: high
  description: 'Implement deterministic target precedence, canonical directory/manifest
    equivalence, typed effective configuration, and redacted provenance. Acceptance:
    unit and black-box precedence, ambiguity, containment, and no-descendant-search
    tests.'
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
  started_at: '2026-08-12T07:49:27+00:00'
  completed_at: '2026-08-12T14:10:45+00:00'
---

## Transition note (2026-08-12T07:49:27+00:00)

First dependency slice: target discovery and effective configuration foundations.


## Transition note (2026-08-12T14:10:39+00:00)

Implementation and acceptance evidence complete on feat/phase-02-contracts-doctor; authoritative validation, security, race, schema, black-box, and relevant fuzz checks pass.


## Transition note (2026-08-12T14:10:45+00:00)

Accepted on the recorded Phase 2 conformance evidence and clean processkit release audit.
