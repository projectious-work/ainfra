//! Fake-runner integration coverage for isolated Rust planning.

#![allow(clippy::unwrap_used)]

use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::Mutex;
use std::time::Duration;

use ainfra::error::AinfraError;
use ainfra::lifecycle::{
    EnvironmentSource, apply, apply_project, collect_output, collect_output_project, configure,
    configure_project, destroy, plan, plan_project,
};
use ainfra::plan_record::{Operation, PROJECT_FORMAT_VERSION, PlanRecord, template_hash};
use ainfra::process::{
    ProcessRequest, ProcessResult, Runner, SubprocessRunner, child_environment, redact,
    require_success,
};
use ainfra::project::{Project, initialize};
use ainfra::run_record::{RecoveryCategory, RunOutcome, RunRecord, RunStage};
use ainfra::status::inspect;
use jsonschema::Draft;

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
    tofu_output: Option<String>,
    tamper_record: Option<PathBuf>,
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
        let is_output = request.argv.get(1).map(String::as_str) == Some("output");
        if is_output && let Some(path) = &self.tamper_record {
            let mut record: serde_json::Value =
                serde_json::from_slice(&fs::read(path).unwrap()).unwrap();
            record["plan_path"] = serde_json::json!("/tmp/attacker/plan.tfplan");
            fs::write(path, serde_json::to_vec_pretty(&record).unwrap()).unwrap();
        }
        Ok(ProcessResult {
            argv: request.argv.clone(),
            return_code: i32::from(failed),
            stdout: if is_output {
                self.tofu_output.clone().unwrap_or_default()
            } else {
                "ok\n".to_owned()
            },
            stderr: if failed {
                redact("fixture-secret failed\n", &request.secrets)
            } else {
                String::new()
            },
        })
    }
}

struct ProjectTamperingRunner {
    calls: Mutex<Vec<RecordedCall>>,
    config_path: PathBuf,
}

impl Runner for ProjectTamperingRunner {
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError> {
        self.calls.lock().unwrap().push((
            request.argv.clone(),
            request.cwd.clone(),
            request.environment.clone(),
        ));
        if request.argv.get(1).map(String::as_str) == Some("init") {
            fs::write(&self.config_path, "apiVersion: changed\n").unwrap();
        }
        Ok(ProcessResult {
            argv: request.argv.clone(),
            return_code: 0,
            stdout: "ok\n".to_owned(),
            stderr: String::new(),
        })
    }
}

struct ProjectOutputTamperingRunner {
    calls: Mutex<Vec<RecordedCall>>,
    config_path: PathBuf,
}

struct InitThenErrorRunner {
    calls: Mutex<Vec<RecordedCall>>,
}

impl Runner for InitThenErrorRunner {
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError> {
        let mut calls = self.calls.lock().unwrap();
        calls.push((
            request.argv.clone(),
            request.cwd.clone(),
            request.environment.clone(),
        ));
        if calls.len() == 1 {
            return Ok(ProcessResult {
                argv: request.argv.clone(),
                return_code: 0,
                stdout: "ok\n".to_owned(),
                stderr: String::new(),
            });
        }
        Err(AinfraError::dependency(
            "fixture-secret must not enter durable events",
        ))
    }
}

