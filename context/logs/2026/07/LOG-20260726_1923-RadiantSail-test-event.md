---
apiVersion: processkit.projectious.work/v2
kind: LogEntry
metadata:
  id: LOG-20260726_1923-RadiantSail-test-event
  created: '2026-07-26T19:23:17+00:00'
spec:
  event_type: test.event
  timestamp: '2026-07-26T19:23:17+00:00'
  summary: Third disposable Hetzner test established trusted SSH but stopped at cloud-init
    schema validation and was fully destroyed.
  actor: codex
  subject: BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  subject_kind: WorkItem
  details:
    result: failed-safely
    passed:
    - deterministic host key active
    - strict SSH verification
    - exact six-resource destroy
    failed: cloud-init schema rejected empty ssh_genkeytypes
    resources_created: 6
    resources_destroyed: 6
    correction: 'Use ssh_genkeytypes: [ed25519]; local cloud-init schema validation
      now passes.'
    next_plan: fourth-test.tfplan
---
