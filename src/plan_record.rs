//! Versioned immutable bindings for reviewed `OpenTofu` plans.

use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

use regex::Regex;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::error::AinfraError;

/// Current on-disk plan-record protocol.
pub const FORMAT_VERSION: &str = "ainfra.plan/v1alpha1";

/// Intended use of an immutable reviewed plan.
#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Operation {
    /// Create or update infrastructure.
    Apply,
    /// Remove infrastructure through a reviewed destroy plan.
    Destroy,
}

impl Operation {
    /// Return the stable CLI spelling.
    #[must_use]
    pub const fn as_str(self) -> &'static str {
        match self {
            Self::Apply => "apply",
            Self::Destroy => "destroy",
        }
    }
}

/// Immutable binding between configuration, template, and plan bytes.
#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(deny_unknown_fields)]
pub struct PlanRecord {
    /// Version of this security protocol.
    #[serde(default = "legacy_format_version")]
    pub format_version: String,
    /// Exact approval identifier.
    pub id: String,
    /// Authorized operation.
    pub operation: Operation,
    /// Template identity.
    pub template: String,
    /// Template manifest version.
    pub template_version: String,
    /// Environment identity.
    pub environment: String,
    /// Canonical input path.
    pub input_path: String,
    /// Digest of the original input bytes.
    pub input_sha256: String,
    /// Digest of the resolved template tree.
    pub template_sha256: String,
    /// Absolute immutable `OpenTofu` plan path.
    pub plan_path: String,
    /// Digest of the `OpenTofu` plan bytes.
    pub plan_sha256: String,
}

/// Expected bindings supplied again at apply or destroy time.
pub struct ExpectedBindings<'a> {
    /// Selected template identity.
    pub template: &'a str,
    /// Current template manifest version.
    pub template_version: &'a str,
    /// Current environment identity.
    pub environment: &'a str,
    /// Current input file.
    pub input_path: &'a Path,
    /// Current template tree.
    pub template_root: &'a Path,
    /// Requested operation.
    pub operation: Operation,
}

impl PlanRecord {
    /// Allocate an unpersisted record using Python-compatible binding fields.
    ///
    /// # Errors
    ///
    /// Returns a guard error when paths cannot be canonicalized or hashed.
    pub fn create(
        project_root: &Path,
        operation: Operation,
        template: &str,
        template_version: &str,
        environment: &str,
        input_path: &Path,
        template_root: &Path,
    ) -> Result<Self, AinfraError> {
        let id = Uuid::new_v4().simple().to_string()[..20].to_owned();
        let run_root = run_root(project_root).join(&id);
        let plan_path = run_root.join(format!("{}.tfplan", operation.as_str()));
        let canonical_input = input_path.canonicalize().map_err(|error| {
            AinfraError::guard(format!(
                "cannot resolve input path {}: {error}",
                input_path.display()
            ))
        })?;
        Ok(Self {
            format_version: FORMAT_VERSION.to_owned(),
            id,
            operation,
            template: template.to_owned(),
            template_version: template_version.to_owned(),
            environment: environment.to_owned(),
            input_path: canonical_input.display().to_string(),
            input_sha256: file_hash(&canonical_input)?,
            template_sha256: template_hash(template_root)?,
            plan_path: plan_path.display().to_string(),
            plan_sha256: String::new(),
        })
    }

    /// Bind the produced plan bytes to this record.
    ///
    /// # Errors
    ///
    /// Returns a guard error when the declared plan is missing or unsafe.
    pub fn bind_plan(&mut self, project_root: &Path) -> Result<(), AinfraError> {
        let plan = contained_plan_path(project_root, self)?;
        self.plan_sha256 = file_hash(&plan)?;
        Ok(())
    }

    /// Persist this record without overwriting an existing binding.
    ///
    /// # Errors
    ///
    /// Returns a guard error for invalid records or filesystem failures.
    pub fn write(&self, project_root: &Path) -> Result<PathBuf, AinfraError> {
        validate_id(&self.id)?;
        if self.format_version != FORMAT_VERSION {
            return Err(AinfraError::guard("unsupported plan record version"));
        }
        let run = run_root(project_root).join(&self.id);
        fs::create_dir_all(&run).map_err(|error| {
            AinfraError::guard(format!("cannot create {}: {error}", run.display()))
        })?;
        let path = run.join("plan.json");
        let mut file = OpenOptions::new()
            .create_new(true)
            .write(true)
            .open(&path)
            .map_err(|error| {
                AinfraError::guard(format!("cannot create {}: {error}", path.display()))
            })?;
        let document = serde_json::to_string_pretty(self)
            .map_err(|error| AinfraError::guard(format!("cannot serialize plan: {error}")))?;
        file.write_all(document.as_bytes()).map_err(|error| {
            AinfraError::guard(format!("cannot write {}: {error}", path.display()))
        })?;
        Ok(path)
    }

    /// Load one exact reviewed plan record.
    ///
    /// Normal Python records without `format_version` are read as v1alpha1.
    ///
    /// # Errors
    ///
    /// Returns a guard error for unsafe IDs, missing, malformed, or unsupported
    /// records.
    pub fn load(project_root: &Path, id: &str) -> Result<Self, AinfraError> {
        validate_id(id)?;
        let path = run_root(project_root).join(id).join("plan.json");
        let content = fs::read_to_string(&path)
            .map_err(|_| AinfraError::guard(format!("reviewed plan does not exist: {id}")))?;
        let record: Self = serde_json::from_str(&content)
            .map_err(|error| AinfraError::guard(format!("invalid plan record: {error}")))?;
        if record.id != id {
            return Err(AinfraError::guard(
                "reviewed plan record ID does not match its directory",
            ));
        }
        if record.format_version != FORMAT_VERSION {
            return Err(AinfraError::guard("unsupported plan record version"));
        }
        Ok(record)
    }

