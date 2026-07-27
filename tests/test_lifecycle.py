"""Guarded lifecycle orchestration tests use no real infrastructure tools."""

from __future__ import annotations

from collections.abc import Sequence
from pathlib import Path

import pytest

from ainfra import lifecycle as lifecycle_module
from ainfra.contracts import validate_path
from ainfra.errors import GuardError
from ainfra.lifecycle import Lifecycle, _backend_args
from ainfra.runner import Result

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
    assert "-lockfile=readonly" in runner.calls[0][0]
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


def test_destroy_requires_exact_plan_id(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    lifecycle = Lifecycle(FakeRunner())
    with pytest.raises(GuardError):
        lifecycle.destroy(
            "hetzner-kubernetes-baseline",
            INPUT,
            "destroy-wrong",
        )


def test_destroy_runs_only_with_exact_plan_id(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    runner = FakeRunner()
    record = Lifecycle(runner).plan(
        "hetzner-kubernetes-baseline",
        INPUT,
        destroy=True,
    )
    runner.calls.clear()
    result = Lifecycle(runner).destroy(
        "hetzner-kubernetes-baseline",
        INPUT,
        record.id,
    )
    assert result.returncode == 0
    assert runner.calls[-1][0] == (
        "tofu",
        "apply",
        "-input=false",
        record.plan_path,
    )


def test_plan_ids_are_unique(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    lifecycle = Lifecycle(FakeRunner())
    first = lifecycle.plan("hetzner-kubernetes-baseline", INPUT)
    second = lifecycle.plan("hetzner-kubernetes-baseline", INPUT)
    assert first.id != second.id
    assert first.plan_path != second.plan_path


def test_apply_rejects_tampered_plan(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    lifecycle = Lifecycle(FakeRunner())
    record = lifecycle.plan("hetzner-kubernetes-baseline", INPUT)
    Path(record.plan_path).write_bytes(b"replaced")
    with pytest.raises(GuardError, match="modified"):
        lifecycle.apply(
            "hetzner-kubernetes-baseline",
            INPUT,
            record.id,
        )


def test_apply_plan_cannot_authorize_destroy(monkeypatch: object) -> None:
    monkeypatch.setenv("HCLOUD_TOKEN", "fixture-secret")  # type: ignore[attr-defined]
    lifecycle = Lifecycle(FakeRunner())
    record = lifecycle.plan("hetzner-kubernetes-baseline", INPUT)
    with pytest.raises(GuardError, match="cannot authorize"):
        lifecycle.destroy(
            "hetzner-kubernetes-baseline",
            INPUT,
            record.id,
        )


def test_remote_backend_is_passed_to_init(
    tmp_path: Path,
    monkeypatch: object,
) -> None:
    root = tmp_path / "repo"
    backend = root / ".ainfra" / "backend.hcl"
    backend.parent.mkdir(parents=True)
    backend.write_text('bucket = "fixture"\n', encoding="utf-8")
    monkeypatch.setattr(  # type: ignore[attr-defined]
        lifecycle_module,
        "repository_root",
        lambda: root,
    )
    remote_input = (
        ROOT
        / "tests"
        / "fixtures"
        / "policy"
        / "v1alpha1"
        / "valid"
        / "remote-input.json"
    )
    document = validate_path(remote_input)
    assert _backend_args(document) == [f"-backend-config={backend}"]