impl Runner for ProjectOutputTamperingRunner {
    fn run(&self, request: &ProcessRequest) -> Result<ProcessResult, AinfraError> {
        self.calls.lock().unwrap().push((
            request.argv.clone(),
            request.cwd.clone(),
            request.environment.clone(),
        ));
        if request.argv.get(1).map(String::as_str) == Some("output") {
            fs::write(&self.config_path, "apiVersion: changed\n").unwrap();
        }
        Ok(ProcessResult {
            argv: request.argv.clone(),
            return_code: 0,
            stdout: tofu_output(),
            stderr: String::new(),
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

fn tofu_output() -> String {
    serde_json::json!({
        "inventory_nodes": {
            "sensitive": false,
            "type": ["tuple"],
            "value": [{
                "name": "ainfra-development-control-01",
                "role": "control-plane-capable",
                "private_ipv4": "10.42.0.10",
                "public_ipv4": null,
                "public_ipv6": null,
                "image": "debian-13",
            }],
        },
        "ownership": {
            "sensitive": false,
            "type": ["map", "string"],
            "value": {
                "managed-by": "ainfra",
                "template": "hetzner-kubernetes-baseline",
                "environment": "development",
            },
        },
    })
    .to_string()
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
        tofu_output: None,
        tamper_record: None,
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
        tofu_output: None,
        tamper_record: None,
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
        tofu_output: None,
        tamper_record: None,
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

#[test]
fn output_requires_apply_and_persists_a_valid_run_artifact() {
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
    let error = collect_output(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(runner.calls.lock().unwrap().is_empty());

    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let output_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };
    let output = collect_output(
        &output_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
    )
    .unwrap();
    assert_eq!(output["kind"], "InfrastructureOutput");
    assert_eq!(output["metadata"]["runId"], record.id);
    assert_eq!(output["spec"]["nodes"][0]["privateIPv4"], "10.42.0.10");
    let run = Path::new(&record.plan_path).parent().unwrap();
    assert!(run.join("output.json").is_file());
    assert_eq!(output_runner.calls.lock().unwrap()[0].0[1], "output");
}

#[test]
fn sensitive_output_fails_without_an_artifact() {
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
    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let mut raw: serde_json::Value = serde_json::from_str(&tofu_output()).unwrap();
    raw["inventory_nodes"]["sensitive"] = serde_json::json!(true);
    let output_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(raw.to_string()),
        tamper_record: None,
    };
    let error = collect_output(
        &output_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E300");
    assert!(
        !Path::new(&record.plan_path)
            .parent()
            .unwrap()
            .join("output.json")
            .exists()
    );
    let events = RunRecord::load(project.path(), &record.id)
        .unwrap()
        .events(project.path())
        .unwrap();
    let failure = events.last().unwrap();
    assert_eq!(failure.outcome, RunOutcome::Failed);
    assert_eq!(failure.recovery, Some(RecoveryCategory::InspectOutputState));
    assert!(
        !serde_json::to_string(&events)
            .unwrap()
            .contains("fixture-secret")
    );
}

#[test]
fn runner_errors_after_stage_start_become_sanitized_failures() {
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
    let apply_runner = InitThenErrorRunner {
        calls: Mutex::new(Vec::new()),
    };

    apply(
        &apply_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap_err();
    let events = RunRecord::load(project.path(), &record.id)
        .unwrap()
        .events(project.path())
        .unwrap();
    let failure = events.last().unwrap();
    assert_eq!(failure.outcome, RunOutcome::Failed);
    assert_eq!(failure.error_code.as_deref(), Some("AINFRA-E500"));
    assert_eq!(
        failure.recovery,
        Some(RecoveryCategory::InspectInfrastructureState)
    );
    assert!(
        !serde_json::to_string(&events)
            .unwrap()
            .contains("fixture-secret")
    );
}

#[cfg(unix)]
#[test]
fn configure_uses_validated_inventory_and_no_provider_secret() {
    use std::os::unix::fs::PermissionsExt;

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
    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o600)).unwrap();
    let configure_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };
    configure(
        &configure_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
        &known_hosts,
        true,
    )
    .unwrap();
    let calls = configure_runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 2);
    assert_eq!(calls[0].0[1], "output");
    assert_eq!(calls[1].0[0], "ansible-playbook");
    assert_eq!(calls[1].0[1], "--inventory");
    assert!(calls[1].0.contains(&"--check".to_owned()));
    assert!(calls[1].0.contains(&"--diff".to_owned()));
    assert!(calls[1].1.ends_with("workspace/ansible"));
    assert!(!calls[1].2.contains_key("HCLOUD_TOKEN"));
    assert_eq!(
        calls[1]
            .2
            .get("ANSIBLE_HOST_KEY_CHECKING")
            .map(String::as_str),
        Some("True")
    );
    assert_eq!(
        calls[1].2.get("ANSIBLE_SSH_ARGS"),
        Some(&format!(
            "-o UserKnownHostsFile={} -o StrictHostKeyChecking=yes",
            known_hosts.canonicalize().unwrap().display()
        ))
    );
    assert!(
        Path::new(&record.plan_path)
            .parent()
            .unwrap()
            .join("inventory.yml")
            .is_file()
    );
    let events = RunRecord::load(project.path(), &record.id)
        .unwrap()
        .events(project.path())
        .unwrap();
    assert!(
        !events
            .iter()
            .any(|event| event.stage == RunStage::Configure)
    );
    assert_eq!(events.last().unwrap().stage, RunStage::Output);
}

#[test]
fn configure_requires_verified_host_keys_before_any_runner_call() {
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
    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let configure_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };
    let error = configure(
        &configure_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
        &project.path().join("missing-known-hosts"),
        false,
    )
    .unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(configure_runner.calls.lock().unwrap().is_empty());
}

#[cfg(unix)]
#[test]
fn configure_rejects_writable_known_hosts_before_any_runner_call() {
    use std::os::unix::fs::PermissionsExt;

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
    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o620)).unwrap();
    let configure_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };

    let error = configure(
        &configure_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
        &known_hosts,
        false,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(
        error
            .to_string()
            .contains("must not be group/world writable")
    );
    assert!(configure_runner.calls.lock().unwrap().is_empty());
}

