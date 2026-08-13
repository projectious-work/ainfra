---
name: pk-release
description: "Use the release-semver skill to prepare a $ARGUMENTS release: bump the"
---

Use the release-semver skill to prepare a $ARGUMENTS release: bump the
version, curate and verify the matching CHANGELOG.md entry, create the tag
on the designated release-integration branch, publish, and verify.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-release` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
