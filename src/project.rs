//! Strict project configuration, lockfile validation, and safe initialization.

use std::collections::BTreeMap;
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Component, Path, PathBuf};

use regex::Regex;
use serde::{Deserialize, Serialize};

use crate::contracts::validate_path;
use crate::error::AinfraError;
use crate::policy::validate_policy;
use crate::template::discover_builtin;

const API_VERSION: &str = "ainfra.projectious.work/v1alpha1";
const PROJECT_KIND: &str = "AinfraProject";
const LOCK_KIND: &str = "AinfraLock";
const BUILTIN_SOURCE: &str = "builtin";
const EXAMPLE_INPUT: &str =
    include_str!("../templates/hetzner-kubernetes-baseline/inputs/example.input.yaml");
const IGNORE_BLOCK: &str = "# ainfra local operational state\n.ainfra/\n";

/// Human-authored project configuration.
#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ProjectConfig {
    /// Contract API version.
    pub api_version: String,
    /// Contract kind.
    pub kind: String,
    /// Project identity.
    pub metadata: ProjectMetadata,
    /// Desired template and environments.
    pub spec: ProjectSpec,
}

/// Stable project identity.
#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectMetadata {
    /// DNS-label-like project name.
    pub name: String,
}

/// Desired project inputs.
#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectSpec {
    /// Built-in template identity.
    pub template: String,
    /// Named environment input paths.
    pub environments: BTreeMap<String, EnvironmentConfig>,
}

/// One environment input reference.
#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct EnvironmentConfig {
    /// Project-relative `TemplateInput` path.
    pub input: PathBuf,
}

/// Generated template lockfile.
#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ProjectLock {
    /// Contract API version.
    pub api_version: String,
    /// Contract kind.
    pub kind: String,
    /// Exact built-in template binding.
    pub template: TemplateLock,
}

/// Exact template source, version, and content digest.
#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct TemplateLock {
    /// Template source type.
    pub source: String,
    /// Template identity.
    pub name: String,
    /// Template manifest version.
    pub version: String,
    /// Lowercase SHA-256 content digest.
    pub sha256: String,
}

/// Fully validated project boundary.
#[derive(Debug)]
pub struct Project {
    /// Canonical project root.
    pub root: PathBuf,
    /// Parsed project configuration.
    pub config: ProjectConfig,
    /// Parsed and verified lockfile.
    pub lock: ProjectLock,
    /// Canonical environment inputs by name.
    pub inputs: BTreeMap<String, PathBuf>,
}

/// Result returned by safe project initialization.
#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct InitResult {
    /// Machine-output protocol.
    pub api_version: &'static str,
    /// Whether initialization succeeded.
    pub ok: bool,
    /// Canonical initialized directory.
    pub project_root: String,
    /// Exact pinned template.
    pub template: TemplateLock,
    /// Project-relative paths created.
    pub created: Vec<String>,
    /// Project-relative paths updated append-only.
    pub updated: Vec<String>,
    /// Safe next commands.
    pub next: [&'static str; 2],
}

impl Project {
    /// Discover the nearest ancestor project and validate all committed inputs.
    ///
    /// # Errors
    ///
    /// Returns a stable guard or input error for missing or unsafe projects.
    pub fn discover(start: &Path) -> Result<Self, AinfraError> {
        let start = start.canonicalize().map_err(|error| {
            AinfraError::guard(format!(
                "cannot resolve project search path {}: {error}",
                start.display()
            ))
        })?;
        let mut directory = if start.is_dir() {
            start
        } else {
            start
                .parent()
                .ok_or_else(|| AinfraError::guard("project search path has no parent"))?
                .to_path_buf()
        };
        loop {
            let marker = directory.join("ainfra.yaml");
            if fs::symlink_metadata(&marker).is_ok() {
                return Self::load(&directory);
            }
            if !directory.pop() {
                return Err(AinfraError::guard(
                    "no ainfra.yaml found in this directory or its ancestors",
                ));
            }
        }
    }

