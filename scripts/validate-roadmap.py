#!/usr/bin/env python3
"""Check roadmap references and identity beyond JSON Schema validation."""

from pathlib import Path
import sys

import yaml


ROOT = Path(__file__).resolve().parents[1]
roadmap_path = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "spec/doc/v1/roadmap.yaml"
roadmap = yaml.safe_load(roadmap_path.read_text())
seen: set[str] = set()
errors: list[str] = []

for group in roadmap["groups"]:
    for phase in group["phases"]:
        phase_id = str(phase["id"])
        if phase_id in seen:
            errors.append(f"duplicate phase ID: {phase_id}")
        seen.add(phase_id)
        note = phase.get("devNote")
        if note and not (ROOT / note).is_file():
            errors.append(f"phase {phase_id}: absent development note {note}")

if errors:
    raise SystemExit("\n".join(errors))
print(f"roadmap references and {len(seen)} phase IDs are valid")
