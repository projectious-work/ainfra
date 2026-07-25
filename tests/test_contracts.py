"""Contract and fixture tests for the v1alpha1 foundation."""

from __future__ import annotations

import json
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

from ainfra.contracts import SCHEMA_FILES, schema_for, validate_path
from ainfra.errors import ContractError

ROOT = Path(__file__).resolve().parents[1]
FIXTURES = ROOT / "tests" / "fixtures" / "contracts" / "v1alpha1"


@pytest.mark.parametrize("kind", sorted(SCHEMA_FILES))
def test_schema_is_valid_draft_2020_12(kind: str) -> None:
    Draft202012Validator.check_schema(schema_for(kind))


@pytest.mark.parametrize(
    "path",
    sorted((FIXTURES / "valid").iterdir()),
)
def test_valid_contract_fixtures(path: Path) -> None:
    document = validate_path(path)
    assert document["apiVersion"] == "ainfra.projectious.work/v1alpha1"


@pytest.mark.parametrize(
    "path",
    sorted((FIXTURES / "invalid").iterdir()),
)
def test_invalid_contract_fixtures(path: Path) -> None:
    with pytest.raises(ContractError) as caught:
        validate_path(path)
    assert caught.value.code in {"AINFRA-E200", "AINFRA-E300"}
    assert str(path) in caught.value.message


def test_manifest_rejects_parent_path(tmp_path: Path) -> None:
    source = FIXTURES / "valid" / "template-manifest.yaml"
    document = validate_path(source)
    document["spec"]["engines"]["tofu"]["workingDirectory"] = "../outside"
    path = tmp_path / "manifest.json"
    path.write_text(json.dumps(document), encoding="utf-8")

    with pytest.raises(ContractError, match="workingDirectory"):
        validate_path(path)


def test_output_rejects_kubernetes_claim(tmp_path: Path) -> None:
    source = FIXTURES / "valid" / "template-output.json"
    document = validate_path(source)
    document["spec"]["target"]["type"] = "kubernetes"
    path = tmp_path / "output.json"
    path.write_text(json.dumps(document), encoding="utf-8")

    with pytest.raises(ContractError) as caught:
        validate_path(path)
    assert caught.value.code == "AINFRA-E300"


def test_input_rejects_malformed_private_network(tmp_path: Path) -> None:
    source = FIXTURES / "valid" / "template-input.json"
    document = validate_path(source)
    document["spec"]["network"]["privateCidr"] = "10.42.0.3/16"
    path = tmp_path / "input.json"
    path.write_text(json.dumps(document), encoding="utf-8")

    with pytest.raises(ContractError, match="privateCidr"):
        validate_path(path)
