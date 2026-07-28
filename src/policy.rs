//! Semantic safety policy for validated contract documents.

use std::path::{Component, Path};

use ipnet::IpNet;
use regex::Regex;
use serde_json::Value;

use crate::error::AinfraError;

const SECRET_KEY_PATTERN: &str = r"(?i)(?:password|private[_-]?key|secret|token|kubeconfig)";
const SECRET_VALUE_PATTERN: &str = concat!(
    r"(?:-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----",
    r"|hcloud_[A-Za-z0-9]{16,}",
    r"|gh[opsu]_[A-Za-z0-9]{20,}",
    r"|AINFRA_TEST_SECRET_DO_NOT_USE)",
);

/// Apply semantic safety rules after structural contract validation.
///
/// # Errors
///
/// Returns `AINFRA-E400` with a stable policy identifier when a rule fails.
pub fn validate_policy(
    document: &Value,
    source: &Path,
    template_name: Option<&str>,
    project_root: &Path,
) -> Result<(), AinfraError> {
    let secret_key = compile_regex(SECRET_KEY_PATTERN)?;
    let secret_value = compile_regex(SECRET_VALUE_PATTERN)?;
    match document.get("kind").and_then(Value::as_str) {
        Some("TemplateInput") => {
            reject_secret_content(document, source, "", &secret_key, &secret_value)?;
            validate_input(document, source, template_name, project_root)
        }
        Some("InfrastructureOutput") => {
            reject_secret_content(document, source, "", &secret_key, &secret_value)?;
            validate_output(document, source, template_name)
        }
        _ => Ok(()),
    }
}

fn validate_input(
    document: &Value,
    source: &Path,
    template_name: Option<&str>,
    project_root: &Path,
) -> Result<(), AinfraError> {
    let metadata = &document["metadata"];
    let spec = &document["spec"];
    let state = &spec["state"];
    if template_name.is_some_and(|name| metadata["template"] != name) {
        return Err(policy_error(
            "P100",
            source,
            "/metadata/template",
            "input template does not match selected template",
        ));
    }
    let disposable = metadata["disposable"].as_bool().unwrap_or(false);
    let mode = state["mode"].as_str().unwrap_or_default();
    if mode == "local-disposable" && !disposable {
        return Err(policy_error(
            "P101",
            source,
            "/spec/state/mode",
            "local state is allowed only for disposable environments",
        ));
    }
    if !disposable && mode != "remote" {
        return Err(policy_error(
            "P102",
            source,
            "/spec/state/mode",
            "non-disposable environments require remote state",
        ));
    }
    if mode == "remote" {
        if state.get("backendConfigRef").is_none() {
            return Err(policy_error(
                "P103",
                source,
                "/spec/state/backendConfigRef",
                "remote state requires an external backend reference",
            ));
        }
        if state.get("capabilities").is_none() {
            return Err(policy_error(
                "P104",
                source,
                "/spec/state/capabilities",
                "remote state requires declared safety capabilities",
            ));
        }
    }
    validate_reference(
        &spec["provider"]["projectTokenRef"],
        source,
        "/spec/provider/projectTokenRef",
        project_root,
    )?;
    if let Some(reference) = state.get("backendConfigRef") {
        validate_reference(
            reference,
            source,
            "/spec/state/backendConfigRef",
            project_root,
        )?;
    }
    if let Some(cidrs) = spec["network"]["managementIngressCidrs"].as_array() {
        for (index, value) in cidrs.iter().enumerate() {
            let Some(cidr) = value.as_str() else {
                continue;
            };
            let Ok(network) = cidr.parse::<IpNet>() else {
                continue;
            };
            let minimum = if network.addr().is_ipv4() { 24 } else { 64 };
            if network.prefix_len() < minimum {
                return Err(policy_error(
                    "P105",
                    source,
                    &format!("/spec/network/managementIngressCidrs/{index}"),
                    &format!("management CIDR must be /{minimum} or narrower"),
                ));
            }
        }
    }
    Ok(())
}

