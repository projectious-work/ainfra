//! Versioned immutable bindings for reviewed `OpenTofu` plans.

use std::fs::{self, OpenOptions};
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

use regex::Regex;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use uuid::Uuid;

use crate::error::AinfraError;

/// Current on-disk plan-record protocol.
pub const FORMAT_VERSION: &str = "ainfra.plan/v1alpha1";
/// Project-bound on-disk plan-record protocol.
pub const PROJECT_FORMAT_VERSION: &str = "ainfra.plan/v1alpha2";

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
    /// Canonical project root for project-driven lifecycle.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_root: Option<String>,
    /// Canonical human-authored project configuration.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_config_path: Option<String>,
    /// Digest of the exact project configuration bytes.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_config_sha256: Option<String>,
    /// Canonical generated project lockfile.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_lock_path: Option<String>,
    /// Digest of the exact project lockfile bytes.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub project_lock_sha256: Option<String>,
    /// Canonical reviewed backend configuration path.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub backend_config_path: Option<String>,
    /// Digest of the reviewed backend configuration bytes.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub backend_config_sha256: Option<String>,
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
    /// Current canonical backend configuration path.
    pub backend_config_path: Option<&'a str>,
    /// Current backend configuration digest.
    pub backend_config_sha256: Option<&'a str>,
    /// Current project boundary, absent only for explicit compatibility mode.
    pub project: Option<&'a ProjectBinding>,
}

/// Exact project files bound into one reviewed plan.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct ProjectBinding {
    /// Canonical project root.
    pub root: PathBuf,
    /// Canonical direct-child `ainfra.yaml`.
    pub config_path: PathBuf,
    /// SHA-256 of the exact configuration bytes.
    pub config_sha256: String,
    /// Canonical direct-child `ainfra.lock`.
    pub lock_path: PathBuf,
    /// SHA-256 of the exact lockfile bytes.
    pub lock_sha256: String,
}

impl ProjectBinding {
    /// Capture exact regular project files without following static symlinks.
    ///
    /// # Errors
    ///
    /// Rejects missing, redirected, non-UTF-8, or unreadable project paths.
    pub fn capture(project_root: &Path) -> Result<Self, AinfraError> {
        let root = project_root
            .canonicalize()
            .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
        let config_path = binding_file(&root, "ainfra.yaml")?;
        let lock_path = binding_file(&root, "ainfra.lock")?;
        Ok(Self {
            config_sha256: file_hash(&config_path)?,
            lock_sha256: file_hash(&lock_path)?,
            root,
            config_path,
            lock_path,
        })
    }

