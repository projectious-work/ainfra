---
name: pk-work
description: "Use the workitem-management skill to create a new work item: $ARGUMENTS."
---

Use the workitem-management skill to create a new work item: $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-work` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
