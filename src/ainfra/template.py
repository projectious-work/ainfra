"""Safe template discovery and manifest resolution."""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from packaging.specifiers import InvalidSpecifier, SpecifierSet
from packaging.version import Version

from ainfra import __version__
from ainfra.contracts import repository_root, validate_path
from ainfra.errors import ContractError, SafetyError

TEMPLATE_NAME = re.compile(r"^[a-z][a-z0-9-]{2,62}$")
SUPPORTED_CAPABILITIES = {
    "provider.hetzner-cloud",
    "target.kubernetes-ready",
    "access.ssh",
    "network.private",
    "security.hardening",
}


@dataclass(frozen=True, slots=True)
class Template:
    """A discovered template and its validated manifest."""

    name: str
    root: Path
    manifest_path: Path
    manifest: dict[str, Any]


def _inside(path: Path, root: Path, *, rule: str) -> Path:
    resolved = path.resolve()
    try:
        resolved.relative_to(root.resolve())
    except ValueError as exc:
        raise SafetyError(f"[{rule}] path escapes {root}: {path}") from exc
    return resolved


def discover_template(name: str) -> Template:
    """Discover one direct child beneath the repository template root."""

    if not TEMPLATE_NAME.fullmatch(name):
        raise ContractError(f"invalid template name {name!r}")

    templates_root = (repository_root() / "templates").resolve()
    root = _inside(templates_root / name, templates_root, rule="P001")
    if root.parent != templates_root or not root.is_dir():
        raise ContractError(f"template {name!r} does not exist")

    manifest_path = _inside(
        root / "ainfra-template.yaml",
        root,
        rule="P002",
    )
    manifest = validate_path(manifest_path)
    actual_name = manifest["metadata"]["name"]
    if actual_name != name:
        raise SafetyError(
            f"[P003] manifest name {actual_name!r} does not match {name!r}"
        )
    validate_manifest_compatibility(manifest)
    _validate_manifest_paths(manifest, root)
    return Template(name, root, manifest_path, manifest)


def validate_manifest_compatibility(manifest: dict[str, Any]) -> None:
    """Require the running wrapper and every declared capability."""

    wrapper_range = manifest["spec"]["requiresWrapper"]
    try:
        specifier = SpecifierSet(",".join(wrapper_range.split()))
    except InvalidSpecifier as exc:
        raise SafetyError(
            f"[P008] invalid wrapper version range {wrapper_range!r}"
        ) from exc
    running = Version(Version(__version__).base_version)
    if running not in specifier:
        raise SafetyError(
            f"[P008] wrapper {running} does not satisfy {wrapper_range!r}"
        )
    unknown = set(manifest["spec"]["capabilities"]) - SUPPORTED_CAPABILITIES
    if unknown:
        raise SafetyError(
            f"[P009] unsupported capabilities: {', '.join(sorted(unknown))}"
        )


def _validate_manifest_paths(
    manifest: dict[str, Any],
    template_root: Path,
) -> None:
    """Validate path containment without requiring future files to exist."""

    for engine in ("tofu", "ansible"):
        value = manifest["spec"]["engines"][engine]["workingDirectory"]
        _inside(template_root / value, template_root, rule="P004")

    for example in manifest["spec"]["inputs"]["examples"]:
        _inside(template_root / example, template_root, rule="P005")

    output_file = manifest["spec"]["output"]["file"]
    _inside(template_root / output_file, template_root, rule="P006")

    schema_root = (repository_root() / "schemas").resolve()
    for section in ("inputs", "output"):
        value = manifest["spec"][section]["schema"]
        resolved = _inside(
            template_root / value,
            schema_root,
            rule="P007",
        )
        if resolved.parent != schema_root:
            raise SafetyError(
                f"[P007] schema must be a direct child of {schema_root}"
            )
