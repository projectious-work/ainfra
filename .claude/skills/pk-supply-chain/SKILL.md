---
name: pk-supply-chain
description: Run the supply-chain-audit skill.
---

Run the supply-chain-audit skill.

Default behavior:

- Discover manifests under `--root` (or the project root).
- Run local manifest checks only.
- Keep security and quality probes off unless `--run-security`
  and/or `--run-quality` are passed.


---

This command is a processkit skill shim. Load and follow the matching skill for `pk-supply-chain` from `context/skills/` instead of executing underlying helper scripts directly. Do not run `context/skills/**/scripts/*.py`, `doctor.py`, or `uv run .../scripts/...` unless the skill instructions explicitly require that implementation detail for the current step.
