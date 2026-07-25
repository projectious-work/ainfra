"""Command-line entry point for ainfra."""

from __future__ import annotations

import argparse
import json
import sys
from collections.abc import Sequence
from dataclasses import asdict
from pathlib import Path

from ainfra import __version__
from ainfra.contracts import validate_path
from ainfra.doctor import run_doctor, serialized_checks
from ainfra.errors import AinfraError, DependencyError, GuardError
from ainfra.inventory import write_inventory
from ainfra.lifecycle import Lifecycle
from ainfra.policy import validate_policy
from ainfra.runner import SubprocessRunner
from ainfra.template import discover_template


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
    validate.add_argument(
        "target",
        help="document path or template name beneath templates/",
    )
    validate.add_argument(
        "--input",
        type=Path,
        help="input document to validate with the selected template",
    )
    validate.add_argument("--format", choices=("text", "json"), default="text")

    doctor = subparsers.add_parser(
        "doctor",
        help="report local dependency readiness without mutation",
    )
    doctor.add_argument("--format", choices=("text", "json"), default="text")
    doctor.add_argument(
        "--input",
        type=Path,
        help="validate backend readiness for this actual environment input",
    )
    plan = subparsers.add_parser("plan", help="create a reviewable plan")
    plan.add_argument("template")
    plan.add_argument("--input", type=Path, required=True)
    plan.add_argument(
        "--destroy",
        action="store_true",
        help="create a reviewed destroy plan without applying it",
    )
    plan.add_argument("--format", choices=("text", "json"), default="text")

    apply = subparsers.add_parser("apply", help="apply an exact reviewed plan")
    apply.add_argument("template")
    apply.add_argument("--input", type=Path, required=True)
    apply.add_argument("--approve", required=True, metavar="PLAN_ID")

    destroy = subparsers.add_parser(
        "destroy",
        help="destroy an exact ownership scope",
    )
    destroy.add_argument("template")
    destroy.add_argument("--input", type=Path, required=True)
    destroy.add_argument(
        "--approve-destroy",
        required=True,
        metavar="SCOPE_TOKEN",
    )

    outputs = subparsers.add_parser(
        "outputs",
        help="read the sanitized standardized output",
    )
    outputs.add_argument("template")
    outputs.add_argument("--format", choices=("json", "yaml"), default="json")

    inventory = subparsers.add_parser(
        "inventory",
        help="generate Ansible inventory from standardized output",
    )
    inventory.add_argument("--output", type=Path, required=True)
    inventory.add_argument("--destination", type=Path, required=True)
    return parser


def _run(args: argparse.Namespace) -> int:
    if args.command == "validate":
        path = Path(args.target)
        if path.exists():
            document = validate_path(path)
            validate_policy(document, source=path)
            result = {
                "ok": True,
                "apiVersion": document["apiVersion"],
                "kind": document["kind"],
                "path": str(path),
            }
        else:
            template = discover_template(args.target)
            validated = [str(template.manifest_path)]
            if args.input is not None:
                input_document = validate_path(args.input)
                validate_policy(
                    input_document,
                    source=args.input,
                    template_name=template.name,
                )
                validated.append(str(args.input))
            result = {
                "ok": True,
                "apiVersion": template.manifest["apiVersion"],
                "kind": template.manifest["kind"],
                "template": template.name,
                "validated": validated,
            }
        if args.format == "json":
            print(json.dumps(result, sort_keys=True))
        else:
            print(f"valid {result['kind']}: {args.target}")
        return 0

    if args.command == "doctor":
        checks = run_doctor(args.input)
        if args.format == "json":
            print(json.dumps(serialized_checks(checks), sort_keys=True))
        else:
            for check in checks:
                found = f" ({check.found})" if check.found else ""
                print(f"{check.status:4} {check.id}{found}")
        failures = [check for check in checks if check.status == "fail"]
        if failures:
            names = ", ".join(check.id for check in failures)
            raise DependencyError(f"doctor checks failed: {names}")
        return 0

    lifecycle = Lifecycle(SubprocessRunner())
    if args.command == "plan":
        record = lifecycle.plan(
            args.template,
            args.input,
            destroy=args.destroy,
        )
        if args.format == "json":
            print(json.dumps(asdict(record), sort_keys=True))
        else:
            print(f"plan {record.id}")
            print(f"{record.operation} approval: {record.id}")
        return 0
    if args.command == "apply":
        command_result = lifecycle.apply(
            args.template,
            args.input,
            args.approve,
        )
        print(command_result.stdout, end="")
        return 0
    if args.command == "destroy":
        command_result = lifecycle.destroy(
            args.template,
            args.input,
            args.approve_destroy,
        )
        print(command_result.stdout, end="")
        return 0
    if args.command == "outputs":
        document = lifecycle.outputs(args.template)
        if args.format == "yaml":
            import yaml  # type: ignore[import-untyped]

            print(yaml.safe_dump(document, sort_keys=False), end="")
        else:
            print(json.dumps(document, indent=2, sort_keys=True))
        return 0
    if args.command == "inventory":
        write_inventory(args.output, args.destination)
        print(f"inventory written: {args.destination}")
        return 0
    raise GuardError(f"unsupported lifecycle command: {args.command}")


def main(argv: Sequence[str] | None = None) -> int:
    """Run the CLI and translate expected errors into stable exits."""

    try:
        return _run(_parser().parse_args(argv))
    except AinfraError as exc:
        print(str(exc), file=sys.stderr)
        return exc.exit_code
