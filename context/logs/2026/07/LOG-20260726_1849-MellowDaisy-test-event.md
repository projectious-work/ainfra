---
apiVersion: processkit.projectious.work/v2
kind: LogEntry
metadata:
  id: LOG-20260726_1849-MellowDaisy-test-event
  created: '2026-07-26T18:49:26+00:00'
spec:
  event_type: test.event
  timestamp: '2026-07-26T18:49:26+00:00'
  summary: Disposable Hetzner infrastructure test applied and fully destroyed; Ansible
    blocked at trusted SSH host-key gate.
  actor: codex
  subject: BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  subject_kind: WorkItem
  details:
    environment: first-test-20260726
    resources_created: 6
    resources_destroyed: 6
    api_inventory_after_destroy:
      servers: 0
      networks: 0
      firewalls: 0
      ssh_keys: 0
    result: partial
    passed:
    - saved plan applied
    - ownership labels
    - CX23 type
    - Debian 13 image
    - public IPv4 present
    - public IPv6 absent
    - exact destroy plan generated
    - all resources removed
    blocked: No deterministic SSH host key was provisioned and unattended ssh-keyscan
      is prohibited.
    next_action: Provision a deterministic host key or automate trusted-console fingerprint
      retrieval, then repeat Ansible check/apply/idempotence.
---
