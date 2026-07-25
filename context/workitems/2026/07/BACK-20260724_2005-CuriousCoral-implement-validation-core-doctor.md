---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260724_2005-CuriousCoral-implement-validation-core-doctor
  created: '2026-07-24T20:05:10+00:00'
  updated: '2026-07-25T09:05:52+00:00'
spec:
  title: Implement contract validation core and validate/doctor commands
  state: in-progress
  type: story
  priority: high
  description: Milestone 1 / PR 2. Implement template discovery confinement, strict
    schema and version validation, secret-reference rules, environment and backend
    safety policies, image choices, CIDR and SSH invariants, output sanitization,
    and actionable validate/doctor diagnostics with negative tests.
  parent: BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  started_at: '2026-07-25T09:05:52+00:00'
---

## Transition note (2026-07-25T09:05:52+00:00)

Starting semantic policy validation, template discovery, and doctor implementation.
