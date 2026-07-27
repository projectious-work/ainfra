---
apiVersion: processkit.projectious.work/v2
kind: LogEntry
metadata:
  id: LOG-20260726_1934-CoolRose-test-event
  created: '2026-07-26T19:34:21+00:00'
spec:
  event_type: test.event
  timestamp: '2026-07-26T19:34:21+00:00'
  summary: Fifth Hetzner test exposed the analogous auditd pristine-host check-mode
    defect and was fully destroyed.
  actor: codex
  subject: BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  subject_kind: WorkItem
  details:
    result: failed-safely
    passed:
    - strict SSH
    - cloud-init schema
    - Ansible through observability role
    failed: auditd service management ran in check mode before the simulated package
      install existed.
    correction: Skip auditd service management only during check mode; full service
      sweep completed.
    local_validation:
    - 6 policy tests passed
    - ansible-lint passed
    - Ansible syntax passed
    resources_created: 6
    resources_destroyed: 6
    remaining_planned_runs: 1
---
