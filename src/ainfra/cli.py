"""Command-line entry point for ainfra."""

from __future__ import annotations

import argparse
import json
import sys
from collections.abc import Sequence
from pathlib import Path

from ainfra import __version__
from ainfra.contracts import validate_path
from ainfra.errors import AinfraError, GuardError


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="ainfra",
        description="Validate and operate explicit infrastructure templates.",
    )
    parser.add_argument("--version", action="version", version=__version__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    validate = subparsers.add_parser(
        "validate",
        help="validate a manifest, input, or standardized output",
    )
    validate.add_argument("path", type=Path)
    validate.add_argument("--format", choices=("text", "json"), default="text")

    subparsers.add_parser(
        "doctor",
        help="report local dependency readiness without mutation",
    )
    for command in ("plan", "apply", "destroy", "outputs"):
        lifecycle = subparsers.add_parser(
            command,
            help=f"{command} support is introduced in Milestone 2",
        )
        lifecycle.add_argument("template")
    return parser


def _run(args: argparse.Namespace) -> int:
    if args.command == "validate":
        document = validate_path(args.path)
        result = {
            "ok": True,
            "apiVersion": document["apiVersion"],
            "kind": document["kind"],
            "path": str(args.path),
        }
        if args.format == "json":
            print(json.dumps(result, sort_keys=True))
        else:
            print(f"valid {result['kind']}: {args.path}")
        return 0

    if args.command == "doctor":
        print("ainfra doctor: foundation contracts available")
        return 0

    raise GuardError(
        f"{args.command} is not available until the guarded lifecycle milestone"
    )


def main(argv: Sequence[str] | None = None) -> int:
    """Run the CLI and translate expected errors into stable exits."""

    try:
        return _run(_parser().parse_args(argv))
    except AinfraError as exc:
        print(str(exc), file=sys.stderr)
        return exc.exit_code