#[cfg(unix)]
#[test]
fn configure_rejects_plan_record_replacement_before_ansible() {
    use std::os::unix::fs::PermissionsExt;

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
    apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &input(),
        &record.id,
    )
    .unwrap();
    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o600)).unwrap();
    let record_path = project
        .path()
        .join(".ainfra/runs")
        .join(&record.id)
        .join("plan.json");
    let configure_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: Some(record_path),
    };

    let error = configure(
        &configure_runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &record.id,
        &known_hosts,
        false,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(
        error
            .to_string()
            .contains("reviewed plan record changed during output collection")
    );
    let calls = configure_runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 1);
    assert_eq!(calls[0].0[1], "output");
}

#[test]
fn project_lifecycle_binds_configuration_and_lock() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = FakeRunner::default();
    let record = plan_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    assert_eq!(record.format_version, PROJECT_FORMAT_VERSION);
    assert!(record.project_config_sha256.is_some());
    assert!(record.project_lock_sha256.is_some());
    let loaded_project = Project::load(project.path()).unwrap();
    let status = inspect(&loaded_project, "development").unwrap();
    assert_eq!(status.lifecycle.state, "planned");
    let schema: serde_json::Value =
        serde_json::from_str(include_str!("../schemas/status.v1alpha1.json")).unwrap();
    let validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .build(&schema)
        .unwrap();
    assert!(validator.is_valid(&serde_json::to_value(&status).unwrap()));

    apply_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap();
    let status = inspect(&loaded_project, "development").unwrap();
    assert_eq!(status.lifecycle.state, "applied");
    let output_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };
    collect_output_project(
        &output_runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap();
    assert_eq!(output_runner.calls.lock().unwrap().len(), 1);
    let status = inspect(&loaded_project, "development").unwrap();
    assert_eq!(status.lifecycle.state, "output-collected");
    let config_path = project.path().join("ainfra.yaml");
    let changed = fs::read_to_string(&config_path)
        .unwrap()
        .replace("example-infrastructure", "changed-infrastructure");
    fs::write(config_path, changed).unwrap();
    let changed_project = Project::load(project.path()).unwrap();
    let status = inspect(&changed_project, "development").unwrap();
    assert_eq!(status.lifecycle.state, "stale");
    assert_eq!(status.lifecycle.integrity, "stale");
}

#[test]
fn started_run_stage_is_reported_as_partial() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = FakeRunner::default();
    let plan = plan_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    let record = RunRecord::load(project.path(), &plan.id).unwrap();
    record.start(project.path(), RunStage::Apply).unwrap();

    let loaded_project = Project::load(project.path()).unwrap();
    let status = inspect(&loaded_project, "development").unwrap();

    assert_eq!(status.lifecycle.state, "partial");
    assert_eq!(status.latest_run.unwrap().stages.len(), 2);
    assert_eq!(status.next[0].command, "docs: lifecycle recovery");
}

#[test]
fn corrupt_run_evidence_forces_manual_recovery() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runs = project.path().join(".ainfra/runs");
    fs::create_dir_all(&runs).unwrap();
    fs::write(runs.join("untrusted-entry"), b"fixture-secret").unwrap();
    let loaded_project = Project::load(project.path()).unwrap();

    let status = inspect(&loaded_project, "development").unwrap();

    assert_eq!(status.lifecycle.state, "corrupt");
    assert_eq!(status.lifecycle.integrity, "corrupt");
    assert_eq!(status.next[0].command, "docs: lifecycle recovery");
    assert!(
        !serde_json::to_string(&status)
            .unwrap()
            .contains("fixture-secret")
    );
}

#[test]
fn corrupt_run_cannot_be_bypassed_by_an_older_valid_run() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = FakeRunner::default();
    let plan = plan_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    fs::write(
        project.path().join(".ainfra/runs/untrusted-entry"),
        b"fixture-secret",
    )
    .unwrap();
    let loaded_project = Project::load(project.path()).unwrap();

    let status = inspect(&loaded_project, "development").unwrap();

    assert_eq!(status.lifecycle.state, "corrupt");
    assert_eq!(status.latest_run.as_ref().unwrap().id, plan.id);
    assert_eq!(status.next[0].command, "docs: lifecycle recovery");
    assert!(
        !serde_json::to_string(&status)
            .unwrap()
            .contains("fixture-secret")
    );
}

