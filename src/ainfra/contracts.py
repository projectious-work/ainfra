"""Load and validate ainfra v1alpha1 contract documents."""

from __future__ import annotations

import json
from ipaddress import IPv4Network, IPv6Network
from pathlib import Path
from typing import Any, cast

import yaml  # type: ignore[import-untyped]
from jsonschema import (  # type: ignore[import-untyped]
    Draft202012Validator,
    FormatChecker,
)

from ainfra.errors import ContractError

API_VERSION = "ainfra.projectious.work/v1alpha1"
FORMAT_CHECKER = FormatChecker()
SCHEMA_FILES = {
    "InfrastructureTemplate": "template-manifest.v1alpha1.json",
    "TemplateInput": "template-input.v1alpha1.json",
    "InfrastructureOutput": "template-output.v1alpha1.json",
}


@FORMAT_CHECKER.checks("ipv4-network", raises=ValueError)  # type: ignore[misc]
def _is_ipv4_network(value: object) -> bool:
    if not isinstance(value, str):
        return False
    IPv4Network(value, strict=True)
    return True


@FORMAT_CHECKER.checks("ipv6-network", raises=ValueError)  # type: ignore[misc]
def _is_ipv6_network(value: object) -> bool:
    if not isinstance(value, str):
        return False
    IPv6Network(value, strict=True)
    return True


def repository_root() -> Path:
    """Return the source checkout root containing the public schemas."""

    return Path(__file__).resolve().parents[2]


def load_document(path: Path) -> dict[str, Any]:
    """Read a JSON or YAML object without resolving any secret reference."""

    try:
        content = path.read_text(encoding="utf-8")
    except OSError as exc:
        raise ContractError(f"cannot read {path}: {exc}") from exc

    try:
        value = (
            json.loads(content)
            if path.suffix.lower() == ".json"
            else yaml.safe_load(content)
        )
    except (json.JSONDecodeError, yaml.YAMLError) as exc:
        raise ContractError(f"cannot parse {path}: {exc}") from exc

    if not isinstance(value, dict):
        raise ContractError(f"{path}: document must be an object")
    return value


def schema_for(kind: str) -> dict[str, Any]:
    """Load the schema associated with a supported document kind."""

    filename = SCHEMA_FILES.get(kind)
    if filename is None:
        supported = ", ".join(sorted(SCHEMA_FILES))
        raise ContractError(f"unsupported kind {kind!r}; expected {supported}")
    path = repository_root() / "schemas" / filename
    return cast(
        dict[str, Any],
        json.loads(path.read_text(encoding="utf-8")),
    )


def validate_document(document: dict[str, Any], *, source: Path) -> None:
    """Validate a document and report the first deterministic error."""

    kind = document.get("kind")
    if not isinstance(kind, str):
        raise ContractError(f"{source}: /kind is required and must be a string")

    validator = Draft202012Validator(
        schema_for(kind),
        format_checker=FORMAT_CHECKER,
    )
    errors = sorted(
        validator.iter_errors(document), key=lambda err: list(err.path)
    )
    if not errors:
        return

    error = errors[0]
    pointer = "/" + "/".join(str(part) for part in error.absolute_path)
    if pointer == "/":
        pointer = "<root>"
    is_output = kind == "InfrastructureOutput"
    raise ContractError(
        f"{source}: {pointer}: {error.message}",
        output=is_output,
    )


def validate_path(path: Path) -> dict[str, Any]:
    """Load and validate one public contract document."""

    document = load_document(path)
    validate_document(document, source=path)
    return document
