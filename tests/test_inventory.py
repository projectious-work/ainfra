from pathlib import Path

import pytest
import yaml  # type: ignore[import-untyped]

from ainfra.errors import SafetyError
from ainfra.inventory import build_inventory, write_inventory

ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "tests" / "fixtures" / "output.valid.yaml"


def test_inventory_is_deterministic_and_uses_private_addresses() -> None:
    inventory = build_inventory(OUTPUT)
    control = inventory["all"]["children"]["control_plane"]["hosts"]
    workers = inventory["all"]["children"]["workers"]["hosts"]
    assert list(control) == ["fixture-control-01"]
    assert control["fixture-control-01"]["ansible_host"] == "10.42.0.10"
    assert workers["fixture-worker-01"]["ansible_host"] == "10.42.0.11"
    assert inventory["all"]["vars"]["ainfra_private_cidr"] == "10.42.0.0/16"


def test_inventory_write_contains_no_public_or_secret_fields(
    tmp_path: Path,
) -> None:
    destination = tmp_path / "inventory.yml"
    write_inventory(OUTPUT, destination)
    rendered = destination.read_text(encoding="utf-8")
    assert yaml.safe_load(rendered) == build_inventory(OUTPUT)
    assert "publicIPv" not in rendered
    assert "credential" not in rendered


def test_inventory_rejects_secret_looking_output(tmp_path: Path) -> None:
    document = OUTPUT.read_text(encoding="utf-8").replace(
        "fixture-worker-01",
        "AINFRA_TEST_SECRET_DO_NOT_USE",
    )
    unsafe = tmp_path / "unsafe-output.yaml"
    unsafe.write_text(document, encoding="utf-8")
    with pytest.raises(SafetyError, match="resembles secret"):
        build_inventory(unsafe)
