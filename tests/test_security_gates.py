import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_security_gates_are_strict_local_commands() -> None:
    script = (ROOT / "scripts" / "security-all").read_text(encoding="utf-8")
    assert "checkov_version" in script
    assert "3.2.529" in script
    assert "--directory templates" in script
    assert "--framework terraform" in script
    assert "--skip-download" in script
    assert "--download-external-modules false" in script
    assert "gitleaks dir --redact --no-banner --exit-code 1 ." in script
    assert "--soft-fail" not in script
    assert "|| true" not in script
    assert "gitleaks detect" not in script
    assert "scripts/security-smoke" in script


def test_no_github_workflows_exist() -> None:
    assert not (ROOT / ".github" / "workflows").exists()


def test_security_tools_are_exactly_pinned() -> None:
    tools = tomllib.loads((ROOT / "tools.lock").read_text(encoding="utf-8"))[
        "tools"
    ]
    assert tools["checkov"] == {"version": "3.2.529", "policy": "exact"}
    assert tools["gitleaks"] == {"version": "8.30.1", "policy": "exact"}


def test_gitleaks_config_has_no_baseline_or_rule_disable() -> None:
    config = (ROOT / ".gitleaks.toml").read_text(encoding="utf-8")
    assert "useDefault = true" in config
    assert "baseline" not in config.lower()
    assert "disabledRule" not in config
    assert "context/templates/" in config


def test_security_smoke_asserts_specific_findings() -> None:
    smoke = (ROOT / "scripts" / "security-smoke").read_text(encoding="utf-8")
    assert "--check CKV_AWS_24" in smoke
    assert "grep -q 'CKV_AWS_24'" in smoke
    assert "--report-format json" in smoke
    assert '"RuleID": "generic-api-key"' in smoke


def test_bootstrap_is_exact_and_checksum_verified() -> None:
    bootstrap = (ROOT / "scripts" / "bootstrap-security-tools").read_text(
        encoding="utf-8"
    )
    assert "checkov==3.2.529" in bootstrap
    assert "gitleaks_8.30.1_checksums.txt" in bootstrap
    assert "sha256sum -c -" in bootstrap
    assert "latest" not in bootstrap
