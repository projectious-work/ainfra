---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_1717-PatientHorizon-automate-provider-anchored-ssh-host-key
  created: '2026-08-21T17:17:39+00:00'
spec:
  title: Automate provider-anchored SSH host-key attestation
  state: accepted
  decision: Future live-certification phases must replace manual browser-console fingerprint
    transcription with an auditable, provider-anchored host-key attestation workflow
    that can populate strict known_hosts inputs without trusting unauthenticated network
    scans.
  context: Phase 9 required manually opening the Hetzner browser console, resetting
    or entering credentials, and typing a fingerprint command. The operator confirmed
    the bastion fingerprint but found the ceremony unacceptably burdensome.
  rationale: Automated attestation should preserve the independent trust boundary
    while reducing transcription error, operator effort, and billable certification
    delay.
  consequences: Phase 9 may continue from the manually confirmed bastion trust anchor.
    A follow-up design must evaluate provider console or metadata attestations, pre-provisioned
    host certificates, ephemeral SSH CAs, and evidence retention without placing reusable
    private host credentials in state.
  related_workitems:
  - BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
  decided_at: '2026-08-21T17:17:39+00:00'
---
