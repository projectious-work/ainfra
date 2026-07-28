"""Repository-wide policy checks that must remain local and deterministic."""

from __future__ import annotations

import re
import tomllib
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


def test_python_is_development_tooling_not_a_product_runtime() -> None:
    assert not any((ROOT / "src" / "ainfra").glob("*.py"))
    config = tomllib.loads(
        (ROOT / "pyproject.toml").read_text(encoding="utf-8")
    )
    assert config["project"]["dependencies"] == []
    assert "scripts" not in config["project"]
    assert config["tool"]["uv"]["package"] is False


def test_user_documentation_has_no_python_cli_runtime() -> None:
    excluded = {
        ROOT / "docs/content/docs/contributing/_index.md",
        ROOT / "docs/content/docs/guides/local-tooling.md",
    }
    for path in (ROOT / "docs/content/docs").rglob("*.md"):
        if path in excluded:
            continue
        text = path.read_text(encoding="utf-8")
        assert "uv run ainfra" not in text
        assert "Python 3.12" not in text
