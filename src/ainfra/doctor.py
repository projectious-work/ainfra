"""Read-only local readiness checks."""

from __future__ import annotations

import shutil
import subprocess
import sys
import tomllib
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

from ainfra.contracts import repository_root, validate_path
from ainfra.policy import validate_policy
from ainfra.template import discover_template


@dataclass(frozen=True, slots=True)
class Check:
    """One machine-readable doctor result."""

    id: str
    status: str
    found: str | None
    required: str
    remediation: str


COMMANDS = {
    "uv": ["uv", "--version"],
    "tofu": ["tofu", "version"],
    "ansible-playbook": ["ansible-playbook", "--version"],
    "checkov": ["checkov", "--version"],
    "gitleaks": ["gitleaks", "version"],
}


def run_doctor(input_path: Path | None = None) -> list[Check]:
    """Return readiness checks without installing or changing anything."""

    checks = [
        Check(
            "python",
            "pass" if sys.version_info[:2] == (3, 12) else "fail",
            ".".join(str(part) for part in sys.version_info[:3]),
            "Python 3.12",
            "Run through `uv run` with the committed .python-version.",
        )
    ]
    root = repository_root()
    tool_pins = _load_tool_pins(root / "tools.lock")
    for check_id, argv in COMMANDS.items():
        pin = tool_pins[check_id]
        required = f"{pin['policy']} {pin['version']}"
        executable = shutil.which(argv[0])
        if executable is None:
            checks.append(
                Check(
                    check_id,
                    "fail",
                    None,
                    required,
                    f"Install {argv[0]} and make it available on PATH.",
                )
            )
            continue
        result = subprocess.run(
            argv,
            check=False,
            capture_output=True,
            text=True,
            timeout=10,
        )
        output = (result.stdout or result.stderr).splitlines()
        first_line = output[0] if output else executable
        found_version = _extract_version(first_line)
        version_ok = (
            result.returncode == 0
            and found_version is not None
            and _version_matches(
                found_version,
                str(pin["version"]),
                str(pin["policy"]),
            )
        )
        checks.append(
            Check(
                check_id,
                "pass" if version_ok else "fail",
                found_version or first_line,
                required,
                f"Install {argv[0]} at {required}.",
            )
        )

    checks.extend(
        [
            _file_check(root / "uv.lock", "uv-lock"),
            _file_check(root / "tools.lock", "tool-lock"),
            _file_check(root / ".gitignore", "gitignore"),
            _python_pin_check(root / "pyproject.toml"),
            _template_check(),
            _backend_policy_check(input_path),
            Check(
                "github-workflows",
                (
                    "fail"
                    if (root / ".github" / "workflows").exists()
                    else "pass"
                ),
                None,
                "no GitHub workflows",
                "Remove .github/workflows; all automation is local.",
            ),
        ]
    )
    return checks


def serialized_checks(
    checks: list[Check] | None = None,
) -> list[dict[str, str | None]]:
    """Serialize doctor output without exposing local environment details."""

    return [asdict(check) for check in checks or run_doctor()]


def _file_check(path: Path, check_id: str) -> Check:
    return Check(
        check_id,
        "pass" if path.is_file() else "fail",
        str(path) if path.is_file() else None,
        f"{path.name} present",
        f"Restore the committed {path.name} file.",
    )


def _load_tool_pins(path: Path) -> dict[str, dict[str, Any]]:
    with path.open("rb") as handle:
        document = tomllib.load(handle)
    tools = document.get("tools")
    if not isinstance(tools, dict):
        raise ValueError("tools.lock must contain [tools.*] entries")
    return tools


def _extract_version(output: str) -> str | None:
    import re

    match = re.search(r"(?<!\d)(\d+\.\d+(?:\.\d+)?)", output)
    return match.group(1) if match else None


def _version_tuple(value: str) -> tuple[int, ...]:
    return tuple(int(part) for part in value.split("."))


def _version_matches(found: str, required: str, policy: str) -> bool:
    if policy == "exact":
        return _version_tuple(found) == _version_tuple(required)
    if policy == "minimum":
        return _version_tuple(found) >= _version_tuple(required)
    return False


def _python_pin_check(path: Path) -> Check:
    with path.open("rb") as handle:
        document = tomllib.load(handle)
    project = document["project"]
    dependencies = project.get("dependencies", [])
    dev_dependencies = document.get("dependency-groups", {}).get("dev", [])
    all_dependencies = [*dependencies, *dev_dependencies]
    pinned = all("==" in dependency for dependency in all_dependencies)
    return Check(
        "python-pins",
        "pass" if pinned else "fail",
        ", ".join(all_dependencies),
        "all Python dependencies exactly pinned",
        "Pin every runtime and development dependency with ==.",
    )


def _template_check() -> Check:
    try:
        template = discover_template("hetzner-kubernetes-baseline")
    except Exception as exc:
        return Check(
            "template-contract",
            "fail",
            str(exc),
            "reference template resolves safely",
            "Repair the reference manifest and its contained paths.",
        )
    return Check(
        "template-contract",
        "pass",
        str(template.manifest_path),
        "reference template resolves safely",
        "Repair the reference manifest and its contained paths.",
    )


def _backend_policy_check(input_path: Path | None = None) -> Check:
    root = repository_root()
    path = input_path or (
        root
        / "templates"
        / "hetzner-kubernetes-baseline"
        / "inputs"
        / "example.input.yaml"
    )
    try:
        document = validate_path(path)
        validate_policy(
            document,
            source=path,
            template_name="hetzner-kubernetes-baseline",
        )
    except Exception as exc:
        return Check(
            "backend-policy",
            "fail",
            str(exc),
            "example input satisfies state policy",
            "Repair the example state mode and backend capabilities.",
        )
    return Check(
        "backend-policy",
        "pass",
        str(path),
        "example input satisfies state policy",
        "Repair the example state mode and backend capabilities.",
    )
