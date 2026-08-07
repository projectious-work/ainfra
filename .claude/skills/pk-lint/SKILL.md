---
name: pk-lint
description: Read the `<!-- pk-commands BEGIN -->` ... `<!-- pk-commands END -->` block
---

Read the `<!-- pk-commands BEGIN -->` ... `<!-- pk-commands END -->` block
from the project's AGENTS.md. Extract the value for the `lint` key.

If the value is a non-empty string, run it via Bash. Report the result:
pass/fail status and a brief summary of the output (last 20 lines on
failure, one-line confirmation on success).

If the value is an empty string `""`, tell the user: "No lint command
configured for this project. Add it to the pk-commands block in AGENTS.md."


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-lint` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