fn validate_output(
    document: &Value,
    source: &Path,
    template_name: Option<&str>,
) -> Result<(), AinfraError> {
    if template_name.is_some_and(|name| document["metadata"]["template"] != name) {
        return Err(policy_error(
            "P200",
            source,
            "/metadata/template",
            "output template does not match selected template",
        ));
    }
    if let Some(nodes) = document["spec"]["nodes"].as_array() {
        for (index, node) in nodes.iter().enumerate() {
            if node["image"] != "debian-13" {
                return Err(policy_error(
                    "P201",
                    source,
                    &format!("/spec/nodes/{index}/image"),
                    "only debian-13 is supported",
                ));
            }
        }
    }
    Ok(())
}

fn validate_reference(
    reference: &Value,
    source: &Path,
    pointer: &str,
    project_root: &Path,
) -> Result<(), AinfraError> {
    let name = reference["name"].as_str().unwrap_or_default();
    if reference["type"] == "environment" {
        if !compile_regex(r"^[A-Z][A-Z0-9_]*$")?.is_match(name) {
            return Err(policy_error(
                "P110",
                source,
                &format!("{pointer}/name"),
                "environment reference must be an uppercase variable name",
            ));
        }
        return Ok(());
    }
    let path = Path::new(name);
    if path.is_absolute()
        || path
            .components()
            .any(|component| component == Component::ParentDir)
    {
        return Err(policy_error(
            "P111",
            source,
            &format!("{pointer}/name"),
            "local-file reference must be relative and contain no '..'",
        ));
    }
    let allowed_root = canonicalize_allow_missing(&project_root.join(".ainfra"))?;
    let resolved = canonicalize_allow_missing(&project_root.join(path))?;
    if !resolved.starts_with(&allowed_root) {
        return Err(policy_error(
            "P112",
            source,
            &format!("{pointer}/name"),
            "local-file reference must stay beneath .ainfra/",
        ));
    }
    Ok(())
}

fn canonicalize_allow_missing(path: &Path) -> Result<std::path::PathBuf, AinfraError> {
    let mut existing = path;
    let mut missing = Vec::new();
    while !existing.exists() {
        let Some(name) = existing.file_name() else {
            break;
        };
        missing.push(name.to_os_string());
        let Some(parent) = existing.parent() else {
            break;
        };
        existing = parent;
    }
    let mut resolved = existing.canonicalize().map_err(|error| {
        AinfraError::dependency(format!("cannot resolve {}: {error}", path.display()))
    })?;
    for name in missing.iter().rev() {
        resolved.push(name);
    }
    Ok(resolved)
}

fn reject_secret_content(
    value: &Value,
    source: &Path,
    pointer: &str,
    secret_key: &Regex,
    secret_value: &Regex,
) -> Result<(), AinfraError> {
    match value {
        Value::Object(object) => {
            for (key, child) in object {
                let child_pointer = format!("{pointer}/{key}");
                if secret_key.is_match(key) && !key.to_lowercase().ends_with("ref") {
                    return Err(policy_error(
                        "P300",
                        source,
                        &child_pointer,
                        "secret-bearing fields are prohibited; use a reference",
                    ));
                }
                reject_secret_content(child, source, &child_pointer, secret_key, secret_value)?;
            }
        }
        Value::Array(items) => {
            for (index, child) in items.iter().enumerate() {
                reject_secret_content(
                    child,
                    source,
                    &format!("{pointer}/{index}"),
                    secret_key,
                    secret_value,
                )?;
            }
        }
        Value::String(text) if secret_value.is_match(text) => {
            return Err(policy_error(
                "P301",
                source,
                if pointer.is_empty() {
                    "<root>"
                } else {
                    pointer
                },
                "value resembles secret material",
            ));
        }
        _ => {}
    }
    Ok(())
}

fn compile_regex(pattern: &str) -> Result<Regex, AinfraError> {
    Regex::new(pattern).map_err(|error| {
        AinfraError::dependency(format!("embedded policy pattern is invalid: {error}"))
    })
}

fn policy_error(rule: &str, source: &Path, pointer: &str, message: &str) -> AinfraError {
    AinfraError::safety(format!(
        "[{rule}] {}:{pointer}: {message}",
        source.display()
    ))
}
