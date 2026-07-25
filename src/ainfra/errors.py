"""Stable public error categories for the ainfra CLI."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(slots=True)
class AinfraError(Exception):
    """An expected error with a stable identifier and process exit code."""

    code: str
    message: str
    exit_code: int

    def __str__(self) -> str:
        return f"{self.code}: {self.message}"


class ContractError(AinfraError):
    """A manifest, input, or output does not satisfy its contract."""

    def __init__(self, message: str, *, output: bool = False) -> None:
        code = "AINFRA-E300" if output else "AINFRA-E200"
        super().__init__(code, message, 3)


class SafetyError(AinfraError):
    """A configuration violates an ainfra safety invariant."""

    def __init__(self, message: str) -> None:
        super().__init__("AINFRA-E400", message, 4)


class DependencyError(AinfraError):
    """A required local tool is missing or incompatible."""

    def __init__(self, message: str) -> None:
        super().__init__("AINFRA-E500", message, 5)


class GuardError(AinfraError):
    """A guarded lifecycle command lacks explicit approval."""

    def __init__(self, message: str) -> None:
        super().__init__("AINFRA-E600", message, 6)
