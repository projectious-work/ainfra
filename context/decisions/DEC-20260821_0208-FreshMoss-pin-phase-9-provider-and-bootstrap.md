---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_0208-FreshMoss-pin-phase-9-provider-and-bootstrap
  created: '2026-08-21T02:08:10+00:00'
spec:
  title: Pin Phase 9 provider and bootstrap dependencies
  state: accepted
  decision: Phase 9 pins hetznercloud/hcloud 1.64.0, K3s v1.36.1+k3s1, and cloudflared
    2026.7.2. Runtime binaries require operator-supplied upstream SHA-256 values and
    external secret tokens.
  context: Phase 9 is the first provider-backed production template. The v1 specification
    requires implementation-time selections to be explicit, pinned, security-reviewed,
    and native-tool owned.
  rationale: These are current upstream releases compatible with the supported execution
    boundary. Exact provider locking and explicit runtime checksums prevent mutable
    dependency resolution; external tokens preserve the secret-generation prohibition.
  consequences: Version upgrades require a new disposable lifecycle. Mixed host architectures
    require architecture-specific checksum variables. No new ainfra engine or provider-specific
    CLI schema is introduced.
  decided_at: '2026-08-21T02:08:10+00:00'
---
