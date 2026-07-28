//! Immutable run identity and append-only lifecycle event records.

use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use regex::Regex;
use serde::{Deserialize, Serialize};

use crate::error::AinfraError;
use crate::plan_record::{Operation, PlanRecord};

/// Current immutable run-record protocol.
pub const RUN_FORMAT_VERSION: &str = "ainfra.run/v1alpha2";
/// Current append-only run-event protocol.
pub const EVENT_FORMAT_VERSION: &str = "ainfra.run-event/v1alpha1";

/// Immutable identity copied from a reviewed plan.
#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(deny_unknown_fields)]
pub struct RunRecord {
    /// Run-record protocol.
    pub format_version: String,
    /// Exact reviewed plan and run identifier.
    pub id: String,
    /// Bound plan-record protocol.
    pub plan_format_version: String,
    /// Intended lifecycle operation.
    pub operation: Operation,
    /// Template identity.
    pub template: String,
    /// Template version.
    pub template_version: String,
    /// Environment identity.
    pub environment: String,
    /// Digest of the exact reviewed plan bytes.
    pub plan_sha256: String,
    /// Project configuration digest for project-bound plans.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_config_sha256: Option<String>,
    /// Project lock digest for project-bound plans.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_lock_sha256: Option<String>,
    /// Creation time as nanoseconds since the Unix epoch.
    pub created_unix_nanos: u64,
}

/// One lifecycle phase represented in the event stream.
#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "kebab-case")]
pub enum RunStage {
    /// Reviewed plan became durable.
    Planned,
    /// Exact apply operation.
    Apply,
    /// Standardized output collection.
    Output,
    /// Ansible host configuration.
    Configure,
    /// Exact destroy operation.
    Destroy,
}

impl RunStage {
    /// Stable file and machine-output spelling.
    #[must_use]
    pub const fn as_str(self) -> &'static str {
        match self {
            Self::Planned => "planned",
            Self::Apply => "apply",
            Self::Output => "output",
            Self::Configure => "configure",
            Self::Destroy => "destroy",
        }
    }
}

/// Outcome of one stage attempt.
#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum RunOutcome {
    /// The external or local phase is about to begin.
    Started,
    /// The phase completed and durable evidence was written.
    Succeeded,
    /// The phase returned a controlled failure.
    Failed,
}

/// Finite, non-secret operator recovery guidance.
#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "kebab-case")]
pub enum RecoveryCategory {
    /// Inspect the infrastructure provider and `OpenTofu` state.
    InspectInfrastructureState,
    /// Inspect generated output and its immutable local evidence.
    InspectOutputState,
    /// Inspect managed hosts and verified SSH host-key evidence.
    InspectHostState,
}

impl RunOutcome {
    const fn as_str(self) -> &'static str {
        match self {
            Self::Started => "started",
            Self::Succeeded => "succeeded",
            Self::Failed => "failed",
        }
    }
}

/// One immutable sequenced lifecycle event.
#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(deny_unknown_fields)]
pub struct RunEvent {
    /// Event protocol.
    pub format_version: String,
    /// Parent run identifier.
    pub run_id: String,
    /// Contiguous one-based sequence.
    pub sequence: u64,
    /// Lifecycle phase.
    pub stage: RunStage,
    /// Attempt outcome.
    pub outcome: RunOutcome,
    /// Recording time as nanoseconds since the Unix epoch.
    pub recorded_unix_nanos: u64,
    /// Stable public error identifier for failed attempts.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub error_code: Option<String>,
    /// Non-secret recovery category.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub recovery: Option<RecoveryCategory>,
}

impl RunRecord {
    /// Create run identity and the initial planned event without overwrite.
    ///
    /// # Errors
    ///
    /// Rejects unsafe paths, existing records, or incomplete reviewed plans.
    pub fn create(project_root: &Path, plan: &PlanRecord) -> Result<Self, AinfraError> {
        validate_id(&plan.id)?;
        if plan.plan_sha256.is_empty() {
            return Err(AinfraError::guard(
                "cannot record a run before plan bytes are bound",
            ));
        }
        let record = Self {
            format_version: RUN_FORMAT_VERSION.to_owned(),
            id: plan.id.clone(),
            plan_format_version: plan.format_version.clone(),
            operation: plan.operation,
            template: plan.template.clone(),
            template_version: plan.template_version.clone(),
            environment: plan.environment.clone(),
            plan_sha256: plan.plan_sha256.clone(),
            project_config_sha256: plan.project_config_sha256.clone(),
            project_lock_sha256: plan.project_lock_sha256.clone(),
            created_unix_nanos: timestamp()?,
        };
        let run = run_directory(project_root, &record.id)?;
        let path = run.join("run.json");
        write_new_json(&path, &record)?;
        fs::create_dir(run.join("events")).map_err(|error| {
            let _ = fs::remove_file(&path);
            AinfraError::guard(format!("cannot create run event directory: {error}"))
        })?;
        if let Err(error) = record.append(
            project_root,
            RunStage::Planned,
            RunOutcome::Succeeded,
            None,
            None,
        ) {
            let _ = fs::remove_dir(run.join("events"));
            let _ = fs::remove_file(path);
            return Err(error);
        }
        Ok(record)
    }

