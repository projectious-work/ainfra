---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260810_1914-SmartHarbor-implement-container-side-release-packaging
  created: '2026-08-10T19:14:06+00:00'
  updated: '2026-08-10T19:18:32+00:00'
spec:
  title: Implement container-side release packaging
  state: review
  type: task
  priority: high
  description: Create a deterministic release packaging command that runs in the development
    container, builds Linux/macOS amd64/arm64 archives, generates SBOMs and a complete
    SHA-256 manifest under dist/release/<version>, validates inputs, and hands the
    result to the host release-sign command.
  started_at: '2026-08-10T19:14:09+00:00'
---

## Transition note (2026-08-10T19:14:09+00:00)

Implementation started.


## Transition note (2026-08-10T19:18:32+00:00)

Implemented release-package command. Isolated clean-repository integration built four archives and SPDX SBOMs, verified all checksums, and ran the Linux arm64 binary with the injected release version. Full validation and tests pass; awaiting review/commit before authoritative packaging.
