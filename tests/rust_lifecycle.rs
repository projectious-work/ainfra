//! Fake-runner integration coverage for isolated Rust planning.

#![allow(clippy::unwrap_used)]

use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::Mutex;
use std::time::Duration;

use ainfra::error::AinfraError;
use ainfra::lifecycle::{EnvironmentSource, apply, destroy, plan};
use ainfra::plan_record::{Operation, PlanRecord, template_hash};
use ainfra::process::{
    ProcessRequest, ProcessResult, Runner, SubprocessRunner, child_environment, redact,
    require_success,
};

struct FixtureEnvironment(BTreeMap<String, String>);

impl EnvironmentSource for FixtureEnvironment {
    fn get(&self, name: &str) -> Option<String> {
        self.0.get(name).cloned()
    }
}

type RecordedCall = (Vec<String>, PathBuf, BTreeMap<String, String>);

#[derive(Default)]
struct FakeRunner {
    calls: Mutex<Vec<RecordedCall>>,
    fail_on_call: Option<usize>,
}

impl Runner for FakeRunner {
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError> {
        let call_number = {
            let mut calls = self.calls.lock().unwrap();
            calls.push((
                request.argv.clone(),
                request.cwd.clone(),
                request.environment.clone(),
            ));
            calls.len()
        };
        if let Some(path) = request
            .argv
            .iter()
            .find_map(|value| value.strip_prefix("-out="))
        {
            fs::write(path, b"fake-plan").unwrap();
        }
        let failed = self.fail_on_call == Some(call_number);
        Ok(ProcessResult {
            argv: request.argv.clone(),
            return_code: i32::from(failed),
            stdout: "ok\n".to_owned(),
            stderr: if failed {
                redact("fixture-secret failed\n", &request.secrets)
            } else {
                String::new()
            },
        })
    }
}

fn input() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/contracts/v1alpha1/valid/template-input.json")
}

fn environment() -> FixtureEnvironment {
    FixtureEnvironment(BTreeMap::from([(
        "HCLOUD_TOKEN".to_owned(),
        "fixture-secret".to_owned(),
    )]))
}

#[test]
fn plan_materializes_an_isolated_workspace_and_exact_argv() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap();
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 2);
    assert_eq!(
        &calls[0].0[..5],
        [
            "tofu",
            "init",
            "-input=false",
            "-lockfile=readonly",
            "-reconfigure"
        ]
    );
    assert_eq!(&calls[1].0[..3], ["tofu", "plan", "-input=false"]);
    assert!(!calls[1].0.contains(&"-destroy".to_owned()));
    assert!(calls[1].0.iter().any(|value| value.starts_with("-out=")));
    assert!(
        calls[1]
            .0
            .iter()
            .any(|value| value.starts_with("-var-file="))
    );
    assert_eq!(calls[0].1, calls[1].1);
    assert!(calls[0].1.ends_with("workspace/tofu"));
    assert_eq!(
        calls[0].2.get("HCLOUD_TOKEN").map(String::as_str),
        Some("fixture-secret")
    );
    assert!(!calls[0].2.contains_key("CARGO_MANIFEST_DIR"));
    drop(calls);

    let run = Path::new(&record.plan_path).parent().unwrap();
    let workspace = run.join("workspace");
    assert!(workspace.join("tofu/providers.tf").is_file());
    assert!(workspace.join("ansible/site.yml").is_file());
    assert_eq!(template_hash(&workspace).unwrap(), record.template_sha256);
    assert!(run.join("input.auto.tfvars.json").is_file());
    assert!(run.join("plan.json").is_file());
    assert_eq!(
        PlanRecord::load(project.path(), &record.id).unwrap(),
        record
    );
}

#[test]
fn destroy_plan_uses_only_the_plan_destroy_flag() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Destroy,
    )
    .unwrap();
    let calls = runner.calls.lock().unwrap();
    assert!(calls[1].0.contains(&"-destroy".to_owned()));
    assert!(record.plan_path.ends_with("destroy.tfplan"));
    assert_eq!(record.operation, Operation::Destroy);
}

#[test]
fn failed_init_never_runs_plan_or_writes_a_record() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: Some(1),
    };
    let error = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");
    assert_eq!(runner.calls.lock().unwrap().len(), 1);
    let records: Vec<_> = fs::read_dir(project.path().join(".ainfra/runs"))
        .unwrap()
        .filter_map(Result::ok)
        .filter(|entry| entry.path().join("plan.json").exists())
        .collect();
    assert!(records.is_empty());
}

#[test]
fn process_failures_and_redaction_keep_the_dependency_boundary() {
    let secret = "do-not-leak".to_owned();
    assert_eq!(
        redact("prefix do-not-leak suffix", std::slice::from_ref(&secret)),
        "prefix [REDACTED] suffix"
    );
    let error = require_success(ProcessResult {
        argv: vec!["tofu".to_owned()],
        return_code: 7,
        stdout: String::new(),
        stderr: redact("do-not-leak failed", &[secret]),
    })
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");
    assert!(!error.to_string().contains("do-not-leak"));
}

