//! Read-only inspection of Python-prototype lifecycle evidence.

use std::fs;
use std::path::Path;

use regex::Regex;
use serde::Serialize;

use crate::error::AinfraError;
use crate::plan_record::{FORMAT_VERSION, Operation, PlanRecord, secure_file_hash};

/// Sanitized legacy-evidence inspection report.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LegacyReport {
    /// Machine-output protocol.
    pub api_version: &'static str,
    /// Contract kind.
    pub kind: &'static str,
    /// Whether read-only inspection completed.
    pub ok: bool,
    /// `none`, `legacy`, or `corrupt`.
    pub state: String,
    /// Structurally valid Python-compatible records.
    pub legacy_records: u64,
    /// Malformed, redirected, or untrusted entries.
    pub corrupt_entries: u64,
    /// Sanitized valid record identities.
    pub runs: Vec<LegacyRun>,
    /// Legacy evidence can never become project authorization in place.
    pub requires_new_project_plan: bool,
    /// Stable recovery documentation route.
    pub next: &'static str,
}

/// Non-secret identity and plan-integrity result for one legacy record.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LegacyRun {
    /// Exact legacy plan identifier.
    pub id: String,
    /// Intended operation.
    pub operation: Operation,
    /// Environment identity.
    pub environment: String,
    /// Template identity.
    pub template: String,
    /// `valid`, `missing`, or `mismatch`.
    pub plan_integrity: String,
    /// Lifecycle state remains unknowable from a plan record.
    pub lifecycle_state: &'static str,
}

/// Inspect a Python-prototype repository without invoking tools or credentials.
///
/// # Errors
///
/// Rejects an invalid source root or redirected `.ainfra/runs` boundary.
/// Individual untrusted run entries are counted without exposing their content.
pub fn inspect(source_root: &Path) -> Result<LegacyReport, AinfraError> {
    let root = source_root
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve legacy root: {error}")))?;
    let state_root = root.join(".ainfra");
    if !state_root.exists() {
        return Ok(report(Vec::new(), 0));
    }
    require_directory(&state_root, "legacy operational state directory")?;
    let runs_root = state_root.join("runs");
    if !runs_root.exists() {
        return Ok(report(Vec::new(), 0));
    }
    require_directory(&runs_root, "legacy run root")?;
    let mut runs = Vec::new();
    let mut corrupt = 0_u64;
    for entry in fs::read_dir(&runs_root)
        .map_err(|error| AinfraError::guard(format!("cannot inspect legacy runs: {error}")))?
    {
        let entry = entry.map_err(|error| AinfraError::guard(error.to_string()))?;
        let file_type = entry
            .file_type()
            .map_err(|error| AinfraError::guard(error.to_string()))?;
        if file_type.is_symlink() || !file_type.is_dir() {
            corrupt += 1;
            continue;
        }
        let Some(id) = entry.file_name().to_str().map(str::to_owned) else {
            corrupt += 1;
            continue;
        };
        if !valid_id(&id) {
            corrupt += 1;
            continue;
        }
        match inspect_record(&root, &id) {
            Ok(record) => runs.push(record),
            Err(_) => corrupt += 1,
        }
    }
    runs.sort_by(|left, right| left.id.cmp(&right.id));
    Ok(report(runs, corrupt))
}

fn inspect_record(root: &Path, id: &str) -> Result<LegacyRun, AinfraError> {
    let record = PlanRecord::load(root, id)?;
    if record.format_version != FORMAT_VERSION {
        return Err(AinfraError::guard(
            "project-bound records are not legacy evidence",
        ));
    }
    let expected = root
        .join(".ainfra/runs")
        .join(id)
        .join(format!("{}.tfplan", record.operation.as_str()));
    if Path::new(&record.plan_path) != expected {
        return Err(AinfraError::guard("legacy plan path is not contained"));
    }
    let plan_integrity = match fs::symlink_metadata(&expected) {
        Ok(metadata) if metadata.file_type().is_symlink() || !metadata.is_file() => {
            return Err(AinfraError::guard(
                "legacy plan must be a regular non-symlink file",
            ));
        }
        Ok(_) if secure_file_hash(root, &expected, "legacy plan")? == record.plan_sha256 => "valid",
        Ok(_) => "mismatch",
        Err(_) => "missing",
    };
    Ok(LegacyRun {
        id: record.id,
        operation: record.operation,
        environment: record.environment,
        template: record.template,
        plan_integrity: plan_integrity.to_owned(),
        lifecycle_state: "unknown",
    })
}

fn report(runs: Vec<LegacyRun>, corrupt_entries: u64) -> LegacyReport {
    let legacy_records = u64::try_from(runs.len()).unwrap_or(u64::MAX);
    let degraded = runs.iter().any(|run| run.plan_integrity != "valid");
    let state = if corrupt_entries > 0 || degraded {
        "corrupt"
    } else if legacy_records > 0 {
        "legacy"
    } else {
        "none"
    };
    LegacyReport {
        api_version: "ainfra.legacy-inspection/v1alpha1",
        kind: "AinfraLegacyInspection",
        ok: true,
        state: state.to_owned(),
        legacy_records,
        corrupt_entries,
        runs,
        requires_new_project_plan: true,
        next: "docs: legacy recovery",
    }
}

fn require_directory(path: &Path, label: &str) -> Result<(), AinfraError> {
    let metadata = fs::symlink_metadata(path)
        .map_err(|_| AinfraError::guard(format!("{label} is missing")))?;
    if metadata.file_type().is_symlink() || !metadata.is_dir() {
        return Err(AinfraError::guard(format!(
            "{label} must be a regular non-symlink directory"
        )));
    }
    Ok(())
}

fn valid_id(id: &str) -> bool {
    Regex::new(r"^[0-9a-f]{20}$").is_ok_and(|pattern| pattern.is_match(id))
}
