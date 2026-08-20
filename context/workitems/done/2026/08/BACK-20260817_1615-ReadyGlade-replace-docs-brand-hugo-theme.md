---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260817_1615-ReadyGlade-replace-docs-brand-hugo-theme
  created: '2026-08-17T16:15:49+00:00'
  updated: '2026-08-17T18:42:28+00:00'
spec:
  title: Replace documentation site with projectious.work Hugo theme
  state: done
  type: epic
  priority: high
  description: Replace the Hextra/Docsy-derived Hugo site with brand-theme-hugo-vanilla
    v0.3.3, remove legacy dependencies and DOM/CSS coupling, migrate Blog to the native
    Change log, preserve and redesign the YAML-generated roadmap in compact inverse
    order, validate the live site on port 1314, and report confirmed missing theme
    idioms upstream.
  scope: Documentation
  started_at: '2026-08-17T16:15:56+00:00'
  completed_at: '2026-08-17T18:42:28+00:00'
---

## Transition note (2026-08-17T16:15:56+00:00)

Upstream v0.3.3 release confirmed; implementation begins with Hugo module/config migration and native theme surfaces.


## Transition note (2026-08-17T16:52:48+00:00)

Theme v0.3.3 migration, legacy dependency removal, native Change log, inverse compact YAML roadmap, build/test validation, and upstream issue #51 are complete. Awaiting owner visual verification at port 1314.


## Transition note (2026-08-17T18:42:28+00:00)

Replaced Hextra with brand-theme-hugo-vanilla v0.3.3, migrated content and roadmap, removed legacy dependencies, documented and reported upstream gaps in issue #51, passed full repository validation and 10 desktop/mobile browser checks, and verified the watcher on port 1314.