    /// Load and validate one run identity against its reviewed plan.
    ///
    /// # Errors
    ///
    /// Rejects missing, malformed, redirected, or mismatched records.
    pub fn load(project_root: &Path, id: &str) -> Result<Self, AinfraError> {
        let run = run_directory(project_root, id)?;
        let path = regular_file(&run, "run.json")?;
        let content = fs::read_to_string(&path)
            .map_err(|error| AinfraError::guard(format!("cannot read run record: {error}")))?;
        let record: Self = serde_json::from_str(&content)
            .map_err(|error| AinfraError::guard(format!("invalid run record: {error}")))?;
        record.validate_shape(id)?;
        let plan = PlanRecord::load(project_root, id)?;
        if record.plan_format_version != plan.format_version
            || record.operation != plan.operation
            || record.template != plan.template
            || record.template_version != plan.template_version
            || record.environment != plan.environment
            || record.plan_sha256 != plan.plan_sha256
            || record.project_config_sha256 != plan.project_config_sha256
            || record.project_lock_sha256 != plan.project_lock_sha256
        {
            return Err(AinfraError::guard(
                "run record does not match its reviewed plan",
            ));
        }
        Ok(record)
    }

    /// Load the complete contiguous append-only event stream.
    ///
    /// # Errors
    ///
    /// Rejects symlinks, malformed filenames, gaps, duplicates, and conflicts.
    pub fn events(&self, project_root: &Path) -> Result<Vec<RunEvent>, AinfraError> {
        let run = run_directory(project_root, &self.id)?;
        let directory = run.join("events");
        let metadata = fs::symlink_metadata(&directory)
            .map_err(|_| AinfraError::guard("run event directory is missing"))?;
        if metadata.file_type().is_symlink() || !metadata.is_dir() {
            return Err(AinfraError::guard(
                "run event path must be a regular non-symlink directory",
            ));
        }
        let mut paths = Vec::new();
        for entry in fs::read_dir(&directory)
            .map_err(|error| AinfraError::guard(format!("cannot read run events: {error}")))?
        {
            let entry = entry.map_err(|error| AinfraError::guard(error.to_string()))?;
            let metadata = entry
                .file_type()
                .map_err(|error| AinfraError::guard(error.to_string()))?;
            if metadata.is_symlink() || !metadata.is_file() {
                return Err(AinfraError::guard(
                    "run event entries must be regular files",
                ));
            }
            paths.push(entry.path());
        }
        paths.sort();
        let mut result = Vec::new();
        for (index, path) in paths.iter().enumerate() {
            let event: RunEvent =
                serde_json::from_slice(&fs::read(path).map_err(|error| {
                    AinfraError::guard(format!("cannot read run event: {error}"))
                })?)
                .map_err(|error| AinfraError::guard(format!("invalid run event: {error}")))?;
            let expected =
                u64::try_from(index).map_err(|_| AinfraError::guard("too many run events"))? + 1;
            event.validate(&self.id, expected, path)?;
            result.push(event);
        }
        validate_history(self.operation, &result)?;
        if result
            .first()
            .is_some_and(|event| event.recorded_unix_nanos < self.created_unix_nanos)
        {
            return Err(AinfraError::guard(
                "run events cannot predate their run record",
            ));
        }
        Ok(result)
    }

    /// Record a started stage unless that stage has already succeeded.
    ///
    /// # Errors
    ///
    /// Rejects invalid history or event persistence failures.
    pub fn start(&self, project_root: &Path, stage: RunStage) -> Result<bool, AinfraError> {
        if self
            .events(project_root)?
            .iter()
            .any(|event| event.stage == stage && event.outcome == RunOutcome::Succeeded)
        {
            return Ok(false);
        }
        self.append(project_root, stage, RunOutcome::Started, None, None)?;
        Ok(true)
    }

