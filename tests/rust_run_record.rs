//! Durable immutable run identity and append-only event coverage.

#![allow(clippy::unwrap_used)]

use std::fs;
use std::path::Path;

use ainfra::plan_record::{Operation, PlanRecord};
use ainfra::run_record::{RecoveryCategory, RunOutcome, RunRecord, RunStage};
use jsonschema::Draft;
use serde_json::Value;

fn prepared_plan(root: &Path) -> PlanRecord {
    let input = root.join("input.json");
    let template = root.join("template");
    fs::write(&input, b"{\"fixture\":true}").unwrap();
    fs::create_dir(&template).unwrap();
    fs::write(template.join("main.tf"), b"resource").unwrap();
    let mut plan = PlanRecord::create(
        root,
        Operation::Apply,
        "hetzner-kubernetes-baseline",
        "0.1.0",
        "development",
        &input,
        &template,
    )
    .unwrap();
    let plan_path = std::path::PathBuf::from(&plan.plan_path);
    fs::create_dir_all(plan_path.parent().unwrap()).unwrap();
    fs::write(&plan_path, b"fake-plan").unwrap();
    plan.bind_plan(root).unwrap();
    plan.write(root).unwrap();
    plan
}

#[test]
fn run_identity_and_events_conform_to_their_schemas() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    let record = RunRecord::create(directory.path(), &plan).unwrap();
    let run_schema: Value =
        serde_json::from_str(include_str!("../schemas/run-record.v1alpha2.json")).unwrap();
    let event_schema: Value =
        serde_json::from_str(include_str!("../schemas/run-event.v1alpha1.json")).unwrap();
    let run_validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .build(&run_schema)
        .unwrap();
    let event_validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .build(&event_schema)
        .unwrap();

    assert!(run_validator.is_valid(&serde_json::to_value(&record).unwrap()));
    let events = record.events(directory.path()).unwrap();
    assert_eq!(events.len(), 1);
    assert_eq!(events[0].stage, RunStage::Planned);
    assert_eq!(events[0].outcome, RunOutcome::Succeeded);
    assert!(event_validator.is_valid(&serde_json::to_value(&events[0]).unwrap()));
}

#[test]
fn event_history_is_contiguous_and_failures_can_be_retried() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    let record = RunRecord::create(directory.path(), &plan).unwrap();
    assert!(record.start(directory.path(), RunStage::Apply).unwrap());
    record
        .fail(
            directory.path(),
            RunStage::Apply,
            "AINFRA-E500",
            RecoveryCategory::InspectInfrastructureState,
        )
        .unwrap();
    assert!(record.start(directory.path(), RunStage::Apply).unwrap());
    record.succeed(directory.path(), RunStage::Apply).unwrap();
    assert!(!record.start(directory.path(), RunStage::Apply).unwrap());
    assert!(record.start(directory.path(), RunStage::Output).unwrap());

    let events = record.events(directory.path()).unwrap();
    assert_eq!(events.len(), 6);
    assert_eq!(events.last().unwrap().outcome, RunOutcome::Started);
}

#[test]
fn interrupted_stage_cannot_be_started_again_automatically() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    let record = RunRecord::create(directory.path(), &plan).unwrap();
    record.start(directory.path(), RunStage::Apply).unwrap();

    let error = record.start(directory.path(), RunStage::Apply).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("manual recovery"));
    assert_eq!(record.events(directory.path()).unwrap().len(), 2);
}

#[test]
fn gaps_and_plan_mismatches_are_rejected() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    let record = RunRecord::create(directory.path(), &plan).unwrap();
    record.start(directory.path(), RunStage::Apply).unwrap();
    let events = directory
        .path()
        .join(".ainfra/runs")
        .join(&plan.id)
        .join("events");
    fs::rename(
        events.join("000002-apply-started.json"),
        events.join("000003-apply-started.json"),
    )
    .unwrap();
    let error = record.events(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("run event"));

    let run_path = directory
        .path()
        .join(".ainfra/runs")
        .join(&plan.id)
        .join("run.json");
    let mut document: Value =
        serde_json::from_str(&fs::read_to_string(&run_path).unwrap()).unwrap();
    document["plan_sha256"] = Value::String("0".repeat(64));
    fs::write(&run_path, serde_json::to_vec_pretty(&document).unwrap()).unwrap();
    let error = RunRecord::load(directory.path(), &plan.id).unwrap_err();
    assert!(error.to_string().contains("does not match"));
}

#[test]
fn arbitrary_failure_codes_are_rejected_as_corrupt() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    let record = RunRecord::create(directory.path(), &plan).unwrap();
    record.start(directory.path(), RunStage::Apply).unwrap();
    record
        .fail(
            directory.path(),
            RunStage::Apply,
            "AINFRA-E500",
            RecoveryCategory::InspectInfrastructureState,
        )
        .unwrap();
    let event = directory
        .path()
        .join(".ainfra/runs")
        .join(&plan.id)
        .join("events/000003-apply-failed.json");
    let mut document: Value = serde_json::from_slice(&fs::read(&event).unwrap()).unwrap();
    document["error_code"] = Value::String("fixture-secret".to_owned());
    fs::write(&event, serde_json::to_vec_pretty(&document).unwrap()).unwrap();

    let error = record.events(directory.path()).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(!error.to_string().contains("fixture-secret"));
}

#[test]
fn zero_run_creation_timestamp_is_rejected() {
    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    RunRecord::create(directory.path(), &plan).unwrap();
    let path = directory
        .path()
        .join(".ainfra/runs")
        .join(&plan.id)
        .join("run.json");
    let mut document: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    document["created_unix_nanos"] = Value::from(0);
    fs::write(&path, serde_json::to_vec_pretty(&document).unwrap()).unwrap();

    let error = RunRecord::load(directory.path(), &plan.id).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("timestamp"));
}

#[cfg(unix)]
#[test]
fn symlinked_run_identity_is_rejected() {
    use std::os::unix::fs::symlink;

    let directory = tempfile::tempdir().unwrap();
    let plan = prepared_plan(directory.path());
    RunRecord::create(directory.path(), &plan).unwrap();
    let run = directory.path().join(".ainfra/runs").join(&plan.id);
    let path = run.join("run.json");
    let real = run.join("real-run.json");
    fs::rename(&path, &real).unwrap();
    symlink(&real, &path).unwrap();

    let error = RunRecord::load(directory.path(), &plan.id).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("regular non-symlink"));
}
