---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260822_1032-DeepFrog-publish-the-phase-9-alpha-release
  created: '2026-08-22T10:32:10+00:00'
spec:
  title: Publish the Phase 9 alpha release
  state: accepted
  decision: Prepare, validate, tag, publish, and verify the next v1.0.0 alpha release
    containing the completed Phase 9 Hetzner Kubernetes baseline work.
  context: Phase 9 live certification passed, including immutable lifecycle, three-node
    K3s health, Cloudflare Service Auth tunnel-only SSH, bastion removal, convergence,
    teardown, and independent resource-leak verification.
  rationale: The user explicitly approved proceeding with the Phase 9 alpha release
    after successful live certification.
  consequences: The release workflow will select the next unused alpha version, isolate
    intended changes from unrelated workspace drift, require clean validation gates,
    publish the tag and GitHub Release, and verify the distribution artifact.
  decided_at: '2026-08-22T10:32:10+00:00'
---
