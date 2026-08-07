---
name: pk-release-audit
description: Use the release-audit skill to run the pre-release validation sweep and print
---

Use the release-audit skill to run the pre-release validation sweep and print
a single report. Detect-only; never modifies any file. Exit 0 if no ERRORs
found; exit 1 otherwise.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-release-audit` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
