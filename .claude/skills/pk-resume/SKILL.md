---
name: pk-resume
description: "Use the status-briefing skill to run `pk-reconcile session-start`, then"
---

Use the status-briefing skill to run `pk-reconcile session-start`, then
generate a session-start orientation and catch-up summary from its results.
Use `report-only` reconciliation when the user asks for info-only, dry-run,
report-only, check-only, or no-fix behaviour.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-resume` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
