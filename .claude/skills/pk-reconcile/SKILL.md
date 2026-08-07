---
name: pk-reconcile
description: "Use the project-reconciliation skill to reconcile the requested migration,"
---

Use the project-reconciliation skill to reconcile the requested migration,
health, and repository queues with guarded remediation.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-reconcile` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
