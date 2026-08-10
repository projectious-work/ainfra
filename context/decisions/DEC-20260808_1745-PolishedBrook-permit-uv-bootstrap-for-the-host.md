---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260808_1745-PolishedBrook-permit-uv-bootstrap-for-the-host
  created: '2026-08-08T17:45:55+00:00'
spec:
  title: Permit uv bootstrap for the host container gate
  state: accepted
  decision: The Phase 1 host container gate is a PEP 723 script executed by uv. uv-managed
    Python, its isolated script environment, and its pre-entrypoint cache are the
    sole explicit exception to the gate's evidence-local storage rule; all Docker,
    Syft, Grype, scanner database, and post-entrypoint temporary state remains under
    the run evidence directory.
  context: The owner requires uv to manage Python installation and isolation on the
    macOS host. uv necessarily operates before the Python entrypoint can validate
    the run and exclusively create evidence, so its bootstrap state cannot be placed
    under evidence without changing the evidence lifecycle.
  rationale: This provides a reproducible Python 3.11+ runtime with no Python package
    dependencies while keeping container-runtime authority and all security-relevant
    scanner state inside the existing owner-reviewed gate boundary.
  alternatives:
  - option: Keep system Python as the gate entrypoint
    reason: Rejected because the owner requires uv to manage Python installation and
      isolation.
  - option: Create uv state inside evidence before Python starts
    reason: Rejected because it would make evidence pre-exist the gate and contradict
      exclusive evidence creation and reuse protection.
  consequences: The normative release specification and operator documentation describe
    the uv exception. The script uses a fixed PEP 723 shebang with --no-config, validates
    the resolved uv launcher, records its version, and still rejects pre-existing
    evidence. Hosts must install uv in an approved executable directory before direct
    execution.
  related_workitems:
  - BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  decided_at: '2026-08-08T17:45:55+00:00'
---
