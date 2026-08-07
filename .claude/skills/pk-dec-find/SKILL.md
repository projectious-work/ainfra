---
name: pk-dec-find
description: Use the decision-record skill to query existing decisions matching $ARGUMENTS.
---

Use the decision-record skill to query existing decisions matching $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-dec-find` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