    fn serialized(&self) -> Result<[String; 5], AinfraError> {
        Ok([
            utf8_path(&self.root)?,
            utf8_path(&self.config_path)?,
            self.config_sha256.clone(),
            utf8_path(&self.lock_path)?,
            self.lock_sha256.clone(),
        ])
    }
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
        Self::create_with_template_hash(
            project_root,
            operation,
            template,
            template_version,
            environment,
            input_path,
            &template_hash(template_root)?,
        )
    }

    /// Allocate a record from an already verified template digest.
    ///
    /// # Errors
    ///
    /// Returns a guard error when the input cannot be canonicalized or hashed.
    pub fn create_with_template_hash(
        project_root: &Path,
        operation: Operation,
        template: &str,
        template_version: &str,
        environment: &str,
        input_path: &Path,
        template_sha256: &str,
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
            template_sha256: template_sha256.to_owned(),
            project_root: None,
            project_config_path: None,
            project_config_sha256: None,
            project_lock_path: None,
            project_lock_sha256: None,
            backend_config_path: None,
            backend_config_sha256: None,
            plan_path: plan_path.display().to_string(),
            plan_sha256: String::new(),
        })
    }

    /// Upgrade a new unpersisted record to the project-bound v1alpha2 protocol.
    ///
    /// # Errors
    ///
    /// Rejects non-UTF-8 project paths.
    pub fn bind_project(&mut self, binding: &ProjectBinding) -> Result<(), AinfraError> {
        let [root, config_path, config_sha256, lock_path, lock_sha256] = binding.serialized()?;
        PROJECT_FORMAT_VERSION.clone_into(&mut self.format_version);
        self.project_root = Some(root);
        self.project_config_path = Some(config_path);
        self.project_config_sha256 = Some(config_sha256);
        self.project_lock_path = Some(lock_path);
        self.project_lock_sha256 = Some(lock_sha256);
        self.validate_protocol_shape()
    }

    /// Bind the produced plan bytes to this record.
    ///
    /// # Errors
    ///
    /// Returns a guard error when the declared plan is missing or unsafe.
    pub fn bind_plan(&mut self, project_root: &Path) -> Result<(), AinfraError> {
        self.plan_sha256 = contained_plan_hash(project_root, self)?;
        Ok(())
    }

    /// Persist this record without overwriting an existing binding.
    ///
    /// # Errors
    ///
    /// Returns a guard error for invalid records or filesystem failures.
    pub fn write(&self, project_root: &Path) -> Result<PathBuf, AinfraError> {
        validate_id(&self.id)?;
        self.validate_protocol_shape()?;
        validate_digest(&self.plan_sha256, "plan")?;
        self.validate_project_paths(project_root)?;
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
        let root = project_root
            .canonicalize()
            .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
        let state = root.join(".ainfra");
        require_directory(&state, "operational state directory")?;
        let runs = run_root(&root);
        require_directory(&runs, "run root")?;
        let run = runs.join(id);
        require_directory(&run, "plan run directory")?;
        let path = run.join("plan.json");
        let content = secure_read(&root, &path, "plan record")?;
        let record: Self = serde_json::from_slice(&content)
            .map_err(|error| AinfraError::guard(format!("invalid plan record: {error}")))?;
        if record.id != id {
            return Err(AinfraError::guard(
                "reviewed plan record ID does not match its directory",
            ));
        }
        record.validate_protocol_shape()?;
        validate_digest(&record.plan_sha256, "plan")?;
        record.validate_project_paths(&root)?;
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
        self.validate_protocol_shape()?;
        validate_digest(&self.plan_sha256, "plan")?;
        self.validate_project_paths(project_root)?;
        self.verify_project_binding(project_root, expected.project)?;
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
        if self.backend_config_path.as_deref() != expected.backend_config_path
            || self.backend_config_sha256.as_deref() != expected.backend_config_sha256
        {
            return Err(AinfraError::guard(
                "backend configuration changed after planning",
            ));
        }
        if self.plan_sha256 != contained_plan_hash(project_root, self)? {
            return Err(AinfraError::guard(
                "reviewed plan file was modified after planning",
            ));
        }
        Ok(())
    }

    fn verify_project_binding(
        &self,
        project_root: &Path,
        expected: Option<&ProjectBinding>,
    ) -> Result<(), AinfraError> {
        match (self.format_version.as_str(), expected) {
            (FORMAT_VERSION, None) => Ok(()),
            (FORMAT_VERSION, Some(_)) => Err(AinfraError::guard(
                "legacy reviewed plan cannot authorize project-driven lifecycle",
            )),
            (PROJECT_FORMAT_VERSION, None) => Err(AinfraError::guard(
                "project-bound reviewed plan requires project verification",
            )),
            (PROJECT_FORMAT_VERSION, Some(binding)) => {
                let current = ProjectBinding::capture(project_root)?;
                if &current != binding {
                    return Err(AinfraError::guard(
                        "project configuration changed during verification",
                    ));
                }
                let [root, config_path, config_sha256, lock_path, lock_sha256] =
                    current.serialized()?;
                if self.project_root.as_deref() != Some(root.as_str())
                    || self.project_config_path.as_deref() != Some(config_path.as_str())
                    || self.project_config_sha256.as_deref() != Some(config_sha256.as_str())
                    || self.project_lock_path.as_deref() != Some(lock_path.as_str())
                    || self.project_lock_sha256.as_deref() != Some(lock_sha256.as_str())
                {
                    return Err(AinfraError::guard(
                        "project configuration or lock changed after planning",
                    ));
                }
                Ok(())
            }
            _ => Err(AinfraError::guard("unsupported plan record version")),
        }
    }

    fn validate_protocol_shape(&self) -> Result<(), AinfraError> {
        validate_id(&self.id)?;
        validate_name(&self.template, "template", 3)?;
        validate_name(&self.environment, "environment", 2)?;
        if self.template_version.is_empty() || self.template_version.len() > 64 {
            return Err(AinfraError::guard(
                "invalid template version in reviewed plan",
            ));
        }
        validate_digest(&self.input_sha256, "input")?;
        validate_digest(&self.template_sha256, "template")?;
        if !self.plan_sha256.is_empty() {
            validate_digest(&self.plan_sha256, "plan")?;
        }
        let project_fields = [
            self.project_root.as_ref(),
            self.project_config_path.as_ref(),
            self.project_config_sha256.as_ref(),
            self.project_lock_path.as_ref(),
            self.project_lock_sha256.as_ref(),
        ];
        match self.format_version.as_str() {
            FORMAT_VERSION if project_fields.iter().all(Option::is_none) => Ok(()),
            FORMAT_VERSION => Err(AinfraError::guard(
                "legacy plan record contains project bindings",
            )),
            PROJECT_FORMAT_VERSION if project_fields.iter().all(Option::is_some) => {
                validate_digest(
                    self.project_config_sha256.as_deref().unwrap_or_default(),
                    "project configuration",
                )?;
                validate_digest(
                    self.project_lock_sha256.as_deref().unwrap_or_default(),
                    "project lock",
                )
            }
            PROJECT_FORMAT_VERSION => Err(AinfraError::guard(
                "project-bound plan record has incomplete project bindings",
            )),
            _ => Err(AinfraError::guard("unsupported plan record version")),
        }
    }

    fn validate_project_paths(&self, project_root: &Path) -> Result<(), AinfraError> {
        if self.format_version != PROJECT_FORMAT_VERSION {
            return Ok(());
        }
        let root = project_root
            .canonicalize()
            .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
        let expected_root = utf8_path(&root)?;
        let expected_config = utf8_path(&root.join("ainfra.yaml"))?;
        let expected_lock = utf8_path(&root.join("ainfra.lock"))?;
        if self.project_root.as_deref() != Some(expected_root.as_str())
            || self.project_config_path.as_deref() != Some(expected_config.as_str())
            || self.project_lock_path.as_deref() != Some(expected_lock.as_str())
        {
            return Err(AinfraError::guard(
                "project-bound plan paths do not match the project root",
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

fn contained_plan_hash(project_root: &Path, record: &PlanRecord) -> Result<String, AinfraError> {
    validate_id(&record.id)?;
    let root = project_root
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
    let state = root.join(".ainfra");
    require_directory(&state, "operational state directory")?;
    let runs = run_root(&root);
    require_directory(&runs, "run root")?;
    let expected = runs.join(&record.id);
    let plan = PathBuf::from(&record.plan_path);
    if plan != expected.join(format!("{}.tfplan", record.operation.as_str())) {
        return Err(AinfraError::guard(
            "reviewed plan path does not match its operation",
        ));
    }
    require_directory(&expected, "plan run directory")?;
    let bytes = secure_read(&root, &plan, "reviewed plan file")?;
    Ok(format!("{:x}", Sha256::digest(bytes)))
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

fn require_regular_file(path: &Path, label: &str) -> Result<(), AinfraError> {
    let metadata = fs::symlink_metadata(path)
        .map_err(|_| AinfraError::guard(format!("{label} is missing")))?;
    if metadata.file_type().is_symlink() || !metadata.is_file() {
        return Err(AinfraError::guard(format!(
            "{label} must be a regular non-symlink file"
        )));
    }
    Ok(())
}

fn secure_read(root: &Path, path: &Path, label: &str) -> Result<Vec<u8>, AinfraError> {
    let root = root
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
    let parent = path
        .parent()
        .ok_or_else(|| AinfraError::guard(format!("{label} has no parent directory")))?;
    let before_parent = parent
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve {label} parent: {error}")))?;
    if !before_parent.starts_with(&root) {
        return Err(AinfraError::guard(format!(
            "{label} escaped the project root"
        )));
    }
    require_regular_file(path, label)?;
    let before = fs::symlink_metadata(path)
        .map_err(|_| AinfraError::guard(format!("{label} is missing")))?;
    let mut file = fs::File::open(path)
        .map_err(|error| AinfraError::guard(format!("cannot open {label}: {error}")))?;
    let opened = file
        .metadata()
        .map_err(|error| AinfraError::guard(format!("cannot inspect {label}: {error}")))?;
    if !same_file(&before, &opened) {
        return Err(AinfraError::guard(format!(
            "{label} changed while it was opened"
        )));
    }
    let mut bytes = Vec::new();
    file.read_to_end(&mut bytes)
        .map_err(|error| AinfraError::guard(format!("cannot read {label}: {error}")))?;
    let after = fs::symlink_metadata(path)
        .map_err(|_| AinfraError::guard(format!("{label} changed while it was read")))?;
    let after_parent = parent
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve {label} parent: {error}")))?;
    if !same_file(&opened, &after) || after_parent != before_parent {
        return Err(AinfraError::guard(format!(
            "{label} changed while it was read"
        )));
    }
    Ok(bytes)
}

pub(crate) fn secure_file_hash(
    root: &Path,
    path: &Path,
    label: &str,
) -> Result<String, AinfraError> {
    Ok(format!(
        "{:x}",
        Sha256::digest(secure_read(root, path, label)?)
    ))
}

#[cfg(unix)]
fn same_file(left: &fs::Metadata, right: &fs::Metadata) -> bool {
    use std::os::unix::fs::MetadataExt;

    left.dev() == right.dev() && left.ino() == right.ino()
}

#[cfg(not(unix))]
fn same_file(left: &fs::Metadata, right: &fs::Metadata) -> bool {
    left.len() == right.len()
        && left.modified().ok() == right.modified().ok()
        && left.file_type().is_file()
        && right.file_type().is_file()
}

fn validate_id(id: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[0-9a-f]{20}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(id) {
        return Err(AinfraError::guard("invalid reviewed plan ID"));
    }
    Ok(())
}

fn validate_digest(value: &str, label: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[0-9a-f]{64}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(value) {
        return Err(AinfraError::guard(format!(
            "invalid {label} digest in reviewed plan"
        )));
    }
    Ok(())
}

fn validate_name(value: &str, label: &str, minimum: usize) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[a-z][a-z0-9-]{1,62}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if value.len() < minimum || !pattern.is_match(value) {
        return Err(AinfraError::guard(format!(
            "invalid {label} identity in reviewed plan"
        )));
    }
    Ok(())
}

fn binding_file(root: &Path, name: &str) -> Result<PathBuf, AinfraError> {
    let path = root.join(name);
    let metadata = fs::symlink_metadata(&path)
        .map_err(|_| AinfraError::guard(format!("required project file is missing: {name}")))?;
    if metadata.file_type().is_symlink() || !metadata.is_file() {
        return Err(AinfraError::guard(format!(
            "project binding must be a regular non-symlink file: {name}"
        )));
    }
    let resolved = path
        .canonicalize()
        .map_err(|error| AinfraError::guard(format!("cannot resolve {name}: {error}")))?;
    if resolved != path {
        return Err(AinfraError::guard(format!(
            "project binding path is redirected: {name}"
        )));
    }
    Ok(resolved)
}

fn utf8_path(path: &Path) -> Result<String, AinfraError> {
    path.to_str()
        .map(str::to_owned)
        .ok_or_else(|| AinfraError::guard("project binding path is not valid UTF-8"))
}

fn run_root(project_root: &Path) -> PathBuf {
    project_root.join(".ainfra/runs")
}

fn legacy_format_version() -> String {
    FORMAT_VERSION.to_owned()
}
