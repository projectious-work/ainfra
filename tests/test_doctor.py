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


def test_exact_version_policy() -> None:
    assert doctor._version_matches("8.30.1", "8.30.1", "exact")
    assert not doctor._version_matches("8.30.2", "8.30.1", "exact")


def test_minimum_version_policy() -> None:
    assert doctor._version_matches("1.11.0", "1.10.0", "minimum")
    assert not doctor._version_matches("1.9.9", "1.10.0", "minimum")


def test_doctor_checks_selected_backend_input() -> None:
    path = (
        doctor.repository_root()
        / "tests"
        / "fixtures"
        / "policy"
        / "v1alpha1"
        / "valid"
        / "remote-input.json"
    )
    checks = doctor.run_doctor(path)
    backend = next(check for check in checks if check.id == "backend-policy")
    assert backend.status == "pass"
    assert backend.found == str(path)
