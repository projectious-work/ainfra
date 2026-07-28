//! Cross-language parity coverage for pure Rust behavior.

#![allow(clippy::unwrap_used)]

use std::fs;
use std::path::{Path, PathBuf};

use ainfra::contracts::{validate_document, validate_path};
use ainfra::doctor;
use ainfra::inventory::build_inventory;
use ainfra::policy::validate_policy;
use ainfra::template::{discover_builtin, validate_manifest_compatibility};
use assert_cmd::Command;
use predicates::prelude::*;
use serde_json::{Value, json};

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn contract_fixture(group: &str, name: &str) -> PathBuf {
    root()
        .join("tests/fixtures/contracts/v1alpha1")
        .join(group)
        .join(name)
}

#[test]
fn embedded_schemas_accept_all_public_kinds() {
    for name in [
        "template-manifest.yaml",
        "template-input.json",
        "template-output.json",
    ] {
        validate_path(&contract_fixture("valid", name)).unwrap();
    }
}

#[test]
fn invalid_input_and_output_keep_stable_error_families() {
    let input =
        validate_path(&contract_fixture("invalid", "unsupported-version.json")).unwrap_err();
    assert_eq!(input.code(), "AINFRA-E200");
    assert_eq!(input.exit_code(), 3);

    let output = validate_path(&contract_fixture("invalid", "secret-output.json")).unwrap_err();
    assert_eq!(output.code(), "AINFRA-E300");
    assert_eq!(output.exit_code(), 3);
}

#[test]
fn yaml_and_non_object_documents_match_contract_boundary() {
    let directory = tempfile::tempdir().unwrap();
    let yaml = directory.path().join("input.yaml");
    let source = fs::read_to_string(contract_fixture("valid", "template-input.json")).unwrap();
    let document: Value = serde_json::from_str(&source).unwrap();
    fs::write(&yaml, serde_yaml::to_string(&document).unwrap()).unwrap();
    validate_path(&yaml).unwrap();

    let scalar = directory.path().join("scalar.yaml");
    fs::write(&scalar, "not-an-object\n").unwrap();
    let error = validate_path(&scalar).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E200");
    assert!(error.to_string().contains("document must be an object"));
}

#[test]
fn strict_network_format_rejects_host_bits() {
    let path = contract_fixture("valid", "template-input.json");
    let mut document = validate_path(&path).unwrap();
    document["spec"]["network"]["privateCidr"] = json!("10.42.0.3/16");
    let error = validate_document(&document, Path::new("fixture.json")).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E200");
    assert!(error.to_string().contains("/spec/network/privateCidr"));
}

#[test]
fn secret_policy_preserves_rule_ids_and_reference_exemption() {
    let path = contract_fixture("valid", "template-input.json");
    let mut document = validate_path(&path).unwrap();
    document["spec"]["provider"]["projectTokenRef"] =
        json!({"type": "environment", "name": "AINFRA_TEST_SECRET"});
    validate_policy(&document, &path, None, &root()).unwrap();

    document["spec"]["provider"]["projectTokenRef"]["token"] = json!("not-a-real-value");
    let error = validate_policy(&document, &path, None, &root()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E400");
    assert!(error.to_string().contains("[P300]"));
}

#[test]
fn broad_management_network_is_rejected() {
    let path = contract_fixture("valid", "template-input.json");
    let mut document = validate_path(&path).unwrap();
    document["spec"]["network"]["managementIngressCidrs"] = json!(["10.0.0.0/8"]);
    let error = validate_policy(&document, &path, None, &root()).unwrap_err();
    assert!(error.to_string().contains("[P105]"));
}

#[test]
fn inventory_is_deterministic_and_omits_public_addresses() {
    let output = contract_fixture("valid", "template-output.json");
    let yaml = build_inventory(&output, &root()).unwrap();
    let inventory: Value = serde_yaml::from_str(&yaml).unwrap();
    assert_eq!(
        inventory["all"]["vars"]["ainfra_private_cidr"],
        "10.42.0.0/16"
    );
    let host = &inventory["all"]["children"]["control_plane"]["hosts"]["control-01"];
    assert_eq!(host["ansible_host"], "10.42.0.10");
    assert!(host.get("publicIPv4").is_none());
    assert!(
        inventory["all"]["children"]["workers"]["hosts"]
            .as_object()
            .unwrap()
            .is_empty()
    );
}

#[test]
fn rust_cli_validates_and_writes_inventory() {
    let input = contract_fixture("valid", "template-input.json");
    Command::cargo_bin("ainfra")
        .unwrap()
        .args(["validate", input.to_str().unwrap(), "--format", "json"])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"ok\":true"));

    let directory = tempfile::tempdir().unwrap();
    let destination = directory.path().join("inventory.yaml");
    let output = contract_fixture("valid", "template-output.json");
    Command::cargo_bin("ainfra")
        .unwrap()
        .args([
            "inventory",
            "--output",
            output.to_str().unwrap(),
            "--destination",
            destination.to_str().unwrap(),
        ])
        .assert()
        .success();
    assert!(destination.is_file());
    assert!(!destination.with_extension("yaml.tmp").exists());
}

#[test]
fn built_in_template_is_checkout_independent() {
    let template = discover_builtin("hetzner-kubernetes-baseline").unwrap();
    assert_eq!(template.manifest["metadata"]["name"], template.name);
    let error = discover_builtin("../baseline").unwrap_err();
    assert_eq!(error.code(), "AINFRA-E200");
}

#[test]
fn wrapper_and_capability_policy_keep_rule_ids() {
    let template = discover_builtin("hetzner-kubernetes-baseline").unwrap();
    let mut incompatible = template.manifest.clone();
    incompatible["spec"]["requiresWrapper"] = json!(">=9");
    let error = validate_manifest_compatibility(&incompatible).unwrap_err();
    assert!(error.to_string().contains("[P008]"));

    let mut unsupported = template.manifest;
    unsupported["spec"]["capabilities"] = json!(["unknown.capability"]);
    let error = validate_manifest_compatibility(&unsupported).unwrap_err();
    assert!(error.to_string().contains("[P009]"));
}

#[test]
fn doctor_has_stable_order_and_shape() {
    let checks = doctor::run_doctor(Path::new(env!("CARGO_MANIFEST_DIR")), None);
    assert_eq!(checks.len(), 9);
    assert_eq!(checks.first().map(|check| check.id.as_str()), Some("tofu"));
    assert_eq!(
        checks.last().map(|check| check.id.as_str()),
        Some("github-workflows")
    );
    let serialized = serde_json::to_value(&checks[0]).unwrap();
    assert_eq!(serialized.as_object().map(serde_json::Map::len), Some(5));
}
