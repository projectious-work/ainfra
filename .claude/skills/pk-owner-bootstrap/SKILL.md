---
name: pk-owner-bootstrap
description: Use the owner-profiling skill to run the initial profiling interview for $ARGUMENTS.
---

Use the owner-profiling skill to run the initial profiling interview for $ARGUMENTS.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-owner-bootstrap` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
