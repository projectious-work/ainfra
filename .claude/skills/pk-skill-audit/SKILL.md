---
name: pk-skill-audit
description: Use the skill-reviewer skill to run a full 11-category audit of the $ARGUMENTS skill.
---

Use the skill-reviewer skill to run a full 11-category audit of the $ARGUMENTS skill.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-skill-audit` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
