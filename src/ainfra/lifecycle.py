"""Guarded OpenTofu lifecycle orchestration."""

from __future__ import annotations

import hashlib
import json
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

from ainfra.contracts import repository_root, validate_path
from ainfra.errors import AinfraError, GuardError
from ainfra.policy import validate_policy
from ainfra.runner import Result, Runner, child_environment
from ainfra.template import Template, discover_template


@dataclass(frozen=True, slots=True)
class PlanRecord:
    """Binding between reviewed input and an immutable plan file."""

    id: str
    template: str
    template_version: str
    environment: str
    input_path: str
    input_sha256: str
    plan_path: str
    destroy_scope: str


class Lifecycle:
    """Execute lifecycle operations through an injectable safe runner."""

    def __init__(self, runner: Runner) -> None:
        self.runner = runner

    def plan(self, template_name: str, input_path: Path) -> PlanRecord:
        template, document = self._validated(template_name, input_path)
        record = _plan_record(template, document, input_path)
        run_dir = Path(record.plan_path).parent
        run_dir.mkdir(parents=True, exist_ok=True)
        variables = run_dir / "input.auto.tfvars.json"
        variables.write_text(
            json.dumps(_tofu_variables(document), indent=2, sort_keys=True),
            encoding="utf-8",
        )
        environment, secrets = _environment(document)
        tofu_root = _tofu_root(template)
        self._required(
            self.runner.run(
                ["tofu", "init", "-input=false", "-reconfigure"],
                cwd=tofu_root,
                env=environment,
                secrets=secrets,
            )
        )
        self._required(
            self.runner.run(
                [
                    "tofu",
                    "plan",
                    "-input=false",
                    f"-out={record.plan_path}",
                    f"-var-file={variables}",
                ],
                cwd=tofu_root,
                env=environment,
                secrets=secrets,
            )
        )
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
        if not approval:
            raise GuardError("apply requires the exact reviewed plan ID")
        record = load_plan(approval)
        template, document = self._validated(template_name, input_path)
        _verify_record(record, template, document, input_path)
        environment, secrets = _environment(document)
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
        template, document = self._validated(template_name, input_path)
        expected = destruction_token(template, document, input_path)
        if approval != expected:
            raise GuardError(
                "destroy approval mismatch; expected the exact scope token "
                f"for {template.name}/{document['metadata']['environment']}"
            )
        environment, secrets = _environment(document)
        return self._required(
            self.runner.run(
                ["tofu", "destroy", "-input=false", "-auto-approve"],
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


def load_plan(plan_id: str) -> PlanRecord:
    """Load a plan binding by its exact digest."""

    path = _record_path(plan_id)
    if not path.is_file():
        raise GuardError(f"reviewed plan does not exist: {plan_id}")
    return PlanRecord(**json.loads(path.read_text(encoding="utf-8")))


def destruction_token(
    template: Template,
    document: dict[str, Any],
    input_path: Path,
) -> str:
    """Return the exact ownership scope required for destroy approval."""

    digest = _input_hash(input_path)
    environment = document["metadata"]["environment"]
    payload = f"{template.name}:{environment}:{digest}".encode()
    return "destroy-" + hashlib.sha256(payload).hexdigest()[:16]


def _plan_record(
    template: Template,
    document: dict[str, Any],
    input_path: Path,
) -> PlanRecord:
    digest = _input_hash(input_path)
    identity = (
        f"{template.name}:{template.manifest['metadata']['version']}:{digest}"
    )
    plan_id = hashlib.sha256(identity.encode()).hexdigest()[:20]
    plan_path = _run_root() / plan_id / "plan.tfplan"
    return PlanRecord(
        plan_id,
        template.name,
        template.manifest["metadata"]["version"],
        document["metadata"]["environment"],
        str(input_path.resolve()),
        digest,
        str(plan_path),
        destruction_token(template, document, input_path),
    )


def _verify_record(
    record: PlanRecord,
    template: Template,
    document: dict[str, Any],
    input_path: Path,
) -> None:
    expected = _plan_record(template, document, input_path)
    if record != expected:
        raise GuardError("reviewed plan binding is stale or mismatched")
    if not Path(record.plan_path).is_file():
        raise GuardError("reviewed plan file is missing")


def _input_hash(path: Path) -> str:
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


def _environment(document: dict[str, Any]) -> tuple[dict[str, str], list[str]]:
    reference = document["spec"]["provider"]["projectTokenRef"]
    names = [reference["name"]] if reference["type"] == "environment" else []
    return child_environment(names)


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
