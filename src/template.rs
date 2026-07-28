//! Built-in template discovery and compatibility validation.

use std::collections::BTreeSet;
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Component, Path, PathBuf};
use std::str::FromStr;

use include_dir::{Dir, File, include_dir};
use pep440_rs::{Version, VersionSpecifiers};
use regex::Regex;
use serde_json::Value;
use sha2::{Digest, Sha256};

use crate::contracts::validate_document;
use crate::error::AinfraError;

const BASELINE_NAME: &str = "hetzner-kubernetes-baseline";
const BASELINE_MANIFEST: &str =
    include_str!("../templates/hetzner-kubernetes-baseline/ainfra-template.yaml");
static BASELINE_DIRECTORY: Dir<'_> =
    include_dir!("$CARGO_MANIFEST_DIR/templates/hetzner-kubernetes-baseline");
const COMPATIBILITY_VERSION: &str = "0.1.0";
const SUPPORTED_CAPABILITIES: [&str; 5] = [
    "access.ssh",
    "network.private",
    "provider.hetzner-cloud",
    "security.hardening",
    "target.kubernetes-ready",
];

/// One template embedded in the installed binary.
#[derive(Debug)]
pub struct Template {
    /// Stable template identity.
    pub name: &'static str,
    /// Virtual resource path used in deterministic diagnostics.
    pub manifest_path: &'static str,
    /// Validated manifest document.
    pub manifest: Value,
}

impl Template {
    /// Return the Python-compatible digest of the embedded template tree.
    #[must_use]
    pub fn content_hash(&self) -> String {
        let mut files = Vec::new();
        embedded_files(&BASELINE_DIRECTORY, &mut files);
        files.retain(|file| !embedded_excluded(file.path()));
        files.sort_by_key(|file| file.path());
        let mut digest = Sha256::new();
        for file in files {
            digest.update(file.path().to_string_lossy().replace('\\', "/").as_bytes());
            digest.update(file.contents());
        }
        format!("{:x}", digest.finalize())
    }

    /// Materialize the embedded, immutable template into a new directory.
    ///
    /// # Errors
    ///
    /// Returns a guard error rather than overwriting or escaping the target.
    pub fn materialize(&self, destination: &Path) -> Result<(), AinfraError> {
        fs::create_dir(destination).map_err(|error| {
            AinfraError::guard(format!(
                "cannot create template workspace {}: {error}",
                destination.display()
            ))
        })?;
        let mut files = Vec::new();
        embedded_files(&BASELINE_DIRECTORY, &mut files);
        files.retain(|file| !embedded_excluded(file.path()));
        for file in files {
            let relative = file.path();
            if relative.is_absolute()
                || relative.components().any(|component| {
                    matches!(component, Component::ParentDir | Component::Prefix(_))
                })
            {
                return Err(AinfraError::guard(
                    "embedded template contains an unsafe path",
                ));
            }
            let output = destination.join(relative);
            if let Some(parent) = output.parent() {
                fs::create_dir_all(parent).map_err(|error| {
                    AinfraError::guard(format!(
                        "cannot create template directory {}: {error}",
                        parent.display()
                    ))
                })?;
            }
            let mut target = OpenOptions::new()
                .create_new(true)
                .write(true)
                .open(&output)
                .map_err(|error| {
                    AinfraError::guard(format!("cannot materialize {}: {error}", output.display()))
                })?;
            target.write_all(file.contents()).map_err(|error| {
                AinfraError::guard(format!(
                    "cannot write template file {}: {error}",
                    output.display()
                ))
            })?;
        }
        Ok(())
    }
}

fn embedded_files<'a>(directory: &'a Dir<'a>, files: &mut Vec<&'a File<'a>>) {
    files.extend(directory.files());
    for child in directory.dirs() {
        embedded_files(child, files);
    }
}

fn embedded_excluded(path: &Path) -> bool {
    path.components().any(|part| {
        matches!(
            part.as_os_str().to_str(),
            Some(".ainfra" | ".terraform" | "__pycache__")
        )
    })
}

