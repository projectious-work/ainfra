"""Read-only local readiness checks."""

from __future__ import annotations

import shutil
import subprocess
import sys
from dataclasses import asdict, dataclass
from pathlib import Path

from ainfra.contracts import repository_root


@dataclass(frozen=True, slots=True)
class Check:
    """One machine-readable doctor result."""

    id: str
    status: str
    found: str | None
    required: str
    remediation: str


TOOLS = {
    "uv": (["uv", "--version"], "uv with the project lockfile"),
    "tofu": (["tofu", "version"], "OpenTofu >= 1.10"),
    "ansible": (
        ["ansible-playbook", "--version"],
        "ansible-playbook >= 2.16",
    ),
    "checkov": (["checkov", "--version"], "pinned Checkov"),
    "gitleaks": (["gitleaks", "version"], "pinned Gitleaks"),
}


def run_doctor() -> list[Check]:
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
    for check_id, (argv, required) in TOOLS.items():
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
        checks.append(
            Check(
                check_id,
                "pass" if result.returncode == 0 else "fail",
                output[0] if output else executable,
                required,
                f"Verify the installed {argv[0]} version.",
            )
        )

    root = repository_root()
    checks.extend(
        [
            _file_check(root / "uv.lock", "uv-lock"),
            _file_check(root / ".gitignore", "gitignore"),
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


def serialized_checks() -> list[dict[str, str | None]]:
    """Serialize doctor output without exposing local environment details."""

    return [asdict(check) for check in run_doctor()]


def _file_check(path: Path, check_id: str) -> Check:
    return Check(
        check_id,
        "pass" if path.is_file() else "fail",
        str(path) if path.is_file() else None,
        f"{path.name} present",
        f"Restore the committed {path.name} file.",
    )
