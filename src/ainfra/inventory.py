"""Deterministic Ansible inventory generation from validated output."""

from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml  # type: ignore[import-untyped]

from ainfra.contracts import validate_path
from ainfra.errors import GuardError
from ainfra.policy import validate_policy


def build_inventory(output_path: Path) -> dict[str, Any]:
    """Build a stable, secret-free inventory from an InfrastructureOutput."""

    document = validate_path(output_path)
    if document["kind"] != "InfrastructureOutput":
        raise GuardError("inventory requires an InfrastructureOutput document")
    validate_policy(document, source=output_path)

    children: dict[str, Any] = {
        "control_plane": {"hosts": {}},
        "workers": {"hosts": {}},
    }
    for node in sorted(
        document["spec"]["nodes"], key=lambda item: item["name"]
    ):
        group = (
            "control_plane"
            if node["role"] == "control-plane-capable"
            else "workers"
        )
        children[group]["hosts"][node["name"]] = {
            "ansible_host": node["privateIPv4"],
            "ainfra_image": node["image"],
            "ainfra_role": node["role"],
        }

    return {
        "all": {
            "vars": {
                "ansible_user": "ainfra",
                "ansible_python_interpreter": "/usr/bin/python3",
                "ainfra_private_cidr": document["spec"]["network"][
                    "privateCidrs"
                ][0],
            },
            "children": children,
        }
    }


def write_inventory(output_path: Path, destination: Path) -> None:
    """Write inventory atomically without emitting credentials."""

    inventory = build_inventory(output_path)
    destination.parent.mkdir(parents=True, exist_ok=True)
    temporary = destination.with_suffix(f"{destination.suffix}.tmp")
    temporary.write_text(
        yaml.safe_dump(inventory, sort_keys=False),
        encoding="utf-8",
    )
    temporary.replace(destination)
