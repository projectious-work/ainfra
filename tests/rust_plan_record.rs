//! Security invariants for versioned reviewed-plan records.

#![allow(clippy::unwrap_used)]

use std::fs;
use std::path::Path;

use ainfra::plan_record::{ExpectedBindings, FORMAT_VERSION, Operation, PlanRecord, template_hash};
use jsonschema::Draft;
use serde_json::Value;

fn fixture_tree(root: &Path) {
    fs::create_dir_all(root.join("tofu/.terraform")).unwrap();
    fs::create_dir_all(root.join(".ainfra")).unwrap();
    fs::write(root.join("ainfra-template.yaml"), b"manifest").unwrap();
    fs::write(root.join("tofu/main.tf"), b"resource").unwrap();
    fs::write(root.join("tofu/.terraform/ignored"), b"cache").unwrap();
    fs::write(root.join(".ainfra/ignored"), b"output").unwrap();
}

fn prepared_record(root: &Path) -> (PlanRecord, std::path::PathBuf, std::path::PathBuf) {
    let input = root.join("input.json");
    let template = root.join("template");
    fs::write(&input, b"{\"fixture\":true}").unwrap();
    fixture_tree(&template);
    let mut record = PlanRecord::create(
        root,
        Operation::Apply,
        "hetzner-kubernetes-baseline",
        "0.1.0",
        "development",
        &input,
        &template,
    )
    .unwrap();
    let plan = std::path::PathBuf::from(&record.plan_path);
    fs::create_dir_all(plan.parent().unwrap()).unwrap();
    fs::write(&plan, b"fake-plan").unwrap();
    record.bind_plan(root).unwrap();
    (record, input, template)
}

#[test]
fn reads_the_normal_python_record_shape() {
    let fixture = include_str!("compat/legacy-state/python-plan-record.v1.json");
    let record: PlanRecord = serde_json::from_str(fixture).unwrap();
    assert_eq!(record.format_version, FORMAT_VERSION);
    assert_eq!(record.operation, Operation::Apply);
    assert_eq!(record.template, "hetzner-kubernetes-baseline");
}

#[test]
fn writes_a_versioned_record_without_overwrite() {
    let directory = tempfile::tempdir().unwrap();
    let (record, _, _) = prepared_record(directory.path());
    let path = record.write(directory.path()).unwrap();
    let document: Value = serde_json::from_str(&fs::read_to_string(path).unwrap()).unwrap();
    assert_eq!(document["format_version"], FORMAT_VERSION);
    assert_eq!(document.as_object().map(serde_json::Map::len), Some(11));
    assert!(record.write(directory.path()).is_err());
    assert_eq!(
        PlanRecord::load(directory.path(), &record.id).unwrap(),
        record
    );
}

#[test]
fn rejects_unsafe_ids_and_plan_paths() {
    let directory = tempfile::tempdir().unwrap();
    assert!(PlanRecord::load(directory.path(), "../outside").is_err());

    let (mut record, _, _) = prepared_record(directory.path());
    let outside = directory.path().join("outside.tfplan");
    fs::write(&outside, b"fake-plan").unwrap();
    record.plan_path = outside.display().to_string();
    let error = record.bind_plan(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("plan path"));
}

#[test]
fn verification_rejects_stale_and_cross_operation_plans() {
    let directory = tempfile::tempdir().unwrap();
    let (record, input, template) = prepared_record(directory.path());
    let expected = |operation| ExpectedBindings {
        template: "hetzner-kubernetes-baseline",
        template_version: "0.1.0",
        environment: "development",
        input_path: &input,
        template_root: &template,
        operation,
        backend_config_path: None,
        backend_config_sha256: None,
    };
    record
        .verify(directory.path(), &expected(Operation::Apply))
        .unwrap();

    let error = record
        .verify(directory.path(), &expected(Operation::Destroy))
        .unwrap_err();
    assert!(error.to_string().contains("cannot authorize"));

    fs::write(&input, b"{\"fixture\":false}").unwrap();
    let error = record
        .verify(directory.path(), &expected(Operation::Apply))
        .unwrap_err();
    assert!(error.to_string().contains("input content changed"));
}

#[test]
fn verification_rejects_modified_plan_and_template_bytes() {
    let directory = tempfile::tempdir().unwrap();
    let (record, input, template) = prepared_record(directory.path());
    let expected = || ExpectedBindings {
        template: "hetzner-kubernetes-baseline",
        template_version: "0.1.0",
        environment: "development",
        input_path: &input,
        template_root: &template,
        operation: Operation::Apply,
        backend_config_path: None,
        backend_config_sha256: None,
    };

    fs::write(template.join("tofu/main.tf"), b"changed").unwrap();
    let error = record.verify(directory.path(), &expected()).unwrap_err();
    assert!(error.to_string().contains("template content changed"));
    fs::write(template.join("tofu/main.tf"), b"resource").unwrap();

    fs::write(&record.plan_path, b"changed").unwrap();
    let error = record.verify(directory.path(), &expected()).unwrap_err();
    assert!(error.to_string().contains("plan file was modified"));
}

#[test]
fn template_hash_is_deterministic_and_excludes_runtime_state() {
    let directory = tempfile::tempdir().unwrap();
    fixture_tree(directory.path());
    let initial = template_hash(directory.path()).unwrap();
    fs::write(
        directory.path().join("tofu/.terraform/ignored"),
        b"changed cache",
    )
    .unwrap();
    fs::write(directory.path().join(".ainfra/ignored"), b"changed output").unwrap();
    assert_eq!(template_hash(directory.path()).unwrap(), initial);
}

#[test]
fn template_hash_matches_the_python_oracle() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("templates/hetzner-kubernetes-baseline");
    assert_eq!(
        template_hash(&root).unwrap(),
        "d25f6ce5fc94670b0c47fb108542de3b2beba890bf9f75ac37170e32c0d96490"
    );
}

#[test]
fn new_records_conform_to_the_versioned_schema() {
    let directory = tempfile::tempdir().unwrap();
    let (record, _, _) = prepared_record(directory.path());
    let schema: Value =
        serde_json::from_str(include_str!("../schemas/plan-record.v1alpha1.json")).unwrap();
    let validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .build(&schema)
        .unwrap();
    let document = serde_json::to_value(record).unwrap();
    assert!(validator.is_valid(&document));
}
