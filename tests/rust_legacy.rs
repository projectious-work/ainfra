//! Read-only Python-prototype recovery inspection coverage.

#![allow(clippy::unwrap_used)]

use std::fs;
use std::path::Path;

use ainfra::legacy::inspect;
use ainfra::plan_record::file_hash;
use jsonschema::Draft;
use serde_json::{Value, json};

const RUN_ID: &str = "0123456789abcdefabcd";

fn legacy_record(root: &Path) -> std::path::PathBuf {
    let run = root.join(".ainfra/runs").join(RUN_ID);
    fs::create_dir_all(&run).unwrap();
    let plan = run.join("apply.tfplan");
    fs::write(&plan, b"legacy-plan").unwrap();
    let record = json!({
        "id": RUN_ID,
        "operation": "apply",
        "template": "hetzner-kubernetes-baseline",
        "template_version": "0.1.0",
        "environment": "development",
        "input_path": root.join("input.json"),
        "input_sha256": "0".repeat(64),
        "template_sha256": "1".repeat(64),
        "plan_path": plan,
        "plan_sha256": file_hash(&plan).unwrap(),
    });
    fs::write(
        run.join("plan.json"),
        serde_json::to_vec_pretty(&record).unwrap(),
    )
    .unwrap();
    run
}

#[test]
fn python_record_is_reported_as_unknown_legacy_evidence() {
    let directory = tempfile::tempdir().unwrap();
    legacy_record(directory.path());

    let report = inspect(directory.path()).unwrap();

    assert_eq!(report.state, "legacy");
    assert_eq!(report.legacy_records, 1);
    assert_eq!(report.corrupt_entries, 0);
    assert!(report.requires_new_project_plan);
    assert_eq!(report.runs[0].id, RUN_ID);
    assert_eq!(report.runs[0].plan_integrity, "valid");
    assert_eq!(report.runs[0].lifecycle_state, "unknown");
    let schema: Value =
        serde_json::from_str(include_str!("../schemas/legacy-inspection.v1alpha1.json")).unwrap();
    let validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .build(&schema)
        .unwrap();
    assert!(validator.is_valid(&serde_json::to_value(report).unwrap()));
}

#[test]
fn changed_or_external_legacy_plans_never_become_authority() {
    let directory = tempfile::tempdir().unwrap();
    let run = legacy_record(directory.path());
    fs::write(run.join("apply.tfplan"), b"changed").unwrap();
    let report = inspect(directory.path()).unwrap();
    assert_eq!(report.state, "corrupt");
    assert_eq!(report.runs[0].plan_integrity, "mismatch");
    assert_eq!(report.runs[0].lifecycle_state, "unknown");

    let record_path = run.join("plan.json");
    let mut record: Value = serde_json::from_slice(&fs::read(&record_path).unwrap()).unwrap();
    record["plan_path"] = Value::String("/tmp/fixture-secret.tfplan".to_owned());
    fs::write(record_path, serde_json::to_vec_pretty(&record).unwrap()).unwrap();
    let report = inspect(directory.path()).unwrap();
    let serialized = serde_json::to_string(&report).unwrap();
    assert_eq!(report.state, "corrupt");
    assert_eq!(report.legacy_records, 0);
    assert!(!serialized.contains("fixture-secret"));
}

#[test]
fn arbitrary_legacy_fields_cannot_leak_through_the_report() {
    let directory = tempfile::tempdir().unwrap();
    let run = legacy_record(directory.path());
    let record_path = run.join("plan.json");
    let mut record: Value = serde_json::from_slice(&fs::read(&record_path).unwrap()).unwrap();
    record["environment"] = Value::String("HCLOUD_TOKEN=fixture-secret".to_owned());
    fs::write(record_path, serde_json::to_vec_pretty(&record).unwrap()).unwrap();

    let report = inspect(directory.path()).unwrap();
    let serialized = serde_json::to_string(&report).unwrap();

    assert_eq!(report.state, "corrupt");
    assert_eq!(report.legacy_records, 0);
    assert!(!serialized.contains("fixture-secret"));
    assert!(!serialized.contains("HCLOUD_TOKEN"));
}

#[test]
fn missing_plan_evidence_makes_the_top_level_report_corrupt() {
    let directory = tempfile::tempdir().unwrap();
    let run = legacy_record(directory.path());
    fs::remove_file(run.join("apply.tfplan")).unwrap();

    let report = inspect(directory.path()).unwrap();

    assert_eq!(report.state, "corrupt");
    assert_eq!(report.runs[0].plan_integrity, "missing");
    assert_eq!(report.runs[0].lifecycle_state, "unknown");
}

#[cfg(unix)]
#[test]
fn redirected_legacy_files_are_sanitized_as_corrupt() {
    use std::os::unix::fs::symlink;

    let directory = tempfile::tempdir().unwrap();
    let run = legacy_record(directory.path());
    let record = run.join("plan.json");
    let outside = directory.path().join("fixture-secret.json");
    fs::rename(&record, &outside).unwrap();
    symlink(&outside, &record).unwrap();

    let report = inspect(directory.path()).unwrap();
    let serialized = serde_json::to_string(&report).unwrap();

    assert_eq!(report.state, "corrupt");
    assert_eq!(report.corrupt_entries, 1);
    assert!(!serialized.contains("fixture-secret"));
}

#[cfg(unix)]
#[test]
fn redirected_operational_state_parent_is_rejected() {
    use std::os::unix::fs::symlink;

    let directory = tempfile::tempdir().unwrap();
    let outside = tempfile::tempdir().unwrap();
    legacy_record(outside.path());
    symlink(
        outside.path().join(".ainfra"),
        directory.path().join(".ainfra"),
    )
    .unwrap();

    let error = inspect(directory.path()).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("non-symlink directory"));
}

#[test]
fn missing_legacy_root_is_a_read_only_empty_report() {
    let directory = tempfile::tempdir().unwrap();

    let report = inspect(directory.path()).unwrap();

    assert_eq!(report.state, "none");
    assert_eq!(report.legacy_records, 0);
    assert!(!directory.path().join(".ainfra").exists());
}