    /// Load and validate one exact project root.
    ///
    /// # Errors
    ///
    /// Rejects symlinks, malformed contracts, escaped inputs, and stale locks.
    pub fn load(root: &Path) -> Result<Self, AinfraError> {
        let root = root
            .canonicalize()
            .map_err(|error| AinfraError::guard(format!("cannot resolve project root: {error}")))?;
        let config_path = regular_direct_child(&root, "ainfra.yaml")?;
        let lock_path = regular_direct_child(&root, "ainfra.lock")?;
        let config: ProjectConfig = parse_yaml(&config_path, "project configuration")?;
        validate_header(&config.api_version, &config.kind, PROJECT_KIND)?;
        validate_name(&config.metadata.name, "project")?;
        let template = discover_builtin(&config.spec.template)?;
        let lock: ProjectLock = parse_yaml(&lock_path, "project lock")?;
        validate_header(&lock.api_version, &lock.kind, LOCK_KIND)?;
        verify_lock(&lock, &template)?;
        if config.spec.environments.is_empty() {
            return Err(AinfraError::input_contract(
                "project must declare at least one environment",
            ));
        }
        let mut inputs = BTreeMap::new();
        for (name, environment) in &config.spec.environments {
            validate_name(name, "environment")?;
            let input = contained_regular_file(&root, &environment.input)?;
            let document = validate_path(&input)?;
            if document["metadata"]["environment"] != name.as_str() {
                return Err(AinfraError::input_contract(format!(
                    "environment {name:?} does not match {}",
                    input.display()
                )));
            }
            if document["metadata"]["template"] != config.spec.template {
                return Err(AinfraError::input_contract(format!(
                    "environment {name:?} uses another template"
                )));
            }
            validate_policy(&document, &input, Some(template.name), &root)?;
            inputs.insert(name.clone(), input);
        }
        Ok(Self {
            root,
            config,
            lock,
            inputs,
        })
    }
}

/// Initialize a project without overwriting committed configuration.
///
/// # Errors
///
/// Preflights all primary files and fails before writing when any conflict
/// exists.
pub fn initialize(
    root: &Path,
    name: Option<&str>,
    template_name: &str,
    environment: &str,
) -> Result<InitResult, AinfraError> {
    let root = root.canonicalize().map_err(|error| {
        AinfraError::guard(format!(
            "cannot resolve initialization directory {}: {error}",
            root.display()
        ))
    })?;
    let inferred = root
        .file_name()
        .and_then(|value| value.to_str())
        .ok_or_else(|| AinfraError::input_contract("project name must be provided"))?;
    let name = name.unwrap_or(inferred);
    validate_name(name, "project")?;
    validate_name(environment, "environment")?;
    let environment_relative = PathBuf::from("environments").join(format!("{environment}.yaml"));
    let primary = [
        root.join("ainfra.yaml"),
        root.join("ainfra.lock"),
        root.join(&environment_relative),
    ];
    let conflicts: Vec<_> = primary
        .iter()
        .filter(|path| fs::symlink_metadata(path).is_ok())
        .map(|path| {
            path.strip_prefix(&root)
                .unwrap_or(path)
                .display()
                .to_string()
        })
        .collect();
    if !conflicts.is_empty() {
        return Err(AinfraError::guard(format!(
            "refusing to overwrite existing project files: {}",
            conflicts.join(", ")
        )));
    }
    let environments = root.join("environments");
    if let Ok(metadata) = fs::symlink_metadata(&environments)
        && (metadata.file_type().is_symlink() || !metadata.is_dir())
    {
        return Err(AinfraError::guard(
            "environments must be a regular non-symlink directory",
        ));
    }
    let ignore_path = root.join(".gitignore");
    preflight_ignore(&ignore_path)?;
    let (config, project_lock, lock) =
        initialization_documents(name, template_name, environment, &environment_relative)?;
    let config_yaml = serialize_yaml(&config)?;
    let lock_yaml = serialize_yaml(&project_lock)?;
    let created_environments = !environments.exists();
    fs::create_dir_all(&environments).map_err(|error| {
        AinfraError::guard(format!("cannot create environments directory: {error}"))
    })?;
    let documents = [
        (&primary[0], config_yaml.as_bytes()),
        (&primary[1], lock_yaml.as_bytes()),
        (&primary[2], EXAMPLE_INPUT.as_bytes()),
    ];
    let mut created_paths = Vec::new();
    for (path, content) in documents {
        if let Err(error) = create_new(path, content) {
            rollback_created(&created_paths, &environments, created_environments);
            return Err(error);
        }
        created_paths.push(path.clone());
    }
    let mut updated = Vec::new();
    match append_ignore_rule(&ignore_path) {
        Ok(true) => updated.push(".gitignore".to_owned()),
        Ok(false) => {}
        Err(error) => {
            rollback_created(&created_paths, &environments, created_environments);
            return Err(error);
        }
    }
    Ok(InitResult {
        api_version: "ainfra.init/v1alpha1",
        ok: true,
        project_root: root.display().to_string(),
        template: lock,
        created: vec![
            "ainfra.yaml".to_owned(),
            "ainfra.lock".to_owned(),
            environment_relative.display().to_string(),
        ],
        updated,
        next: ["ainfra validate", "ainfra doctor"],
    })
}

