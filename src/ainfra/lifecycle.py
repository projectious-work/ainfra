"""Guarded OpenTofu lifecycle orchestration."""

from __future__ import annotations

import hashlib
import json
import os
import uuid
from dataclasses import asdict, dataclass, replace
from pathlib import Path
from typing import Any

from ainfra.contracts import repository_root, validate_path
from ainfra.errors import AinfraError, DependencyError, GuardError
from ainfra.policy import validate_policy
from ainfra.runner import Result, Runner, child_environment
from ainfra.template import Template, discover_template


@dataclass(frozen=True, slots=True)
class PlanRecord:
    """Immutable binding between inputs, configuration, and plan bytes."""

    id: str
    operation: str
    template: str
    template_version: str
    environment: str
    input_path: str
    input_sha256: str
    template_sha256: str
    plan_path: str
    plan_sha256: str


class Lifecycle:
    """Execute lifecycle operations through an injectable safe runner."""

    def __init__(self, runner: Runner) -> None:
        self.runner = runner

    def plan(
        self,
        template_name: str,
        input_path: Path,
        *,
        destroy: bool = False,
    ) -> PlanRecord:
        template, document = self._validated(template_name, input_path)
        operation = "destroy" if destroy else "apply"
        record = _new_plan_record(
            template,
            document,
            input_path,
            operation=operation,
        )
        run_dir = Path(record.plan_path).parent
        run_dir.mkdir(parents=True, exist_ok=False)
        variables = run_dir / "input.auto.tfvars.json"
        variables.write_text(
            json.dumps(_tofu_variables(document), indent=2, sort_keys=True),
            encoding="utf-8",
        )
        environment, secrets = _environment(document)
        tofu_root = _tofu_root(template)
        self._initialize(
            document,
            tofu_root,
            environment,
            secrets,
        )
        plan_argv = ["tofu", "plan", "-input=false"]
        if destroy:
            plan_argv.append("-destroy")
        plan_argv.extend(
            [
                f"-out={record.plan_path}",
                f"-var-file={variables}",
            ]
        )
        self._required(
            self.runner.run(
                plan_argv,
                cwd=tofu_root,
                env=environment,
                secrets=secrets,
            )
        )
        plan_path = Path(record.plan_path)
        if not plan_path.is_file():
            raise GuardError("OpenTofu did not produce the declared plan file")
        record = replace(record, plan_sha256=_file_hash(plan_path))
        _record_path(record.id).write_text(
            json.dumps(asdict(record), indent=2, sort_keys=True),
            encoding="utf-8",
        )
        return record

    def apply(
        self,
        template_name: str,
        input_path: Path,
        approval: str,
    ) -> Result:
        record = self._approved_record(
            template_name,
            input_path,
            approval,
            operation="apply",
        )
        template, document = self._validated(template_name, input_path)
        environment, secrets = _environment(document)
        self._initialize(
            document,
            _tofu_root(template),
            environment,
            secrets,
        )
        return self._required(
            self.runner.run(
                ["tofu", "apply", "-input=false", record.plan_path],
                cwd=_tofu_root(template),
                env=environment,
                secrets=secrets,
            )
        )

    def destroy(
        self,
        template_name: str,
        input_path: Path,
        approval: str,
    ) -> Result:
        record = self._approved_record(
            template_name,
            input_path,
            approval,
            operation="destroy",
        )
        template, document = self._validated(template_name, input_path)
        environment, secrets = _environment(document)
        self._initialize(
            document,
            _tofu_root(template),
            environment,
            secrets,
        )
        return self._required(
            self.runner.run(
                ["tofu", "apply", "-input=false", record.plan_path],
                cwd=_tofu_root(template),
                env=environment,
                secrets=secrets,
            )
        )

    def outputs(self, template_name: str) -> dict[str, Any]:
        template = discover_template(template_name)
        path = template.root / template.manifest["spec"]["output"]["file"]
        document = validate_path(path)
        validate_policy(
            document,
            source=path,
            template_name=template.name,
        )
        return document

    def _approved_record(
        self,
        template_name: str,
        input_path: Path,
        approval: str,
        *,
        operation: str,
    ) -> PlanRecord:
        if not approval:
            raise GuardError(f"{operation} requires an exact reviewed plan ID")
        record = load_plan(approval)
        if record.operation != operation:
            raise GuardError(
                f"reviewed {record.operation} plan cannot authorize {operation}"
            )
        template, document = self._validated(template_name, input_path)
        _verify_record(record, template, document, input_path)
        return record

    def _validated(
        self,
        template_name: str,
        input_path: Path,
    ) -> tuple[Template, dict[str, Any]]:
        template = discover_template(template_name)
        document = validate_path(input_path)
        validate_policy(
            document,
            source=input_path,
            template_name=template.name,
        )
        return template, document

    @staticmethod
    def _required(result: Result) -> Result:
        if result.returncode != 0:
            raise AinfraError(
                "AINFRA-E500",
                f"underlying command failed ({result.returncode}): "
                f"{result.stderr.strip()}",
                5,
            )
        return result

    def _initialize(
        self,
        document: dict[str, Any],
        tofu_root: Path,
        environment: dict[str, str],
        secrets: list[str],
    ) -> None:
        self._required(
            self.runner.run(
                [
                    "tofu",
                    "init",
                    "-input=false",
                    "-reconfigure",
                    *_backend_args(document),
                ],
                cwd=tofu_root,
                env=environment,
                secrets=secrets,
            )
        )