#[test]
fn real_runner_bounds_output_and_reports_missing_executables() {
    let root = tempfile::tempdir().unwrap();
    let runner = SubprocessRunner::new(Duration::from_secs(1), 8);
    let result = runner
        .run(&ProcessRequest {
            argv: vec!["/usr/bin/printf".to_owned(), "abcdefghijklmnop".to_owned()],
            cwd: root.path().to_path_buf(),
            environment: child_environment(),
            secrets: Vec::new(),
        })
        .unwrap();
    assert!(result.stdout.starts_with("abcdefgh"));
    assert!(result.stdout.contains("[OUTPUT TRUNCATED]"));

    let error = runner
        .run(&ProcessRequest {
            argv: vec!["/ainfra-fixture-does-not-exist".to_owned()],
            cwd: root.path().to_path_buf(),
            environment: child_environment(),
            secrets: Vec::new(),
        })
        .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");

    let timeout = SubprocessRunner::new(Duration::from_millis(10), 8);
    let error = timeout
        .run(&ProcessRequest {
            argv: vec!["/usr/bin/sleep".to_owned(), "1".to_owned()],
            cwd: root.path().to_path_buf(),
            environment: child_environment(),
            secrets: Vec::new(),
        })
        .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");
    assert!(error.to_string().contains("timed out"));
}

#[test]
fn truncated_secret_fragments_are_not_exposed() {
    let secret = "fixture-secret".to_owned();
    let output = redact(
        "prefix fixture-\n[OUTPUT TRUNCATED]",
        std::slice::from_ref(&secret),
    );
    assert!(!output.contains("fixture-"));
    assert!(output.contains("[REDACTED]"));
}

#[test]
fn apply_rechecks_bindings_then_uses_the_exact_plan() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap();
    runner.calls.lock().unwrap().clear();
    let result = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    assert_eq!(result.return_code, 0);
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 2);
    assert_eq!(
        calls[1].0,
        ["tofu", "apply", "-input=false", record.plan_path.as_str()]
    );
    assert!(calls[1].1.ends_with("workspace/tofu"));
}

#[test]
fn destroy_requires_a_destroy_plan_and_uses_tofu_apply() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner::default();
    let apply_record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap();
    runner.calls.lock().unwrap().clear();
    let error = destroy(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &apply_record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("cannot authorize"));
    assert!(runner.calls.lock().unwrap().is_empty());

    let destroy_record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Destroy,
    )
    .unwrap();
    runner.calls.lock().unwrap().clear();
    destroy(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &destroy_record.id,
    )
    .unwrap();
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls[1].0[1], "apply");
    assert_eq!(calls[1].0[3], destroy_record.plan_path);
}

#[test]
fn tampered_plan_is_rejected_before_init_or_apply() {
    let project = tempfile::tempdir().unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap();
    fs::write(&record.plan_path, b"tampered").unwrap();
    runner.calls.lock().unwrap().clear();
    let error = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("modified"));
    assert!(runner.calls.lock().unwrap().is_empty());
}

#[test]
fn changed_input_and_workspace_are_rejected_before_execution() {
    let project = tempfile::tempdir().unwrap();
    let copied_input = project.path().join("input.json");
    fs::copy(input(), &copied_input).unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &copied_input,
        Operation::Apply,
    )
    .unwrap();
    runner.calls.lock().unwrap().clear();

    let mut changed = fs::read(&copied_input).unwrap();
    changed.push(b'\n');
    fs::write(&copied_input, changed).unwrap();
    let error = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &copied_input,
        &record.id,
    )
    .unwrap_err();
    assert!(error.to_string().contains("input content changed"));
    assert!(runner.calls.lock().unwrap().is_empty());

    fs::copy(input(), &copied_input).unwrap();
    let workspace = Path::new(&record.plan_path)
        .parent()
        .unwrap()
        .join("workspace/tofu/providers.tf");
    fs::write(workspace, b"changed").unwrap();
    let error = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &copied_input,
        &record.id,
    )
    .unwrap_err();
    assert!(error.to_string().contains("template content changed"));
    assert!(runner.calls.lock().unwrap().is_empty());
}

#[test]
fn changed_backend_configuration_is_rejected_before_init() {
    let project = tempfile::tempdir().unwrap();
    let input = project.path().join("remote-input.json");
    fs::copy(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("tests/fixtures/policy/v1alpha1/valid/remote-input.json"),
        &input,
    )
    .unwrap();
    let backend = project.path().join(".ainfra/backend.hcl");
    fs::create_dir_all(backend.parent().unwrap()).unwrap();
    fs::write(&backend, b"bucket = \"reviewed\"\n").unwrap();
    let runner = FakeRunner::default();
    let record = plan(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input,
        Operation::Apply,
    )
    .unwrap();
    assert_eq!(
        record.backend_config_path.as_deref(),
        Some(backend.canonicalize().unwrap().to_str().unwrap())
    );
    assert!(record.backend_config_sha256.is_some());

    fs::write(&backend, b"bucket = \"changed\"\n").unwrap();
    runner.calls.lock().unwrap().clear();
    let error = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input,
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("backend configuration changed"));
    assert!(runner.calls.lock().unwrap().is_empty());
}

#[test]
fn init_and_apply_failures_stop_safely_with_redacted_errors() {
    let project = tempfile::tempdir().unwrap();
    let planning_runner = FakeRunner::default();
    let record = plan(
        &planning_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        Operation::Apply,
    )
    .unwrap();

    let init_failure = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: Some(1),
    };
    let error = apply(
        &init_failure,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");
    assert_eq!(init_failure.calls.lock().unwrap().len(), 1);
    assert!(!error.to_string().contains("fixture-secret"));

    let apply_failure = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: Some(2),
    };
    let error = apply(
        &apply_failure,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E500");
    assert_eq!(apply_failure.calls.lock().unwrap().len(), 2);
    assert!(!error.to_string().contains("fixture-secret"));
    assert!(error.to_string().contains("[REDACTED]"));
}
