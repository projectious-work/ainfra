//! Local-only project lifecycle status derived from durable run evidence.

use std::fs;
use std::path::Path;

use serde::Serialize;

use crate::error::AinfraError;
use crate::plan_record::{
    Operation, PROJECT_FORMAT_VERSION, PlanRecord, ProjectBinding, file_hash, template_hash,
};
use crate::project::Project;
use crate::run_record::{RecoveryCategory, RunEvent, RunOutcome, RunRecord, RunStage};

/// Stable machine-readable status document.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StatusReport {
    /// Status output protocol.
    pub api_version: &'static str,
    /// Contract kind.
    pub kind: &'static str,
    /// Whether inspection itself completed.
    pub ok: bool,
    /// Selected project identity.
    pub project: StatusProject,
    /// Current lifecycle summary.
    pub lifecycle: LifecycleStatus,
    /// Latest valid project-bound run, when available.
    pub latest_run: Option<LatestRun>,
    /// Safe local findings.
    pub checks: Vec<StatusCheck>,
    /// Deterministic next commands.
    pub next: Vec<NextAction>,
}

/// Non-secret selected project identity.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StatusProject {
    /// Canonical root.
    pub root: String,
    /// Selected environment.
    pub environment: String,
    /// Locked template name.
    pub template: String,
    /// Locked template version.
    pub template_version: String,
}

/// Current lifecycle classification.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LifecycleStatus {
    /// Stable state spelling.
    pub state: String,
    /// Run identifier.
    pub run_id: Option<String>,
    /// Intended operation.
    pub operation: Option<Operation>,
    /// Local integrity classification.
    pub integrity: String,
    /// Latest event time as Unix nanoseconds.
    pub updated_unix_nanos: Option<u64>,
}

/// Non-secret latest run details.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LatestRun {
    /// Run identifier.
    pub id: String,
    /// Run creation time as Unix nanoseconds.
    pub created_unix_nanos: u64,
    /// Intended operation.
    pub operation: Operation,
    /// Validated stage evidence.
    pub stages: Vec<StageStatus>,
}

/// One sanitized event in status output.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StageStatus {
    /// Stage name.
    pub name: String,
    /// Outcome.
    pub state: String,
    /// Event time as Unix nanoseconds.
    pub at_unix_nanos: u64,
    /// Stable error code, if failed.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error_code: Option<String>,
    /// Sanitized recovery category, if failed.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub recovery: Option<RecoveryCategory>,
}

/// One safe local status finding.
#[derive(Debug, Serialize)]
pub struct StatusCheck {
    /// Stable finding identifier.
    pub id: String,
    /// `pass`, `warn`, or `fail`.
    pub status: String,
    /// Non-secret explanation.
    pub message: String,
}

/// One deterministic operator action.
#[derive(Debug, Serialize)]
pub struct NextAction {
    /// Exact command or documentation route.
    pub command: String,
    /// Why it is safe and relevant.
    pub reason: String,
}

struct Candidate {
    plan: PlanRecord,
    record: RunRecord,
    events: Vec<RunEvent>,
    integrity: String,
    updated: u64,
}

