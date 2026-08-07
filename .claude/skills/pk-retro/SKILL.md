---
name: pk-retro
description: Use the retrospective skill to generate a post-release blameless
---

Use the retrospective skill to generate a post-release blameless
retrospective for $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-retro` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
