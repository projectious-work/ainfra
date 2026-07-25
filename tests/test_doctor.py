"""Doctor reporting is read-only and machine-readable."""

from __future__ import annotations

from ainfra import doctor


def test_doctor_reports_missing_tool(
    monkeypatch: object,
) -> None:
    original = doctor.shutil.which

    def fake_which(name: str) -> str | None:
        if name == "tofu":
            return None
        return original(name)

    monkeypatch.setattr(doctor.shutil, "which", fake_which)  # type: ignore[attr-defined]
    checks = doctor.run_doctor()
    tofu = next(check for check in checks if check.id == "tofu")
    assert tofu.status == "fail"
    assert tofu.found is None


def test_serialized_doctor_has_stable_fields() -> None:
    checks = doctor.serialized_checks()
    assert checks
    assert set(checks[0]) == {
        "id",
        "status",
        "found",
        "required",
        "remediation",
    }
