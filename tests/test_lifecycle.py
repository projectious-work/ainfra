"""Guarded lifecycle orchestration tests use no real infrastructure tools."""

from __future__ import annotations

from collections.abc import Sequence
from pathlib import Path

import pytest

from ainfra.errors import GuardError
from ainfra.lifecycle import Lifecycle, destruction_token
from ainfra.runner import Result
from ainfra.template import discover_template

ROOT = Path(__file__).resolve().parents[1]
INPUT = (
    ROOT
    / "tests"
    / "fixtures"
    / "contracts"
    / "v1alpha1"
    / "valid"
    / "template-input.json"
)


class FakeRunner:
    def __init__(self) -> None:
        self.calls: list[tuple[tuple[str, ...], Path, dict[str, str]]] = []

    def run(
        self,
        argv: Sequence[str],
        *,
        cwd: Path,
        env: dict[str, str],
        secrets: Sequence[str] = (),
    ) -> Result:
        arguments = tuple(argv)
        self.calls.append((arguments, cwd, env))
        for value in arguments:
            if value.startswith("-out="):
                path = Path(value.removeprefix("-out="))
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(b"fake-plan")
        return Result(arguments, 0, "ok\n", "")


def test_plan_uses_explicit_argv_and_records_binding(
    monkeypatch: object,
) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    runner = FakeRunner()
    record = Lifecycle(runner).plan("hetzner-kubernetes-baseline", INPUT)
    assert len(runner.calls) == 2
    assert runner.calls[0][0][:2] == ("tofu", "init")
    assert runner.calls[1][0][:2] == ("tofu", "plan")
    assert Path(record.plan_path).is_file()
    assert "fixture-secret" not in " ".join(runner.calls[1][0])


def test_apply_requires_exact_plan_binding(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    with pytest.raises(GuardError):
        Lifecycle(FakeRunner()).apply(
            "hetzner-kubernetes-baseline",
            INPUT,
            "not-a-plan",
        )


def test_destroy_requires_exact_scope(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    lifecycle = Lifecycle(FakeRunner())
    with pytest.raises(GuardError):
        lifecycle.destroy(
            "hetzner-kubernetes-baseline",
            INPUT,
            "destroy-wrong",
        )


def test_destroy_runs_only_with_scope_token(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    template = discover_template("hetzner-kubernetes-baseline")
    from ainfra.contracts import validate_path

    document = validate_path(INPUT)
    token = destruction_token(template, document, INPUT)
    runner = FakeRunner()
    result = Lifecycle(runner).destroy(
        template.name,
        INPUT,
        token,
    )
    assert result.returncode == 0
    assert runner.calls[0][0] == (
        "tofu",
        "destroy",
        "-input=false",
        "-auto-approve",
    )