fn validate_header(api_version: &str, kind: &str, expected: &str) -> Result<(), AinfraError> {
    if api_version != API_VERSION {
        return Err(AinfraError::input_contract(format!(
            "unsupported apiVersion {api_version:?}"
        )));
    }
    if kind != expected {
        return Err(AinfraError::input_contract(format!(
            "unsupported kind {kind:?}; expected {expected}"
        )));
    }
    Ok(())
}

fn initialization_documents(
    name: &str,
    template_name: &str,
    environment: &str,
    environment_relative: &Path,
) -> Result<(ProjectConfig, ProjectLock, TemplateLock), AinfraError> {
    let template = discover_builtin(template_name)?;
    let version = template.manifest["metadata"]["version"]
        .as_str()
        .ok_or_else(|| AinfraError::dependency("built-in template version is missing"))?;
    let lock = TemplateLock {
        source: BUILTIN_SOURCE.to_owned(),
        name: template.name.to_owned(),
        version: version.to_owned(),
        sha256: template.content_hash(),
    };
    let config = ProjectConfig {
        api_version: API_VERSION.to_owned(),
        kind: PROJECT_KIND.to_owned(),
        metadata: ProjectMetadata {
            name: name.to_owned(),
        },
        spec: ProjectSpec {
            template: template.name.to_owned(),
            environments: BTreeMap::from([(
                environment.to_owned(),
                EnvironmentConfig {
                    input: environment_relative.to_path_buf(),
                },
            )]),
        },
    };
    let project_lock = ProjectLock {
        api_version: API_VERSION.to_owned(),
        kind: LOCK_KIND.to_owned(),
        template: TemplateLock {
            source: lock.source.clone(),
            name: lock.name.clone(),
            version: lock.version.clone(),
            sha256: lock.sha256.clone(),
        },
    };
    Ok((config, project_lock, lock))
}

fn validate_name(name: &str, label: &str) -> Result<(), AinfraError> {
    let pattern = Regex::new(r"^[a-z][a-z0-9-]{1,62}$")
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if !pattern.is_match(name) {
        return Err(AinfraError::input_contract(format!(
            "invalid {label} name {name:?}"
        )));
    }
    Ok(())
}

fn parse_yaml<T: for<'de> Deserialize<'de>>(path: &Path, label: &str) -> Result<T, AinfraError> {
    let content = fs::read_to_string(path).map_err(|error| {
        AinfraError::input_contract(format!("cannot read {}: {error}", path.display()))
    })?;
    serde_yaml::from_str(&content).map_err(|error| {
        AinfraError::input_contract(format!("invalid {label} {}: {error}", path.display()))
    })
}

fn regular_direct_child(root: &Path, name: &str) -> Result<PathBuf, AinfraError> {
    let path = root.join(name);
    let metadata = fs::symlink_metadata(&path)
        .map_err(|_| AinfraError::guard(format!("required project file is missing: {name}")))?;
    if metadata.file_type().is_symlink() || !metadata.is_file() {
        return Err(AinfraError::guard(format!(
            "project file must be a regular non-symlink: {name}"
        )));
    }
    Ok(path)
}

fn contained_regular_file(root: &Path, relative: &Path) -> Result<PathBuf, AinfraError> {
    if relative.is_absolute()
        || relative.components().any(|part| {
            matches!(
                part,
                Component::ParentDir | Component::RootDir | Component::Prefix(_)
            )
        })
    {
        return Err(AinfraError::guard(format!(
            "environment input must be a contained relative path: {}",
            relative.display()
        )));
    }
    let path = root.join(relative);
    let metadata = fs::symlink_metadata(&path).map_err(|_| {
        AinfraError::guard(format!(
            "environment input is missing: {}",
            relative.display()
        ))
    })?;
    if metadata.file_type().is_symlink() || !metadata.is_file() {
        return Err(AinfraError::guard(format!(
            "environment input must be a regular non-symlink: {}",
            relative.display()
        )));
    }
    let resolved = path
        .canonicalize()
        .map_err(|error| AinfraError::guard(error.to_string()))?;
    if !resolved.starts_with(root) {
        return Err(AinfraError::guard("environment input escaped project root"));
    }
    Ok(resolved)
}

