"""Language-neutral CLI compatibility cases for the Python oracle."""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path
from typing import Any

import yaml  # type: ignore[import-untyped]

ROOT = Path(__file__).resolve().parents[1]
CASES = ROOT / "tests" / "compat" / "cli" / "cases.yaml"


def _cases() -> list[dict[str, Any]]:
    loaded = yaml.safe_load(CASES.read_text(encoding="utf-8"))
    assert isinstance(loaded, list)
    return loaded


def test_python_cli_matches_compatibility_cases() -> None:
    environment = {
        key: value
        for key, value in os.environ.items()
        if key in {"HOME", "LANG", "LC_ALL", "PATH", "PYTHONPATH", "TMPDIR"}
    }
    for case in _cases():
        arguments = [
            str(value).replace("<ROOT>", str(ROOT)) for value in case["args"]
        ]
        result = subprocess.run(
            [sys.executable, "-m", "ainfra", *arguments],
            cwd=ROOT,
            env=environment,
            check=False,
            capture_output=True,
            text=True,
            timeout=30,
        )
        expected = case["expect"]
        assert result.returncode == expected["exit"], case["name"]
        for fragment in expected["stdout_contains"]:
            assert fragment in result.stdout, case["name"]
        for fragment in expected["stderr_contains"]:
            assert fragment in result.stderr, case["name"]
