---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  created: '2026-08-07T18:15:11+00:00'
  labels:
    phase: '1'
    area: release-engineering
    blocked_by: validation-environment
spec:
  title: Complete Phase 1 release-engineering evidence
  state: backlog
  type: task
  priority: high
  description: After aibox exposes Docker, hadolint, syft, grype, cosign, osv-scanner,
    shellcheck, and gitleaks in the ainfra development container, run the fail-closed
    Phase 1 Docker, SBOM, artifact scan, secret scan, OSV, shell, and release checks.
    Obtain native macOS amd64 and arm64 smoke evidence. Re-run independent implementation
    conformance review and keep roadmap Phase 1 in_progress until every required gate
    passes.
---