    /// Record successful completion of one stage.
    ///
    /// # Errors
    ///
    /// Requires the latest event to be the matching started stage.
    pub fn succeed(&self, project_root: &Path, stage: RunStage) -> Result<(), AinfraError> {
        self.append(project_root, stage, RunOutcome::Succeeded, None, None)
    }

    /// Record a controlled non-secret failure.
    ///
    /// # Errors
    ///
    /// Requires the latest event to be the matching started stage.
    pub fn fail(
        &self,
        project_root: &Path,
        stage: RunStage,
        error_code: &str,
        recovery: RecoveryCategory,
    ) -> Result<(), AinfraError> {
        self.append(
            project_root,
            stage,
            RunOutcome::Failed,
            Some(error_code),
            Some(recovery),
        )
    }

    fn append(
        &self,
        project_root: &Path,
        stage: RunStage,
        outcome: RunOutcome,
        error_code: Option<&str>,
        recovery: Option<RecoveryCategory>,
    ) -> Result<(), AinfraError> {
        let events = if stage == RunStage::Planned {
            Vec::new()
        } else {
            self.events(project_root)?
        };
        validate_append(self.operation, &events, stage, outcome)?;
        let sequence =
            u64::try_from(events.len()).map_err(|_| AinfraError::guard("too many run events"))? + 1;
        let event = RunEvent {
            format_version: EVENT_FORMAT_VERSION.to_owned(),
            run_id: self.id.clone(),
            sequence,
            stage,
            outcome,
            recorded_unix_nanos: timestamp()?,
            error_code: error_code.map(str::to_owned),
            recovery,
        };
        let run = run_directory(project_root, &self.id)?;
        let path = run.join("events").join(format!(
            "{sequence:06}-{}-{}.json",
            stage.as_str(),
            outcome.as_str()
        ));
        write_new_json(&path, &event)
    }

    fn validate_shape(&self, id: &str) -> Result<(), AinfraError> {
        if self.format_version != RUN_FORMAT_VERSION {
            return Err(AinfraError::guard("unsupported run record version"));
        }
        if self.id != id {
            return Err(AinfraError::guard(
                "run record ID does not match its directory",
            ));
        }
        validate_id(&self.id)?;
        if self.created_unix_nanos == 0 {
            return Err(AinfraError::guard("run creation timestamp must be nonzero"));
        }
        validate_digest(&self.plan_sha256)?;
        if self.project_config_sha256.is_some() != self.project_lock_sha256.is_some() {
            return Err(AinfraError::guard(
                "run record has incomplete project bindings",
            ));
        }
        if let Some(value) = &self.project_config_sha256 {
            validate_digest(value)?;
        }
        if let Some(value) = &self.project_lock_sha256 {
            validate_digest(value)?;
        }
        Ok(())
    }
}

impl RunEvent {
    fn validate(&self, id: &str, sequence: u64, path: &Path) -> Result<(), AinfraError> {
        if self.format_version != EVENT_FORMAT_VERSION
            || self.run_id != id
            || self.sequence != sequence
        {
            return Err(AinfraError::guard("run event identity is invalid"));
        }
        let error_pattern = Regex::new(r"^AINFRA-E[0-9]{3}$")
            .map_err(|error| AinfraError::dependency(error.to_string()))?;
        let expected = format!(
            "{sequence:06}-{}-{}.json",
            self.stage.as_str(),
            self.outcome.as_str()
        );
        if path.file_name().and_then(|name| name.to_str()) != Some(expected.as_str()) {
            return Err(AinfraError::guard("run event filename is invalid"));
        }
        match self.outcome {
            RunOutcome::Failed
                if self
                    .error_code
                    .as_deref()
                    .is_none_or(|code| !error_pattern.is_match(code))
                    || self.recovery.is_none() =>
            {
                Err(AinfraError::guard(
                    "failed run event requires error and recovery",
                ))
            }
            RunOutcome::Started | RunOutcome::Succeeded
                if self.error_code.is_some() || self.recovery.is_some() =>
            {
                Err(AinfraError::guard(
                    "non-failed run event contains failure details",
                ))
            }
            _ => Ok(()),
        }
    }
}

