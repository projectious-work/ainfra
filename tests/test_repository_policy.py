"""Repository-wide policy checks that must remain local and deterministic."""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_no_github_workflows() -> None:
    workflows = ROOT / ".github" / "workflows"
    assert not workflows.exists(), "GitHub Actions and workflows are prohibited"


def test_mit_license() -> None:
    license_text = (ROOT / "LICENSE").read_text(encoding="utf-8")
    assert license_text.startswith("MIT License\n")
    assert "Copyright (c) 2026 projectious.work" in license_text


def test_fixture_tree_has_no_private_key_or_token_shapes() -> None:
    fixture_root = ROOT / "tests" / "fixtures"
    forbidden = [
        re.compile(r"-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----"),
        re.compile(r"\bgh[opsu]_[A-Za-z0-9]{20,}\b"),
        re.compile(r"\bhcloud_[A-Za-z0-9]{20,}\b", re.IGNORECASE),
    ]
    for path in fixture_root.rglob("*"):
        if not path.is_file():
            continue
        text = path.read_text(encoding="utf-8")
        for pattern in forbidden:
            assert not pattern.search(text), f"{path} matches {pattern.pattern}"
