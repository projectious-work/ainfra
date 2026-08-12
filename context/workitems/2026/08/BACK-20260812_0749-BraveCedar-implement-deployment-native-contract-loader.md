---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260812_0749-BraveCedar-implement-deployment-native-contract-loader
  created: '2026-08-12T07:49:15+00:00'
  updated: '2026-08-12T10:06:26+00:00'
spec:
  title: Implement deployment and native-file contract loader
  state: in-progress
  type: story
  priority: high
  description: 'Strictly parse and validate ainfra-owned deployment/config/lock documents,
    reject unknown fields, enforce contained regular paths, and preserve native input
    bytes and ordering. Acceptance: malformed, traversal, symlink, duplicate, permission,
    and opaque-byte fixtures.'
  parent: BACK-20260812_0749-SparklingCliff-deliver-phase-two-contracts-doctor
  scope: Phase 2
  started_at: '2026-08-12T10:06:26+00:00'
---

## Transition note (2026-08-12T10:06:26+00:00)

Strict loader is now exposed through doctor deployment with schema-valid CLI and black-box evidence.