def load_plan(plan_id: str) -> PlanRecord:
    """Load a plan binding by its unique exact identifier."""

    path = _record_path(plan_id)
    if not path.is_file():
        raise GuardError(f"reviewed plan does not exist: {plan_id}")
    return PlanRecord(**json.loads(path.read_text(encoding="utf-8")))


def _new_plan_record(
    template: Template,
    document: dict[str, Any],
    input_path: Path,
    *,
    operation: str,
) -> PlanRecord:
    plan_id = uuid.uuid4().hex[:20]
    plan_path = _run_root() / plan_id / f"{operation}.tfplan"
    return PlanRecord(
        plan_id,
        operation,
        template.name,
        template.manifest["metadata"]["version"],
        document["metadata"]["environment"],
        str(input_path.resolve()),
        _file_hash(input_path),
        _template_hash(template),
        str(plan_path),
        "",
    )


def _verify_record(
    record: PlanRecord,
    template: Template,
    document: dict[str, Any],
    input_path: Path,
) -> None:
    if record.template != template.name:
        raise GuardError("reviewed plan belongs to another template")
    if record.template_version != template.manifest["metadata"]["version"]:
        raise GuardError("template version changed after planning")
    if record.environment != document["metadata"]["environment"]:
        raise GuardError("environment changed after planning")
    if record.input_path != str(input_path.resolve()):
        raise GuardError("input path changed after planning")
    if record.input_sha256 != _file_hash(input_path):
        raise GuardError("input content changed after planning")
    if record.template_sha256 != _template_hash(template):
        raise GuardError("template content changed after planning")
    plan_path = Path(record.plan_path)
    if not plan_path.is_file():
        raise GuardError("reviewed plan file is missing")
    if record.plan_sha256 != _file_hash(plan_path):
        raise GuardError("reviewed plan file was modified after planning")


def _template_hash(template: Template) -> str:
    digest = hashlib.sha256()
    excluded = {".ainfra", ".terraform", "__pycache__"}
    for path in sorted(template.root.rglob("*")):
        if not path.is_file() or excluded.intersection(path.parts):
            continue
        digest.update(str(path.relative_to(template.root)).encode())
        digest.update(path.read_bytes())
    return digest.hexdigest()


def _file_hash(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def _run_root() -> Path:
    path = repository_root() / ".ainfra" / "runs"
    path.mkdir(parents=True, exist_ok=True)
    return path


def _record_path(plan_id: str) -> Path:
    return _run_root() / plan_id / "plan.json"


def _tofu_root(template: Template) -> Path:
    working_directory = str(
        template.manifest["spec"]["engines"]["tofu"]["workingDirectory"]
    )
    return template.root / working_directory


def _backend_args(document: dict[str, Any]) -> list[str]:
    state = document["spec"]["state"]
    if state["mode"] != "remote":
        return []
    reference = state["backendConfigRef"]
    if reference["type"] == "environment":
        value = os.environ.get(reference["name"])
        if not value:
            raise DependencyError(
                "backend configuration path environment reference is unset: "
                f"{reference['name']}"
            )
        path = Path(value)
    else:
        path = repository_root() / reference["name"]
    resolved = path.resolve()
    if reference["type"] == "local-file":
        allowed_root = (repository_root() / ".ainfra").resolve()
        try:
            resolved.relative_to(allowed_root)
        except ValueError as exc:
            raise GuardError(
                "backend configuration escaped the .ainfra directory"
            ) from exc
    if not resolved.is_file():
        raise DependencyError(
            f"backend configuration file is missing: {resolved}"
        )
    return [f"-backend-config={resolved}"]


def _environment(document: dict[str, Any]) -> tuple[dict[str, str], list[str]]:
    references = [document["spec"]["provider"]["projectTokenRef"]]
    state = document["spec"]["state"]
    if (
        state["mode"] == "remote"
        and state["backendConfigRef"]["type"] == "environment"
    ):
        references.append(state["backendConfigRef"])
    names = [
        reference["name"]
        for reference in references
        if reference["type"] == "environment"
    ]
    environment, secrets = child_environment(names)
    provider_reference = document["spec"]["provider"]["projectTokenRef"]
    if provider_reference["type"] == "local-file":
        path = (repository_root() / provider_reference["name"]).resolve()
        if not path.is_file():
            raise DependencyError(
                f"provider credential file is missing: {path}"
            )
        value = path.read_text(encoding="utf-8").strip()
        if not value:
            raise DependencyError(f"provider credential file is empty: {path}")
        environment["HCLOUD_TOKEN"] = value
        secrets.append(value)
    return environment, secrets


def _tofu_variables(document: dict[str, Any]) -> dict[str, Any]:
    spec = document["spec"]
    return {
        "environment": document["metadata"]["environment"],
        "location": spec["provider"]["location"],
        "control_plane_count": spec["topology"]["controlPlaneCount"],
        "worker_count": spec["topology"]["workerCount"],
        "image": spec["topology"]["image"],
        "server_type": spec["topology"]["serverType"],
        "private_cidr": spec["network"]["privateCidr"],
        "public_ipv4": spec["network"]["publicIPv4"],
        "public_ipv6": spec["network"]["publicIPv6"],
        "management_ingress_cidrs": spec["network"]["managementIngressCidrs"],
        "workload_ingress_cidrs": spec["network"]["workloadIngressCidrs"],
        "admin_ssh_public_keys": spec["access"]["adminSshPublicKeys"],
    }
