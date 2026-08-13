---
name: pk-publish
description: Use the release-semver skill to execute the publish phase of the
---

Use the release-semver skill to execute the publish phase of the
prepared release.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-publish` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
