"""Semantic policy tests beyond JSON Schema shape."""

from __future__ import annotations

import copy
from pathlib import Path

import pytest

from ainfra.contracts import validate_path
from ainfra.errors import SafetyError
from ainfra.policy import validate_policy
from ainfra.template import discover_template

ROOT = Path(__file__).resolve().parents[1]
VALID = ROOT / "tests" / "fixtures" / "contracts" / "v1alpha1" / "valid"


def _input() -> dict[str, object]:
    return copy.deepcopy(validate_path(VALID / "template-input.json"))


def test_discovers_reference_template() -> None:
    template = discover_template("hetzner-kubernetes-baseline")
    assert template.root.parent == ROOT / "templates"
    assert template.manifest["metadata"]["name"] == template.name


@pytest.mark.parametrize("name", ["../outside", "/absolute", "nested/name"])
def test_rejects_unsafe_template_name(name: str) -> None:
    with pytest.raises(Exception):
        discover_template(name)


def test_non_disposable_input_requires_remote_state(tmp_path: Path) -> None:
    document = _input()
    document["metadata"]["disposable"] = False  # type: ignore[index]
    with pytest.raises(SafetyError, match=r"\[P101\]"):
        validate_policy(document, source=tmp_path / "input.json")


def test_remote_state_requires_reference(tmp_path: Path) -> None:
    document = _input()
    document["metadata"]["disposable"] = False  # type: ignore[index]
    document["spec"]["state"]["mode"] = "remote"  # type: ignore[index]
    with pytest.raises(SafetyError, match=r"\[P103\]"):
        validate_policy(document, source=tmp_path / "input.json")


def test_remote_state_capability_contract_passes(tmp_path: Path) -> None:
    document = _input()
    document["metadata"]["disposable"] = False  # type: ignore[index]
    state = document["spec"]["state"]  # type: ignore[index]
    state["mode"] = "remote"
    state["backendConfigRef"] = {
        "type": "local-file",
        "name": ".ainfra/backend.hcl",
    }
    state["capabilities"] = {
        "encryptedAtRest": True,
        "locking": True,
        "versionRecovery": True,
        "tls": True,
        "accessControl": True,
    }
    validate_policy(document, source=tmp_path / "input.json")


def test_rejects_broad_management_network(tmp_path: Path) -> None:
    document = _input()
    document["spec"]["network"]["managementIngressCidrs"] = [  # type: ignore[index]
        "10.0.0.0/8"
    ]
    with pytest.raises(SafetyError, match=r"\[P105\]"):
        validate_policy(document, source=tmp_path / "input.json")


def test_rejects_traversing_local_file_reference(tmp_path: Path) -> None:
    document = _input()
    document["spec"]["provider"]["projectTokenRef"] = {  # type: ignore[index]
        "type": "local-file",
        "name": "../token",
    }
    with pytest.raises(SafetyError, match=r"\[P111\]"):
        validate_policy(document, source=tmp_path / "input.json")


def test_rejects_secret_like_unknown_content(tmp_path: Path) -> None:
    document = _input()
    document["spec"]["password"] = "fixture"  # type: ignore[index]
    with pytest.raises(SafetyError, match=r"\[P300\]"):
        validate_policy(document, source=tmp_path / "input.json")
