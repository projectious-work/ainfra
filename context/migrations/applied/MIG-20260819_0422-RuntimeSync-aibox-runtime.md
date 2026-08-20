---
apiVersion: processkit.projectious.work/v1
kind: Migration
metadata:
  id: MIG-20260819_0422-RuntimeSync-aibox-runtime
  created: 2026-08-19 04:22:27+00:00
  updated: '2026-08-20T07:45:34+00:00'
spec:
  source: aibox-runtime-home
  source_url: aibox://runtime-home
  from_version: 0.32.6
  to_version: 0.33.2
  state: applied
  generated_by: aibox apply
  generated_at: 2026-08-19 04:22:27+00:00
  summary: 0 changed upstream, 0 conflicts, 1 new, 0 removed (1 groups affected)
  affected_groups:
  - runtime-misc
  affected_files:
  - path: .local/bin/aibox-agent-signal
    classification: new-upstream
  started_at: '2026-08-20T07:45:34+00:00'
  applied_at: '2026-08-20T07:45:34+00:00'
  progress_notes:
  - timestamp: '2026-08-20T07:45:34+00:00'
    actor: mcp
    note: Applied after explicit user approval during project reconciliation.
---

# Migration MIG-20260819_0422-RuntimeSync-aibox-runtime

Managed `.aibox-home/` runtime changes from `0.32.6` to `0.33.2`.

0 changed upstream, 0 conflicts, 1 new, 0 removed (1 groups affected)

## Counts

- unchanged: 43
- changed-locally-only: 0
- changed-upstream-only: 0
- conflict: 0
- new-upstream: 1
- removed-upstream: 0

- removed-upstream-stale: 0

## Changes by group

### runtime-misc

**new-upstream**

- `.aibox-home/.local/bin/aibox-agent-signal` -> `.aibox-home/.local/bin/aibox-agent-signal`