/// Inspect one project environment without subprocesses or credential access.
///
/// # Errors
///
/// Returns an error only when the project or selected environment cannot be
/// validated safely. Individual corrupt runs become findings.
#[allow(clippy::too_many_lines)]
pub fn inspect(project: &Project, environment: &str) -> Result<StatusReport, AinfraError> {
    let input = project.inputs.get(environment).ok_or_else(|| {
        AinfraError::input_contract(format!(
            "environment {environment:?} is not declared in ainfra.yaml"
        ))
    })?;
    let binding = ProjectBinding::capture(&project.root)?;
    let runs = project.root.join(".ainfra/runs");
    let mut candidates = Vec::new();
    let mut corrupt = 0_u64;
    let mut legacy_explicit = 0_u64;
    let mut legacy_project = 0_u64;
    if runs.exists() {
        let metadata = fs::symlink_metadata(&runs)
            .map_err(|error| AinfraError::guard(format!("cannot inspect run root: {error}")))?;
        if metadata.file_type().is_symlink() || !metadata.is_dir() {
            return Err(AinfraError::guard(
                "run root must be a regular non-symlink directory",
            ));
        }
        for entry in fs::read_dir(&runs)
            .map_err(|error| AinfraError::guard(format!("cannot list runs: {error}")))?
        {
            let entry = entry.map_err(|error| AinfraError::guard(error.to_string()))?;
            if !entry
                .file_type()
                .map_err(|error| AinfraError::guard(error.to_string()))?
                .is_dir()
            {
                corrupt += 1;
                continue;
            }
            let id = entry.file_name();
            let Some(id) = id.to_str() else {
                corrupt += 1;
                continue;
            };
            let Ok(plan) = PlanRecord::load(&project.root, id) else {
                corrupt += 1;
                continue;
            };
            if plan.environment != environment {
                continue;
            }
            if plan.format_version != PROJECT_FORMAT_VERSION {
                legacy_explicit += 1;
                continue;
            }
            if !entry.path().join("run.json").exists() {
                legacy_project += 1;
                continue;
            }
            let Ok(record) = RunRecord::load(&project.root, id) else {
                corrupt += 1;
                continue;
            };
            let Ok(events) = record.events(&project.root) else {
                corrupt += 1;
                continue;
            };
            let Some(updated) = events.last().map(|event| event.recorded_unix_nanos) else {
                corrupt += 1;
                continue;
            };
            let integrity = local_integrity(&project.root, input, &plan, &binding).to_owned();
            candidates.push(Candidate {
                plan,
                record,
                events,
                integrity,
                updated,
            });
        }
    }
    candidates.sort_by_key(|candidate| candidate.record.created_unix_nanos);
    let tied_latest = candidates.len() > 1
        && candidates
            .last()
            .zip(candidates.iter().rev().nth(1))
            .is_some_and(|(latest, previous)| {
                latest.record.created_unix_nanos == previous.record.created_unix_nanos
            });
    if tied_latest {
        corrupt += 1;
        candidates.clear();
    }
    let latest = candidates.pop();
    let unsafe_evidence = corrupt > 0 || legacy_explicit > 0 || legacy_project > 0;
    let (lifecycle, latest_run, next) =
        summarize(environment, latest.as_ref(), unsafe_evidence, corrupt > 0);
    let mut checks = vec![StatusCheck {
        id: "project".to_owned(),
        status: "pass".to_owned(),
        message: "project configuration, lock, and selected input are valid".to_owned(),
    }];
    checks.push(count_check(
        "corrupt-runs",
        corrupt,
        "corrupt or untrusted run entries were ignored",
    ));
    checks.push(count_check(
        "legacy-explicit-runs",
        legacy_explicit,
        "legacy explicit runs cannot authorize project lifecycle",
    ));
    checks.push(count_check(
        "legacy-project-runs",
        legacy_project,
        "project plans without durable run history require review",
    ));
    Ok(StatusReport {
        api_version: "ainfra.status/v1alpha1",
        kind: "AinfraStatus",
        ok: true,
        project: StatusProject {
            root: project.root.display().to_string(),
            environment: environment.to_owned(),
            template: project.lock.template.name.clone(),
            template_version: project.lock.template.version.clone(),
        },
        lifecycle,
        latest_run,
        checks,
        next,
    })
}

fn local_integrity(
    project_root: &Path,
    input: &Path,
    plan: &PlanRecord,
    binding: &ProjectBinding,
) -> &'static str {
    if plan.project_config_sha256.as_deref() != Some(binding.config_sha256.as_str())
        || plan.project_lock_sha256.as_deref() != Some(binding.lock_sha256.as_str())
        || Path::new(&plan.input_path) != input
        || file_hash(input).ok().as_deref() != Some(plan.input_sha256.as_str())
    {
        return "stale";
    }
    let workspace = project_root
        .join(".ainfra/runs")
        .join(&plan.id)
        .join("workspace");
    if template_hash(&workspace).ok().as_deref() != Some(plan.template_sha256.as_str()) {
        return "stale";
    }
    "valid"
}

