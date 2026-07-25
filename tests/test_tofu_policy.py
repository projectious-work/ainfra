"""Static invariants for the Hetzner OpenTofu reference template."""

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
TOFU = ROOT / "templates" / "hetzner-kubernetes-baseline" / "tofu"


def _read(name: str) -> str:
    return (TOFU / name).read_text(encoding="utf-8")


def test_provider_is_exactly_pinned() -> None:
    assert 'version = "1.64.0"' in _read("versions.tf")


def test_public_addresses_default_to_disabled() -> None:
    variables = _read("variables.tf")
    assert 'variable "public_ipv4"' in variables
    assert 'variable "public_ipv6"' in variables
    assert variables.count("default     = false") >= 2
    servers = _read("servers.tf")
    assert "ipv4_enabled = var.public_ipv4" in servers
    assert "ipv6_enabled = var.public_ipv6" in servers


def test_no_generated_private_key_or_password_resource() -> None:
    contents = "\n".join(
        path.read_text(encoding="utf-8") for path in TOFU.glob("*.tf")
    )
    assert 'resource "tls_private_key"' not in contents
    assert 'resource "random_password"' not in contents
    assert "root_password" not in contents


def test_all_managed_resources_use_ownership_labels() -> None:
    for name in ("network.tf", "servers.tf", "firewall.tf"):
        contents = _read(name)
        assert "local.ownership_labels" in contents


def test_no_public_ssh_rule_without_explicit_cidrs() -> None:
    firewall = _read("firewall.tf")
    assert "var.management_ingress_cidrs" in firewall
    assert "var.public_ipv4 || var.public_ipv6" in firewall


def test_firewall_is_attached_during_server_creation() -> None:
    servers = _read("servers.tf")
    assert servers.count("firewall_ids = [hcloud_firewall.nodes.id]") == 2
    assert 'resource "hcloud_firewall_attachment"' not in _read("firewall.tf")


def test_direct_tool_management_cidrs_are_narrow() -> None:
    variables = _read("variables.tf")
    assert "management CIDRs must be IPv4 /24 or IPv6 /64" in variables
    assert "strcontains(cidr" in variables
