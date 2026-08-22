---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260821_1717-FirmLeaf-design-automated-provider-anchored-ssh-host
  created: '2026-08-21T17:17:40+00:00'
  labels:
    phase: post-phase-9
    security: ssh-host-attestation
spec:
  title: Design automated provider-anchored SSH host-key attestation
  state: backlog
  type: story
  priority: high
  description: Replace manual web-console fingerprint transcription in future live
    certifications with an auditable automated trust bootstrap that preserves strict
    host-key checking and does not trust unauthenticated ssh-keyscan output or persist
    reusable host private keys in infrastructure state.
  parent: BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
---
