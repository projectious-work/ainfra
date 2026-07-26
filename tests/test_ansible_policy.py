from pathlib import Path

ROOT = (
    Path(__file__).resolve().parents[1]
    / "templates"
    / "hetzner-kubernetes-baseline"
    / "ansible"
)


def _read(relative: str) -> str:
    return (ROOT / relative).read_text(encoding="utf-8")


def test_sshd_is_validated_before_reload() -> None:
    tasks = _read("roles/ssh/tasks/main.yml")
    assert "validate: /usr/sbin/sshd -t -f %s" in tasks
    assert "PasswordAuthentication no" in _read(
        "roles/ssh/templates/60-ainfra-hardening.conf.j2"
    )


def test_firewall_is_default_deny_and_validated() -> None:
    tasks = _read("roles/firewall/tasks/main.yml")
    policy = _read("roles/firewall/templates/nftables.conf.j2")
    assert "validate: /usr/sbin/nft -c -f %s" in tasks
    assert "policy drop" in policy
    assert "ainfra_private_cidr" in policy
    assert "ainfra_private_cidr:" not in _read("group_vars/all.yml")
    assert "ainfra_management_ingress_cidrs: []" in _read("group_vars/all.yml")


def test_reboot_policy_is_explicit_and_disabled_by_default() -> None:
    assert "ainfra_unattended_reboot: false" in _read("group_vars/all.yml")
    assert "Automatic-Reboot" in _read(
        "roles/base/templates/52ainfra-unattended-upgrades.j2"
    )


def test_new_package_service_waits_for_apply_after_check_mode() -> None:
    base_tasks = _read("roles/base/tasks/main.yml")
    observability_tasks = _read("roles/observability/tasks/main.yml")
    assert "name: chrony" in base_tasks
    assert "when: not ansible_check_mode" in base_tasks
    assert "name: auditd" in observability_tasks
    assert "when: not ansible_check_mode" in observability_tasks


def test_collections_are_explicitly_declared() -> None:
    requirements = _read("requirements.yml")
    assert "collections: []" in requirements
    assert "latest" not in requirements


def test_local_gate_and_live_idempotence_coverage_are_scripted() -> None:
    validate = (ROOT.parents[2] / "scripts" / "validate-all").read_text(
        encoding="utf-8"
    )
    live = (ROOT.parents[2] / "scripts" / "verify-ansible-host").read_text(
        encoding="utf-8"
    )
    assert "scripts/test-ansible" in validate
    assert "ansible-lint" in (
        ROOT.parents[2] / "scripts" / "test-ansible"
    ).read_text(encoding="utf-8")
    assert "--check --diff" in live
    assert "changed=0" in live