/// Discover and validate a built-in template without a source checkout.
///
/// # Errors
///
/// Returns a stable contract or safety error for invalid names, missing
/// templates, incompatible versions, capabilities, or declared paths.
pub fn discover_builtin(name: &str) -> Result<Template, AinfraError> {
    let name_pattern = Regex::new(r"^[a-z][a-z0-9-]{2,62}$").map_err(|error| {
        AinfraError::dependency(format!(
            "embedded template-name pattern is invalid: {error}"
        ))
    })?;
    if !name_pattern.is_match(name) {
        return Err(AinfraError::input_contract(format!(
            "invalid template name {name:?}"
        )));
    }
    if name != BASELINE_NAME {
        return Err(AinfraError::input_contract(format!(
            "template {name:?} does not exist"
        )));
    }
    let manifest: Value = serde_yaml::from_str(BASELINE_MANIFEST).map_err(|error| {
        AinfraError::dependency(format!("embedded template is invalid: {error}"))
    })?;
    let manifest_path = "templates/hetzner-kubernetes-baseline/ainfra-template.yaml";
    validate_document(&manifest, Path::new(manifest_path))?;
    if manifest["metadata"]["name"] != name {
        return Err(AinfraError::safety(format!(
            "[P003] manifest name {:?} does not match {name:?}",
            manifest["metadata"]["name"]
        )));
    }
    validate_manifest_compatibility(&manifest)?;
    validate_virtual_paths(&manifest, name)?;
    Ok(Template {
        name: BASELINE_NAME,
        manifest_path,
        manifest,
    })
}

/// Validate the PEP 440 wrapper range and capability allowlist.
///
/// # Errors
///
/// Returns `P008` or `P009` safety errors.
pub fn validate_manifest_compatibility(manifest: &Value) -> Result<(), AinfraError> {
    let raw_range = manifest["spec"]["requiresWrapper"]
        .as_str()
        .unwrap_or_default();
    let normalized = raw_range.split_whitespace().collect::<Vec<_>>().join(",");
    let specifiers = VersionSpecifiers::from_str(&normalized).map_err(|_| {
        AinfraError::safety(format!(
            "[P008] invalid wrapper version range {raw_range:?}"
        ))
    })?;
    let running = Version::from_str(COMPATIBILITY_VERSION).map_err(|error| {
        AinfraError::dependency(format!(
            "embedded compatibility version is invalid: {error}"
        ))
    })?;
    if !specifiers.contains(&running) {
        return Err(AinfraError::safety(format!(
            "[P008] wrapper {running} does not satisfy {raw_range:?}"
        )));
    }
    let supported: BTreeSet<_> = SUPPORTED_CAPABILITIES.into_iter().collect();
    let unknown: BTreeSet<_> = manifest["spec"]["capabilities"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(Value::as_str)
        .filter(|capability| !supported.contains(capability))
        .collect();
    if !unknown.is_empty() {
        return Err(AinfraError::safety(format!(
            "[P009] unsupported capabilities: {}",
            unknown.into_iter().collect::<Vec<_>>().join(", ")
        )));
    }
    Ok(())
}

fn validate_virtual_paths(manifest: &Value, name: &str) -> Result<(), AinfraError> {
    let template_root = PathBuf::from("templates").join(name);
    for engine in ["tofu", "ansible"] {
        let path = manifest["spec"]["engines"][engine]["workingDirectory"]
            .as_str()
            .unwrap_or_default();
        contained_virtual_path(&template_root, path, &template_root, "P004")?;
    }
    if let Some(examples) = manifest["spec"]["inputs"]["examples"].as_array() {
        for example in examples.iter().filter_map(Value::as_str) {
            contained_virtual_path(&template_root, example, &template_root, "P005")?;
        }
    }
    let output = manifest["spec"]["output"]["file"]
        .as_str()
        .unwrap_or_default();
    contained_virtual_path(&template_root, output, &template_root, "P006")?;

    for section in ["inputs", "output"] {
        let schema = manifest["spec"][section]["schema"]
            .as_str()
            .unwrap_or_default();
        let resolved =
            contained_virtual_path(&template_root, schema, Path::new("schemas"), "P007")?;
        if resolved.parent() != Some(Path::new("schemas")) {
            return Err(AinfraError::safety(
                "[P007] schema must be a direct child of schemas",
            ));
        }
    }
    Ok(())
}

fn contained_virtual_path(
    base: &Path,
    value: &str,
    allowed_root: &Path,
    rule: &str,
) -> Result<PathBuf, AinfraError> {
    let resolved = normalize_virtual(&base.join(value)).ok_or_else(|| {
        AinfraError::safety(format!(
            "[{rule}] path escapes {}: {value}",
            allowed_root.display()
        ))
    })?;
    if !resolved.starts_with(allowed_root) {
        return Err(AinfraError::safety(format!(
            "[{rule}] path escapes {}: {value}",
            allowed_root.display()
        )));
    }
    Ok(resolved)
}

fn normalize_virtual(path: &Path) -> Option<PathBuf> {
    let mut normalized = PathBuf::new();
    for component in path.components() {
        match component {
            Component::Normal(value) => normalized.push(value),
            Component::CurDir => {}
            Component::ParentDir => {
                if !normalized.pop() {
                    return None;
                }
            }
            Component::RootDir | Component::Prefix(_) => return None,
        }
    }
    Some(normalized)
}
