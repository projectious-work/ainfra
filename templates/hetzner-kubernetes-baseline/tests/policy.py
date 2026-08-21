#!/usr/bin/env python3
"""Offline Phase 9 policy and native-variable documentation checks."""

from pathlib import Path
import re


ROOT = Path(__file__).resolve().parents[1]


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(f"phase9 policy failed: {message}")


manifest = (ROOT / "ainfra-template.yaml").read_text()
variables = (ROOT / "tofu/variables.tf").read_text()
docs = (ROOT / "docs/variables.md").read_text()
terraform = "\n".join(path.read_text() for path in (ROOT / "tofu").glob("*.tf"))
ansible = "\n".join(path.read_text() for path in (ROOT / "ansible").rglob("*.*"))

require("apiVersion: ainfra.projectious.work/v1" in manifest, "v1 manifest missing")
require('version = "1.64.0"' in terraform, "hcloud provider is not exact-pinned")
require((ROOT / "tofu/.terraform.lock.hcl").is_file(), "provider lockfile missing")
require("host_key_checking = True" in ansible, "host-key checking is not enabled")
require("AllowAgentForwarding no" in ansible, "agent forwarding is not disabled")
require("random_password" not in terraform, "password generation is prohibited")
require("tls_private_key" not in terraform, "private-key generation is prohibited")
require("ssh-keyscan" not in ansible.lower(), "ssh-keyscan cannot establish trust")
require('default     = false' in variables, "bastion must default to disabled")
require('temporary = "true"' in terraform, "temporary resources need labels")
for port in ("2379-2380", "6443", "10250", "8472"):
    require(port in terraform and port in ansible, f"private K3s port {port} missing")
require("cloudflare_tunnel_token" in ansible, "external tunnel token is missing")
require("no_log: true" in ansible, "secret-bearing tasks must suppress logs")
require("ainfra_management_ingress_cidrs" not in ansible, "direct node SSH exception exists")
require("ProxyJump" in (ROOT / "README.md").read_text(), "bastion workflow missing")

headings = [
    docs.index("## OpenTofu variables"),
    docs.index("## Ansible variables"),
    docs.index("## Cross-variable rules"),
    docs.index("## Examples"),
]
require(headings == sorted(headings), "variable-reference headings are out of order")

for name in re.findall(r'variable "([a-z0-9_]+)"', variables):
    require(f"`{name}`" in docs, f"OpenTofu variable {name} is undocumented")

defaults = ROOT / "ansible/group_vars/all.yml"
for name in re.findall(r"^([a-z][a-z0-9_]+):", defaults.read_text(), re.MULTILINE):
    require(f"`{name}`" in docs, f"Ansible variable {name} is undocumented")

print("phase9 offline policy checks passed")
