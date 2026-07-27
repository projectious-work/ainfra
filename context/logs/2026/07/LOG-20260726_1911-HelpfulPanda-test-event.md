---
apiVersion: processkit.projectious.work/v2
kind: LogEntry
metadata:
  id: LOG-20260726_1911-HelpfulPanda-test-event
  created: '2026-07-26T19:11:54+00:00'
spec:
  event_type: test.event
  timestamp: '2026-07-26T19:11:54+00:00'
  summary: Second disposable Hetzner test destroyed safely after deterministic SSH
    host key failed to activate.
  actor: codex
  subject: BACK-20260724_2005-CoolLynx-integrate-security-live-verification
  subject_kind: WorkItem
  details:
    environment: first-test-20260726
    result: failed-safely
    resources_created: 6
    resources_destroyed: 6
    expected_host_key_fingerprint: SHA256:IdCEWASqky2+pmsdgep89T7ezzvUbJjPSgZVwKSzBDw
    observed_host_key_fingerprint: SHA256:nUCWYRf1kYT2eiaY/clNjPRvgpD43BCztyvFork2pik
    ansible_run: false
    cleanup_verified:
      servers: 0
      networks: 0
      firewalls: 0
      ssh_keys: 0
      tofu_state: empty
    next_action: Replace late cloud-init host-key overwrite with a boot-stage mechanism
      or trusted console fingerprint retrieval, validate locally, then request a new
      cost approval.
---
