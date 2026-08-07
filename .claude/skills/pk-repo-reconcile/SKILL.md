---
name: pk-repo-reconcile
description: Use the repo-management skill to plan and apply guarded repository
---

Use the repo-management skill to plan and apply guarded repository
reconciliation for $ARGUMENTS: inspect provider state, open issues,
open change requests, local changes, commits, and pushes.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-repo-reconcile` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