fn summarize(
    environment: &str,
    candidate: Option<&Candidate>,
    unsafe_evidence: bool,
    corrupt_evidence: bool,
) -> (LifecycleStatus, Option<LatestRun>, Vec<NextAction>) {
    if unsafe_evidence {
        let latest_run = candidate.map(|candidate| LatestRun {
            id: candidate.record.id.clone(),
            created_unix_nanos: candidate.record.created_unix_nanos,
            operation: candidate.record.operation,
            stages: candidate.events.iter().map(stage_status).collect(),
        });
        return (
            LifecycleStatus {
                state: if corrupt_evidence {
                    "corrupt"
                } else {
                    "legacy"
                }
                .to_owned(),
                run_id: candidate.map(|candidate| candidate.record.id.clone()),
                operation: candidate.map(|candidate| candidate.record.operation),
                integrity: if corrupt_evidence {
                    "corrupt"
                } else {
                    "legacy"
                }
                .to_owned(),
                updated_unix_nanos: candidate.map(|candidate| candidate.updated),
            },
            latest_run,
            vec![manual_recovery()],
        );
    }
    let Some(candidate) = candidate else {
        return (
            LifecycleStatus {
                state: "none".to_owned(),
                run_id: None,
                operation: None,
                integrity: "none".to_owned(),
                updated_unix_nanos: None,
            },
            None,
            vec![NextAction {
                command: format!("ainfra plan --environment {environment}"),
                reason: "no durable project run exists".to_owned(),
            }],
        );
    };
    let last = candidate.events.last();
    let state = if candidate.integrity == "stale" {
        "stale"
    } else {
        state_from(candidate.plan.operation, last)
    };
    let next = next_actions(environment, &candidate.plan, state);
    let latest_run = LatestRun {
        id: candidate.record.id.clone(),
        created_unix_nanos: candidate.record.created_unix_nanos,
        operation: candidate.record.operation,
        stages: candidate.events.iter().map(stage_status).collect(),
    };
    (
        LifecycleStatus {
            state: state.to_owned(),
            run_id: Some(candidate.record.id.clone()),
            operation: Some(candidate.record.operation),
            integrity: candidate.integrity.clone(),
            updated_unix_nanos: Some(candidate.updated),
        },
        Some(latest_run),
        next,
    )
}

fn state_from(operation: Operation, event: Option<&RunEvent>) -> &'static str {
    let Some(event) = event else {
        return "corrupt";
    };
    if event.outcome != RunOutcome::Succeeded {
        return "partial";
    }
    match (operation, event.stage) {
        (Operation::Apply, RunStage::Planned) => "planned",
        (Operation::Apply, RunStage::Apply) => "applied",
        (Operation::Apply, RunStage::Output) => "output-collected",
        (Operation::Apply, RunStage::Configure) => "configured",
        (Operation::Destroy, RunStage::Planned) => "destroy-planned",
        (Operation::Destroy, RunStage::Destroy) => "destroyed",
        _ => "corrupt",
    }
}

fn next_actions(environment: &str, plan: &PlanRecord, state: &str) -> Vec<NextAction> {
    let command = match state {
        "planned" => format!(
            "ainfra apply --environment {environment} --approve {}",
            plan.id
        ),
        "applied" => format!(
            "ainfra outputs --environment {environment} --run {}",
            plan.id
        ),
        "output-collected" => format!(
            "ainfra configure --environment {environment} --run {} \
             --known-hosts PATH",
            plan.id
        ),
        "configured" => format!("ainfra plan --environment {environment} --destroy"),
        "destroy-planned" => format!(
            "ainfra destroy --environment {environment} --approve-destroy {}",
            plan.id
        ),
        "destroyed" => format!("ainfra status --environment {environment}"),
        _ => return vec![manual_recovery()],
    };
    vec![NextAction {
        command,
        reason: format!("latest durable lifecycle state is {state}"),
    }]
}

fn manual_recovery() -> NextAction {
    NextAction {
        command: "docs: lifecycle recovery".to_owned(),
        reason: "state is partial, stale, corrupt, or legacy; inspect it before retrying"
            .to_owned(),
    }
}

fn stage_status(event: &RunEvent) -> StageStatus {
    StageStatus {
        name: event.stage.as_str().to_owned(),
        state: match event.outcome {
            RunOutcome::Started => "started",
            RunOutcome::Succeeded => "succeeded",
            RunOutcome::Failed => "failed",
        }
        .to_owned(),
        at_unix_nanos: event.recorded_unix_nanos,
        error_code: event.error_code.clone(),
        recovery: event.recovery,
    }
}

fn count_check(id: &str, count: u64, message: &str) -> StatusCheck {
    StatusCheck {
        id: id.to_owned(),
        status: if count == 0 { "pass" } else { "warn" }.to_owned(),
        message: if count == 0 {
            format!("no {message}")
        } else {
            format!("{count} {message}")
        },
    }
}
