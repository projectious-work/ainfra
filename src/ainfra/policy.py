"""Semantic safety policy for validated ainfra documents."""

from __future__ import annotations

import re
from ipaddress import ip_network
from pathlib import Path, PurePosixPath
from typing import Any

from ainfra.errors import SafetyError

SECRET_KEY = re.compile(
    r"(?:password|private[_-]?key|secret|token|kubeconfig)",
    re.IGNORECASE,
)
SECRET_VALUE = re.compile(
    r"(?:-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----"
    r"|hcloud_[A-Za-z0-9]{16,}|gh[opsu]_[A-Za-z0-9]{20,})"
)


def validate_policy(
    document: dict[str, Any],
    *,
    source: Path,
    template_name: str | None = None,
) -> None:
    """Apply semantic rules after structural schema validation."""

    kind = document["kind"]
    if kind == "TemplateInput":
        _reject_secret_content(document, source=source)
        _validate_input(document, source, template_name)
    elif kind == "InfrastructureOutput":
        _reject_secret_content(document, source=source)
        _validate_output(document, source, template_name)


def _validate_input(
    document: dict[str, Any],
    source: Path,
    template_name: str | None,
) -> None:
    metadata = document["metadata"]
    spec = document["spec"]
    state = spec["state"]

    if template_name is not None and metadata["template"] != template_name:
        raise _error(
            "P100",
            source,
            "/metadata/template",
            "input template does not match selected template",
        )

    disposable = metadata["disposable"]
    mode = state["mode"]
    if mode == "local-disposable" and not disposable:
        raise _error(
            "P101",
            source,
            "/spec/state/mode",
            "local state is allowed only for disposable environments",
        )
    if not disposable and mode != "remote":
        raise _error(
            "P102",
            source,
            "/spec/state/mode",
            "non-disposable environments require remote state",
        )
    if mode == "remote":
        if "backendConfigRef" not in state:
            raise _error(
                "P103",
                source,
                "/spec/state/backendConfigRef",
                "remote state requires an external backend reference",
            )
        if "capabilities" not in state:
            raise _error(
                "P104",
                source,
                "/spec/state/capabilities",
                "remote state requires declared safety capabilities",
            )

    _validate_reference(
        spec["provider"]["projectTokenRef"],
        source,
        "/spec/provider/projectTokenRef",
    )
    if "backendConfigRef" in state:
        _validate_reference(
            state["backendConfigRef"],
            source,
            "/spec/state/backendConfigRef",
        )

    for index, cidr in enumerate(spec["network"]["managementIngressCidrs"]):
        network = ip_network(cidr, strict=True)
        minimum = 24 if network.version == 4 else 64
        if network.prefixlen < minimum:
            raise _error(
                "P105",
                source,
                f"/spec/network/managementIngressCidrs/{index}",
                f"management CIDR must be /{minimum} or narrower",
            )


def _validate_output(
    document: dict[str, Any],
    source: Path,
    template_name: str | None,
) -> None:
    if (
        template_name is not None
        and document["metadata"]["template"] != template_name
    ):
        raise _error(
            "P200",
            source,
            "/metadata/template",
            "output template does not match selected template",
        )
    for index, node in enumerate(document["spec"]["nodes"]):
        if node["image"] != "debian-13":
            raise _error(
                "P201",
                source,
                f"/spec/nodes/{index}/image",
                "only debian-13 is supported",
            )


def _validate_reference(
    reference: dict[str, Any],
    source: Path,
    pointer: str,
) -> None:
    name = reference["name"]
    if reference["type"] == "environment":
        if not re.fullmatch(r"[A-Z][A-Z0-9_]*", name):
            raise _error(
                "P110",
                source,
                f"{pointer}/name",
                "environment reference must be an uppercase variable name",
            )
        return

    path = PurePosixPath(name)
    if path.is_absolute() or ".." in path.parts:
        raise _error(
            "P111",
            source,
            f"{pointer}/name",
            "local-file reference must be relative and contain no '..'",
        )


def _reject_secret_content(
    value: Any,
    *,
    source: Path,
    pointer: str = "",
) -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            child_pointer = f"{pointer}/{key}"
            if SECRET_KEY.search(key) and not key.lower().endswith("ref"):
                raise _error(
                    "P300",
                    source,
                    child_pointer,
                    "secret-bearing fields are prohibited; use a reference",
                )
            _reject_secret_content(
                child,
                source=source,
                pointer=child_pointer,
            )
    elif isinstance(value, list):
        for index, child in enumerate(value):
            _reject_secret_content(
                child,
                source=source,
                pointer=f"{pointer}/{index}",
            )
    elif isinstance(value, str) and SECRET_VALUE.search(value):
        raise _error(
            "P301",
            source,
            pointer or "<root>",
            "value resembles secret material",
        )


def _error(
    rule: str,
    source: Path,
    pointer: str,
    message: str,
) -> SafetyError:
    return SafetyError(f"[{rule}] {source}:{pointer}: {message}")
