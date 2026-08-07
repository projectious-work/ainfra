---
apiVersion: processkit.projectious.work/v1
kind: Migration
metadata:
  id: MIG-20260807_1821-ContentSync-processkit-content-sync
  created: 2026-08-07 18:21:07+00:00
  updated: '2026-08-07T18:53:22+00:00'
spec:
  source: processkit
  source_url: https://github.com/projectious-work/processkit.git
  from_version: v0.28.4
  to_version: v0.28.5
  state: applied
  generated_by: aibox apply
  generated_at: 2026-08-07 18:21:07+00:00
  summary: 0 changed upstream, 1 conflicts, 0 new, 0 removed, 0 stale-removed (1 groups
    affected)
  affected_groups:
  - AGENTS
  affected_files:
  - path: AGENTS.md
    classification: conflict
  started_at: '2026-08-07T18:53:22+00:00'
  applied_at: '2026-08-07T18:53:22+00:00'
  progress_notes:
  - timestamp: '2026-08-07T18:53:22+00:00'
    actor: mcp
    note: Resolved AGENTS.md conflict by retaining ainfra project-specific setup/build/test
      commands; upstream generic docs-site command does not apply to this repository.
---

# Migration MIG-20260807_1821-ContentSync-processkit-content-sync

From `v0.28.4` to `v0.28.5` (source: `https://github.com/projectious-work/processkit.git`).

0 changed upstream, 1 conflicts, 0 new, 0 removed, 0 stale-removed (1 groups affected)

## Counts

- unchanged: 721
- changed-locally-only: 0
- changed-upstream-only: 0
- conflict: 1
- new-upstream: 0
- removed-upstream: 0
- removed-upstream-stale: 0

## Changes by group

### AGENTS

**conflict**

- `AGENTS.md` → `AGENTS.md`
