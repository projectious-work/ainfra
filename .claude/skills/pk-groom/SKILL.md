---
name: pk-groom
description: Use the context-grooming skill to prune and compact the current project context.
---

Use the context-grooming skill to prune and compact the current project context.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-groom` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
