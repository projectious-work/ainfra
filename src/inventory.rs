//! Deterministic Ansible inventory generation.

use std::collections::BTreeMap;
use std::fs;
use std::path::Path;

use serde::Serialize;

use crate::contracts::validate_path;
use crate::error::AinfraError;
use crate::policy::validate_policy;

#[derive(Debug, Serialize)]
struct Inventory {
    all: All,
}

#[derive(Debug, Serialize)]
struct All {
    vars: Variables,
    children: Children,
}

#[derive(Debug, Serialize)]
struct Variables {
    ansible_user: &'static str,
    ansible_python_interpreter: &'static str,
    ainfra_private_cidr: String,
}

#[derive(Debug, Serialize)]
struct Children {
    control_plane: Group,
    workers: Group,
}

#[derive(Debug, Default, Serialize)]
struct Group {
    hosts: BTreeMap<String, Host>,
}

#[derive(Debug, Serialize)]
struct Host {
    #[serde(rename = "ansible_host")]
    address: String,
    #[serde(rename = "ainfra_image")]
    image: String,
    #[serde(rename = "ainfra_role")]
    role: String,
}

/// Build a validated, secret-free inventory as YAML.
///
/// # Errors
///
/// Returns a stable contract, policy, or guard error.
pub fn build_inventory(output_path: &Path, project_root: &Path) -> Result<String, AinfraError> {
    let document = validate_path(output_path)?;
    if document["kind"] != "InfrastructureOutput" {
        return Err(AinfraError::guard(
            "inventory requires an InfrastructureOutput document",
        ));
    }
    validate_policy(&document, output_path, None, project_root)?;
    let mut control_plane = Group::default();
    let mut workers = Group::default();
    let mut nodes = document["spec"]["nodes"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    nodes.sort_by_key(|node| node["name"].as_str().unwrap_or_default().to_owned());
    for node in nodes {
        let name = node["name"].as_str().unwrap_or_default().to_owned();
        let host = Host {
            address: node["privateIPv4"].as_str().unwrap_or_default().to_owned(),
            image: node["image"].as_str().unwrap_or_default().to_owned(),
            role: node["role"].as_str().unwrap_or_default().to_owned(),
        };
        if node["role"] == "control-plane-capable" {
            control_plane.hosts.insert(name, host);
        } else {
            workers.hosts.insert(name, host);
        }
    }
    let private_cidr = document["spec"]["network"]["privateCidrs"][0]
        .as_str()
        .unwrap_or_default()
        .to_owned();
    serde_yaml::to_string(&Inventory {
        all: All {
            vars: Variables {
                ansible_user: "ainfra",
                ansible_python_interpreter: "/usr/bin/python3",
                ainfra_private_cidr: private_cidr,
            },
            children: Children {
                control_plane,
                workers,
            },
        },
    })
    .map_err(|error| AinfraError::dependency(error.to_string()))
}

/// Write inventory atomically through a sibling temporary file.
///
/// # Errors
///
/// Returns a stable validation or filesystem error.
pub fn write_inventory(
    output_path: &Path,
    destination: &Path,
    project_root: &Path,
) -> Result<(), AinfraError> {
    let inventory = build_inventory(output_path, project_root)?;
    if let Some(parent) = destination.parent() {
        fs::create_dir_all(parent).map_err(|error| AinfraError::dependency(error.to_string()))?;
    }
    let temporary_extension = destination.extension().map_or_else(
        || "tmp".to_owned(),
        |suffix| format!("{}.tmp", suffix.to_string_lossy()),
    );
    let temporary = destination.with_extension(temporary_extension);
    fs::write(&temporary, inventory).map_err(|error| AinfraError::dependency(error.to_string()))?;
    fs::rename(&temporary, destination).map_err(|error| AinfraError::dependency(error.to_string()))
}
