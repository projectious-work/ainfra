---
name: pk-status
description: "Use the standup-context skill to generate a status update: current"
---

Use the standup-context skill to generate a status update: current
progress, blockers, and recommended next steps.

When in a git repository with a GitHub remote, include open GitHub
issues and pull requests via `gh` when available/authenticated. Call out
review-needed PRs, stale/high-priority issues, and external blockers;
note briefly if GitHub could not be checked.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-status` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