fn validate_history(operation: Operation, events: &[RunEvent]) -> Result<(), AinfraError> {
    if events.is_empty()
        || events[0].stage != RunStage::Planned
        || events[0].outcome != RunOutcome::Succeeded
    {
        return Err(AinfraError::guard(
            "run history must begin with a successful planned event",
        ));
    }
    let mut history = Vec::new();
    let mut previous_time = 0;
    for event in events {
        if event.recorded_unix_nanos == 0 || event.recorded_unix_nanos < previous_time {
            return Err(AinfraError::guard(
                "run event timestamps must be nonzero and monotonic",
            ));
        }
        validate_append(operation, &history, event.stage, event.outcome)?;
        previous_time = event.recorded_unix_nanos;
        history.push(event.clone());
    }
    Ok(())
}

fn validate_append(
    operation: Operation,
    events: &[RunEvent],
    stage: RunStage,
    outcome: RunOutcome,
) -> Result<(), AinfraError> {
    if events.is_empty() {
        if stage == RunStage::Planned && outcome == RunOutcome::Succeeded {
            return Ok(());
        }
        return Err(AinfraError::guard("run history must begin with planned"));
    }
    let latest = events
        .last()
        .ok_or_else(|| AinfraError::guard("run history is empty"))?;
    match outcome {
        RunOutcome::Started => {
            let allowed = match (operation, stage) {
                (Operation::Apply, RunStage::Apply) | (Operation::Destroy, RunStage::Destroy) => {
                    succeeded(events, RunStage::Planned)
                }
                (Operation::Apply, RunStage::Output) => succeeded(events, RunStage::Apply),
                (Operation::Apply, RunStage::Configure) => succeeded(events, RunStage::Output),
                _ => false,
            };
            if !allowed {
                return Err(AinfraError::guard("invalid run stage transition"));
            }
        }
        RunOutcome::Succeeded | RunOutcome::Failed
            if latest.stage != stage || latest.outcome != RunOutcome::Started =>
        {
            return Err(AinfraError::guard(
                "run completion does not match the latest started stage",
            ));
        }
        RunOutcome::Succeeded | RunOutcome::Failed => {}
    }
    Ok(())
}

fn succeeded(events: &[RunEvent], stage: RunStage) -> bool {
    events
        .iter()
        .any(|event| event.stage == stage && event.outcome == RunOutcome::Succeeded)
}

fn run_directory(project_root: &Path, id: &str) -> Result<PathBuf, AinfraError> {
    validate_id(id)?;
    let root = project_root
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
    let run = root.join(".ainfra/runs").join(id);
    let metadata = fs::symlink_metadata(&run)
        .map_err(|_| AinfraError::guard(format!("run directory does not exist: {id}")))?;
    if metadata.file_type().is_symlink() || !metadata.is_dir() {
        return Err(AinfraError::guard(
            "run path must be a regular non-symlink directory",
        ));
    }
    Ok(run)
}

fn regular_file(directory: &Path, name: &str) -> Result<PathBuf, AinfraError> {
    let path = directory.join(name);
    let metadata = fs::symlink_metadata(&path)
        .map_err(|_| AinfraError::guard(format!("{name} is missing")))?;
    if metadata.file_type().is_symlink() || !metadata.is_file() {
        return Err(AinfraError::guard(format!(
            "{name} must be a regular non-symlink file"
        )));
    }
    Ok(path)
}

fn write_new_json<T: Serialize>(path: &Path, value: &T) -> Result<(), AinfraError> {
    let content = serde_json::to_vec_pretty(value)
        .map_err(|error| AinfraError::guard(format!("cannot serialize run state: {error}")))?;
    let mut file = OpenOptions::new()
        .create_new(true)
        .write(true)
        .open(path)
        .map_err(|error| {
            AinfraError::guard(format!("cannot create {}: {error}", path.display()))
        })?;
    file.write_all(&content)
        .and_then(|()| file.sync_all())
        .map_err(|error| AinfraError::guard(format!("cannot write {}: {error}", path.display())))
}

fn timestamp() -> Result<u64, AinfraError> {
    let nanos = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|error| AinfraError::dependency(format!("system clock is invalid: {error}")))?
        .as_nanos();
    u64::try_from(nanos).map_err(|_| AinfraError::dependency("system clock is out of range"))
}

fn validate_id(id: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[0-9a-f]{20}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(id) {
        return Err(AinfraError::guard("invalid run ID"));
    }
    Ok(())
}

fn validate_digest(value: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[0-9a-f]{64}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(value) {
        return Err(AinfraError::guard("invalid digest in run record"));
    }
    Ok(())
}