    /// Verify every current binding before any lifecycle process runs.
    ///
    /// # Errors
    ///
    /// Returns the first stable guard failure in Python-compatible order.
    pub fn verify(
        &self,
        project_root: &Path,
        expected: &ExpectedBindings<'_>,
    ) -> Result<(), AinfraError> {
        if self.operation != expected.operation {
            return Err(AinfraError::guard(format!(
                "reviewed {} plan cannot authorize {}",
                self.operation.as_str(),
                expected.operation.as_str()
            )));
        }
        if self.template != expected.template {
            return Err(AinfraError::guard(
                "reviewed plan belongs to another template",
            ));
        }
        if self.template_version != expected.template_version {
            return Err(AinfraError::guard(
                "template version changed after planning",
            ));
        }
        if self.environment != expected.environment {
            return Err(AinfraError::guard("environment changed after planning"));
        }
        let input = expected.input_path.canonicalize().map_err(|error| {
            AinfraError::guard(format!(
                "cannot resolve input path {}: {error}",
                expected.input_path.display()
            ))
        })?;
        if self.input_path != input.display().to_string() {
            return Err(AinfraError::guard("input path changed after planning"));
        }
        if self.input_sha256 != file_hash(&input)? {
            return Err(AinfraError::guard("input content changed after planning"));
        }
        if self.template_sha256 != template_hash(expected.template_root)? {
            return Err(AinfraError::guard(
                "template content changed after planning",
            ));
        }
        let plan = contained_plan_path(project_root, self)?;
        if self.plan_sha256 != file_hash(&plan)? {
            return Err(AinfraError::guard(
                "reviewed plan file was modified after planning",
            ));
        }
        Ok(())
    }
}

/// Hash raw bytes using lowercase SHA-256.
///
/// # Errors
///
/// Returns a guard error when the file cannot be read.
pub fn file_hash(path: &Path) -> Result<String, AinfraError> {
    let bytes = fs::read(path)
        .map_err(|error| AinfraError::guard(format!("cannot read {}: {error}", path.display())))?;
    Ok(format!("{:x}", Sha256::digest(bytes)))
}

/// Reproduce the Python template-tree digest on portable UTF-8 paths.
///
/// # Errors
///
/// Returns a guard error for unreadable trees, non-UTF-8 paths, or symlinks.
pub fn template_hash(root: &Path) -> Result<String, AinfraError> {
    let mut paths = Vec::new();
    collect_files(root, root, &mut paths)?;
    paths.sort();
    let mut digest = Sha256::new();
    for relative in paths {
        let text = relative
            .to_str()
            .ok_or_else(|| AinfraError::guard("template contains a non-UTF-8 path"))?;
        digest.update(text.replace('\\', "/").as_bytes());
        digest.update(fs::read(root.join(&relative)).map_err(|error| {
            AinfraError::guard(format!(
                "cannot read template file {}: {error}",
                relative.display()
            ))
        })?);
    }
    Ok(format!("{:x}", digest.finalize()))
}

fn collect_files(
    root: &Path,
    directory: &Path,
    result: &mut Vec<PathBuf>,
) -> Result<(), AinfraError> {
    let entries = fs::read_dir(directory).map_err(|error| {
        AinfraError::guard(format!("cannot read {}: {error}", directory.display()))
    })?;
    for entry in entries {
        let entry = entry.map_err(|error| AinfraError::guard(error.to_string()))?;
        let path = entry.path();
        let relative = path
            .strip_prefix(root)
            .map_err(|error| AinfraError::guard(format!("template path escaped root: {error}")))?;
        if relative.components().any(|part| {
            matches!(
                part.as_os_str().to_str(),
                Some(".ainfra" | ".terraform" | "__pycache__")
            )
        }) {
            continue;
        }
        let metadata =
            fs::symlink_metadata(&path).map_err(|error| AinfraError::guard(error.to_string()))?;
        if metadata.file_type().is_symlink() {
            return Err(AinfraError::guard("template tree contains a symbolic link"));
        }
        if metadata.is_dir() {
            collect_files(root, &path, result)?;
        } else if metadata.is_file() {
            result.push(relative.to_path_buf());
        }
    }
    Ok(())
}

fn contained_plan_path(project_root: &Path, record: &PlanRecord) -> Result<PathBuf, AinfraError> {
    validate_id(&record.id)?;
    let expected = run_root(project_root).join(&record.id);
    let plan = PathBuf::from(&record.plan_path);
    if !plan.is_file() {
        return Err(AinfraError::guard("reviewed plan file is missing"));
    }
    let resolved = plan
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve reviewed plan: {error}")))?;
    let allowed = expected
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve plan directory: {error}")))?;
    if !resolved.starts_with(&allowed) {
        return Err(AinfraError::guard(
            "reviewed plan path escaped its run directory",
        ));
    }
    Ok(resolved)
}

fn validate_id(id: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[0-9a-f]{20}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(id) {
        return Err(AinfraError::guard("invalid reviewed plan ID"));
    }
    Ok(())
}

fn run_root(project_root: &Path) -> PathBuf {
    project_root.join(".ainfra/runs")
}

fn legacy_format_version() -> String {
    FORMAT_VERSION.to_owned()
}
