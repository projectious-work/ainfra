"""CLI boundary and stable-exit tests."""

from __future__ import annotations

from pathlib import Path

from ainfra.cli import main

ROOT = Path(__file__).resolve().parents[1]
FIXTURES = ROOT / "tests" / "fixtures" / "contracts" / "v1alpha1"


def test_validate_text(capsys: object) -> None:
    path = FIXTURES / "valid" / "template-input.json"
    assert main(["validate", str(path)]) == 0
    captured = capsys.readouterr()  # type: ignore[attr-defined]
    assert "valid TemplateInput" in captured.out
    assert captured.err == ""


def test_validate_json(capsys: object) -> None:
    path = FIXTURES / "valid" / "template-output.json"
    assert main(["validate", str(path), "--format", "json"]) == 0
    captured = capsys.readouterr()  # type: ignore[attr-defined]
    assert '"ok": true' in captured.out
    assert captured.err == ""


def test_invalid_contract_has_stable_exit(capsys: object) -> None:
    path = FIXTURES / "invalid" / "unsupported-version.json"
    assert main(["validate", str(path)]) == 3
    captured = capsys.readouterr()  # type: ignore[attr-defined]
    assert captured.out == ""
    assert "AINFRA-E200" in captured.err


def test_unimplemented_lifecycle_is_guarded(capsys: object) -> None:
    assert main(["apply", "hetzner-kubernetes-baseline"]) == 6
    captured = capsys.readouterr()  # type: ignore[attr-defined]
    assert "AINFRA-E600" in captured.err
