---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260810_1014-SkilledEmber-let-uv-manage-the-host-gate
  created: '2026-08-10T10:14:55+00:00'
spec:
  title: Let uv manage the host gate Python runtime
  state: accepted
  decision: The Phase 1 host container gate will trust the reviewed uv executable
    to acquire and select the exactly pinned Python version, create the run-local
    virtual environment, and execute the reviewed Python gate. The launcher will record
    uv and Python identities and hashes as evidence but will not require a separately
    maintained approved-python.sha256 file.
  context: The previous offline bootstrap required operators to manually install Python,
    resolve its executable path, copy its digest into an approval file, and maintain
    a fixed cache layout. This contradicted the accepted objective that uv manage
    Python and the virtual environment and made the one-command host gate unnecessarily
    difficult.
  rationale: Pinning Python exactly while delegating acquisition and environment management
    to uv provides a simple one-command operator workflow. Runtime paths, versions,
    and executable digests remain auditable in the bootstrap manifest. Container inputs
    and release evidence remain controlled, and the gate still performs no commit,
    tag, push, or publication.
  alternatives:
  - option: Retain manual executable-digest approval
    reason: Rejected because it defeats uv-managed setup and creates a brittle manual
      ceremony.
  - option: Use the host's arbitrary Python
    reason: Rejected because it loses the exact runtime pin and consistent environment.
  consequences: The first gate invocation may use network access to acquire the pinned
    Python runtime. Subsequent invocations can reuse uv's cache. The manual Python
    approval file and offline-only uv flags are removed from the launcher and documentation.
  related_workitems:
  - BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  decided_at: '2026-08-10T10:14:55+00:00'
---
