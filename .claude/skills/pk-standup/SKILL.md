---
name: pk-standup
description: "Generate today's standup: what was done, what is in progress, what comes"
---

Generate today's standup: what was done, what is in progress, what comes
next, and any blockers. Quick, structured, for the team record.

When in a git repository with a GitHub remote, check open GitHub issues
and pull requests via `gh` when available/authenticated and include
issue/PR blockers or review-needed PRs.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-standup` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
