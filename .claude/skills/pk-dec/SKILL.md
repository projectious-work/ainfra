---
name: pk-dec
description: Use the decision-record skill to record a new decision titled $ARGUMENTS.
---

Use the decision-record skill to record a new decision titled $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-dec` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