fn verify_lock(
    lock: &ProjectLock,
    template: &crate::template::Template,
) -> Result<(), AinfraError> {
    let version = template.manifest["metadata"]["version"]
        .as_str()
        .unwrap_or_default();
    let digest = template.content_hash();
    if lock.template.source != BUILTIN_SOURCE
        || lock.template.name != template.name
        || lock.template.version != version
        || lock.template.sha256 != digest
    {
        return Err(AinfraError::guard(
            "ainfra.lock does not match the selected built-in template",
        ));
    }
    Ok(())
}

fn serialize_yaml<T: Serialize>(value: &T) -> Result<String, AinfraError> {
    serde_yaml::to_string(value)
        .map_err(|error| AinfraError::dependency(format!("cannot serialize YAML: {error}")))
}

fn create_new(path: &Path, content: &[u8]) -> Result<(), AinfraError> {
    let mut file = OpenOptions::new()
        .create_new(true)
        .write(true)
        .open(path)
        .map_err(|error| {
            AinfraError::guard(format!("cannot create {}: {error}", path.display()))
        })?;
    file.write_all(content)
        .map_err(|error| AinfraError::guard(format!("cannot write {}: {error}", path.display())))
}

fn append_ignore_rule(path: &Path) -> Result<bool, AinfraError> {
    preflight_ignore(path)?;
    let existing = match fs::read_to_string(path) {
        Ok(content) => content,
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => String::new(),
        Err(error) => {
            return Err(AinfraError::guard(format!(
                "cannot read {}: {error}",
                path.display()
            )));
        }
    };
    if existing.lines().any(|line| line.trim() == ".ainfra/") {
        return Ok(false);
    }
    let mut replacement = existing.clone();
    if !existing.is_empty() && !existing.ends_with('\n') {
        replacement.push('\n');
    }
    if !existing.is_empty() {
        replacement.push('\n');
    }
    replacement.push_str(IGNORE_BLOCK);
    replace_ignore_atomically(path, replacement.as_bytes())?;
    Ok(true)
}

fn replace_ignore_atomically(path: &Path, content: &[u8]) -> Result<(), AinfraError> {
    let temporary = path.with_file_name(format!(
        ".gitignore.ainfra-{}.tmp",
        uuid::Uuid::new_v4().simple()
    ));
    let result = (|| {
        create_new(&temporary, content)?;
        if let Ok(metadata) = fs::metadata(path) {
            fs::set_permissions(&temporary, metadata.permissions()).map_err(|error| {
                AinfraError::guard(format!("cannot preserve .gitignore permissions: {error}"))
            })?;
        }
        fs::rename(&temporary, path)
            .map_err(|error| AinfraError::guard(format!("cannot update .gitignore: {error}")))
    })();
    if result.is_err() {
        let _ = fs::remove_file(&temporary);
    }
    result
}

fn preflight_ignore(path: &Path) -> Result<(), AinfraError> {
    match fs::symlink_metadata(path) {
        Ok(metadata) if metadata.file_type().is_symlink() || !metadata.is_file() => Err(
            AinfraError::guard(".gitignore must be a regular non-symlink file"),
        ),
        Ok(_) => fs::read_to_string(path)
            .map(|_| ())
            .map_err(|error| AinfraError::guard(format!("cannot read .gitignore: {error}"))),
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(error) => Err(AinfraError::guard(format!(
            "cannot inspect .gitignore: {error}"
        ))),
    }
}

fn rollback_created(paths: &[PathBuf], environments: &Path, remove_directory: bool) {
    for path in paths.iter().rev() {
        let _ = fs::remove_file(path);
    }
    if remove_directory {
        let _ = fs::remove_dir(environments);
    }
}

/// Stable JSON form used by the CLI.
///
/// # Errors
///
/// Returns a dependency error if serialization fails.
pub fn init_json(result: &InitResult) -> Result<String, AinfraError> {
    serde_json::to_string(result).map_err(|error| AinfraError::dependency(error.to_string()))
}
