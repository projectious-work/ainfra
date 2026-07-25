"""Visible, non-shell subprocess execution with output redaction."""

from __future__ import annotations

import os
import subprocess
from collections.abc import Sequence
from dataclasses import dataclass
from pathlib import Path
from typing import Protocol

from ainfra.errors import DependencyError


@dataclass(frozen=True, slots=True)
class Result:
    """Sanitized subprocess result."""

    argv: tuple[str, ...]
    returncode: int
    stdout: str
    stderr: str


class Runner(Protocol):
    """Injectable command runner used by lifecycle orchestration."""

    def run(
        self,
        argv: Sequence[str],
        *,
        cwd: Path,
        env: dict[str, str],
        secrets: Sequence[str] = (),
    ) -> Result: ...


class SubprocessRunner:
    """Run an explicit argument vector without a shell."""

    def run(
        self,
        argv: Sequence[str],
        *,
        cwd: Path,
        env: dict[str, str],
        secrets: Sequence[str] = (),
    ) -> Result:
        try:
            process = subprocess.run(
                list(argv),
                cwd=cwd,
                env=env,
                check=False,
                capture_output=True,
                text=True,
            )
        except FileNotFoundError as exc:
            raise DependencyError(
                f"required executable not found: {argv[0]}"
            ) from exc
        return Result(
            tuple(argv),
            process.returncode,
            redact(process.stdout, secrets),
            redact(process.stderr, secrets),
        )


def child_environment(
    secret_names: Sequence[str],
) -> tuple[dict[str, str], list[str]]:
    """Build a minimal environment and collect values only for redaction."""

    allowed = ("PATH", "HOME", "LANG", "LC_ALL", "TMPDIR")
    environment = {key: os.environ[key] for key in allowed if key in os.environ}
    secret_values: list[str] = []
    for name in secret_names:
        value = os.environ.get(name)
        if value is None:
            raise DependencyError(
                f"referenced environment variable is unset: {name}"
            )
        environment[name] = value
        secret_values.append(value)
    return environment, secret_values


def redact(value: str, secrets: Sequence[str]) -> str:
    """Replace known non-empty secret values in captured output."""

    result = value
    for secret in secrets:
        if secret:
            result = result.replace(secret, "[REDACTED]")
    return result