#[test]
fn controlled_failures_use_only_sanitized_recovery_categories() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let planning_runner = FakeRunner::default();
    let plan = plan_project(
        &planning_runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    let apply_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: Some(2),
        tofu_output: None,
        tamper_record: None,
    };

    apply_project(
        &apply_runner,
        &environment(),
        project.path(),
        "development",
        &plan.id,
    )
    .unwrap_err();
    let events = RunRecord::load(project.path(), &plan.id)
        .unwrap()
        .events(project.path())
        .unwrap();
    let failure = events.last().unwrap();
    assert_eq!(failure.outcome, RunOutcome::Failed);
    assert_eq!(
        failure.recovery,
        Some(RecoveryCategory::InspectInfrastructureState)
    );
    assert!(
        !serde_json::to_string(&events)
            .unwrap()
            .contains("fixture-secret")
    );
}

#[test]
fn changed_project_files_fail_before_any_lifecycle_process() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let planning_runner = FakeRunner::default();
    let record = plan_project(
        &planning_runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    fs::write(project.path().join("ainfra.yaml"), "apiVersion: changed\n").unwrap();
    let apply_runner = FakeRunner::default();

    let error = apply_project(
        &apply_runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap_err();

    assert!(matches!(error.code(), "AINFRA-E200" | "AINFRA-E600"));
    assert!(apply_runner.calls.lock().unwrap().is_empty());
}

#[test]
fn explicit_mode_rejects_a_project_bound_plan() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let planning_runner = FakeRunner::default();
    let record = plan_project(
        &planning_runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    let runner = FakeRunner::default();
    let configured_input = project.path().join("environments/development.yaml");

    let error = apply(
        &runner,
        &environment(),
        project.path(),
        "hetzner-kubernetes-baseline",
        &configured_input,
        &record.id,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("requires project verification"));
    assert!(runner.calls.lock().unwrap().is_empty());
}

#[test]
fn project_change_during_init_prevents_apply_process() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let planning_runner = FakeRunner::default();
    let record = plan_project(
        &planning_runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    let runner = ProjectTamperingRunner {
        calls: Mutex::new(Vec::new()),
        config_path: project.path().join("ainfra.yaml"),
    };

    let error = apply_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(
        error
            .to_string()
            .contains("changed before process execution")
    );
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 1);
    assert_eq!(calls[0].0[1], "init");
}

#[test]
fn project_change_during_init_prevents_plan_process() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = ProjectTamperingRunner {
        calls: Mutex::new(Vec::new()),
        config_path: project.path().join("ainfra.yaml"),
    };

    let error = plan_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(
        error
            .to_string()
            .contains("changed before process execution")
    );
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 1);
    assert_eq!(calls[0].0[1], "init");
}

#[test]
fn changed_project_lock_prevents_output_process() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = FakeRunner::default();
    let record = plan_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    apply_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap();
    fs::write(project.path().join("ainfra.lock"), "apiVersion: changed\n").unwrap();
    let output_runner = FakeRunner {
        calls: Mutex::new(Vec::new()),
        fail_on_call: None,
        tofu_output: Some(tofu_output()),
        tamper_record: None,
    };

    let error = collect_output_project(
        &output_runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap_err();

    assert!(matches!(error.code(), "AINFRA-E200" | "AINFRA-E600"));
    assert!(output_runner.calls.lock().unwrap().is_empty());
}

#[test]
fn undeclared_project_environment_fails_before_process() {
    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let runner = FakeRunner::default();

    let error = plan_project(
        &runner,
        &environment(),
        project.path(),
        "production",
        Operation::Apply,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E200");
    assert!(error.to_string().contains("not declared"));
    assert!(runner.calls.lock().unwrap().is_empty());
}

#[cfg(unix)]
#[test]
fn project_change_during_output_prevents_ansible_process() {
    use std::os::unix::fs::PermissionsExt;

    let project = tempfile::tempdir().unwrap();
    initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let setup_runner = FakeRunner::default();
    let record = plan_project(
        &setup_runner,
        &environment(),
        project.path(),
        "development",
        Operation::Apply,
    )
    .unwrap();
    apply_project(
        &setup_runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
    )
    .unwrap();
    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o600)).unwrap();
    let runner = ProjectOutputTamperingRunner {
        calls: Mutex::new(Vec::new()),
        config_path: project.path().join("ainfra.yaml"),
    };

    let error = configure_project(
        &runner,
        &environment(),
        project.path(),
        "development",
        &record.id,
        &known_hosts,
        false,
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(
        error
            .to_string()
            .contains("changed before process execution")
    );
    let calls = runner.calls.lock().unwrap();
    assert_eq!(calls.len(), 1);
    assert_eq!(calls[0].0[1], "output");
}
