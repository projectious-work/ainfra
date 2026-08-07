---
name: pk-skill-new
description: Use the skill-builder skill to start the interactive skill-creation workflow for $ARGUMENTS.
---

Use the skill-builder skill to start the interactive skill-creation workflow for $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-skill-new` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
