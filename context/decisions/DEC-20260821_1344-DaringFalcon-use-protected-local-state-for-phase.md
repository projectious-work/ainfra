---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_1344-DaringFalcon-use-protected-local-state-for-phase
  created: '2026-08-21T13:44:51+00:00'
spec:
  title: Use protected local state for Phase 9 live certification
  state: accepted
  decision: Use an explicit local OpenTofu backend path in a mode-0700 disposable
    certification directory shared by all reviewed ainfra runs. Do not require Cloudflare
    R2 or another remote backend for the Phase 9 alpha certification.
  context: The certification is a short-lived, single-operator disposable lifecycle.
    OpenTofu state must persist across isolated plan, apply, bastion-removal, replan,
    and destroy runs, but the normative specification does not require a remote backend.
  rationale: An explicit absolute local backend path provides the required multi-run
    continuity without introducing unrelated storage permissions or cost. The state
    contains provider resource data but none of the externally supplied K3s, Cloudflare,
    or SSH private secrets.
  alternatives:
  - option: Cloudflare R2 S3 backend
    reason: Not selected because it adds bucket credentials and setup unrelated to
      this disposable certification.
  - option: Per-run default local state
    reason: Invalid because isolated ainfra workspaces would not share infrastructure
      state.
  consequences: The run must use umask 077, a mode-0700 directory, retained local
    state backups, unique ownership labels, and emergency cleanup. State is removed
    only after independent Hetzner teardown verification. Certification evidence must
    disclose that remote-backend recovery was not tested; production operators may
    still select a capability-tested remote backend.
  related_workitems:
  - BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
  decided_at: '2026-08-21T13:44:51+00:00'
---
