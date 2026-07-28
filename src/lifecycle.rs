//! Guarded `OpenTofu` plan orchestration.

use std::collections::BTreeMap;
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

use serde_json::{Value, json};

use crate::contracts::validate_path;
use crate::error::AinfraError;
use crate::plan_record::{Operation, PlanRecord, template_hash};
use crate::policy::validate_policy;
use crate::process::{ProcessRequest, Runner, child_environment, require_success};
use crate::template::discover_builtin;

/// Create one isolated, immutable reviewed plan.
///
/// # Errors
///
/// Returns stable contract, policy, dependency, or guard errors. A failure
/// never invokes apply or destroy.
pub fn plan(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    input_path: &Path,
    operation: Operation,
) -> Result<PlanRecord, AinfraError> {
    let project_root = project_root.canonicalize().map_err(|error| {
        AinfraError::guard(format!(
            "cannot resolve project root {}: {error}",
            project_root.display()
        ))
    })?;
    let template = discover_builtin(template_name)?;
    let document = validate_path(input_path)?;
    validate_policy(&document, input_path, Some(template.name), &project_root)?;
    let environment_name = document["metadata"]["environment"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("missing environment"))?;
    let template_version = template.manifest["metadata"]["version"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("missing template version"))?;
    let mut record = PlanRecord::create_with_template_hash(
        &project_root,
        operation,
        template.name,
        template_version,
        environment_name,
        input_path,
        &template.content_hash(),
    )?;
    let run = PathBuf::from(&record.plan_path)
        .parent()
        .map(Path::to_path_buf)
        .ok_or_else(|| AinfraError::guard("plan path has no run directory"))?;
    let runs = run
        .parent()
        .ok_or_else(|| AinfraError::guard("run path has no parent"))?;
    fs::create_dir_all(runs).map_err(|error| {
        AinfraError::guard(format!("cannot create {}: {error}", runs.display()))
    })?;
    let resolved_runs = runs.canonicalize().map_err(|error| {
        AinfraError::guard(format!("cannot resolve {}: {error}", runs.display()))
    })?;
    if !resolved_runs.starts_with(&project_root) {
        return Err(AinfraError::guard("run directory escaped the project root"));
    }
    fs::create_dir(&run).map_err(|error| {
        AinfraError::guard(format!("cannot allocate {}: {error}", run.display()))
    })?;
    let workspace = run.join("workspace");
    template.materialize(&workspace)?;
    if template_hash(&workspace)? != record.template_sha256 {
        return Err(AinfraError::guard(
            "materialized template digest does not match embedded template",
        ));
    }
    let variables = run.join("input.auto.tfvars.json");
    write_new_json(&variables, &tofu_variables(&document))?;
    let (environment, secrets) =
        lifecycle_environment(&document, &project_root, environment_source)?;
    let tofu_root = workspace.join(
        template.manifest["spec"]["engines"]["tofu"]["workingDirectory"]
            .as_str()
            .ok_or_else(|| AinfraError::input_contract("missing tofu working directory"))?,
    );
    let mut init = vec![
        "tofu".to_owned(),
        "init".to_owned(),
        "-input=false".to_owned(),
        "-lockfile=readonly".to_owned(),
        "-reconfigure".to_owned(),
    ];
    init.extend(backend_args(&document, &project_root, environment_source)?);
    require_success(runner.run(&ProcessRequest {
        argv: init,
        cwd: tofu_root.clone(),
        environment: environment.clone(),
        secrets: secrets.clone(),
    })?)?;
    let mut plan_argv = vec![
        "tofu".to_owned(),
        "plan".to_owned(),
        "-input=false".to_owned(),
    ];
    if operation == Operation::Destroy {
        plan_argv.push("-destroy".to_owned());
    }
    plan_argv.extend([
        format!("-out={}", record.plan_path),
        format!("-var-file={}", variables.display()),
    ]);
    require_success(runner.run(&ProcessRequest {
        argv: plan_argv,
        cwd: tofu_root,
        environment,
        secrets,
    })?)?;
    record.bind_plan(&project_root)?;
    record.write(&project_root)?;
    Ok(record)
}

/// Read-only environment boundary used for secret references.
pub trait EnvironmentSource {
    /// Return one environment value without exposing it through diagnostics.
    fn get(&self, name: &str) -> Option<String>;
}

/// Real process-environment source.
pub struct OsEnvironment;

impl EnvironmentSource for OsEnvironment {
    fn get(&self, name: &str) -> Option<String> {
        std::env::var(name).ok()
    }
}

fn write_new_json(path: &Path, document: &Value) -> Result<(), AinfraError> {
    let mut file = OpenOptions::new()
        .create_new(true)
        .write(true)
        .open(path)
        .map_err(|error| {
            AinfraError::guard(format!("cannot create {}: {error}", path.display()))
        })?;
    let content = serde_json::to_string_pretty(document)
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    file.write_all(content.as_bytes())
        .map_err(|error| AinfraError::guard(format!("cannot write {}: {error}", path.display())))
}

fn tofu_variables(document: &Value) -> Value {
    let spec = &document["spec"];
    json!({
        "admin_ssh_public_keys": spec["access"]["adminSshPublicKeys"],
        "control_plane_count": spec["topology"]["controlPlaneCount"],
        "environment": document["metadata"]["environment"],
        "image": spec["topology"]["image"],
        "location": spec["provider"]["location"],
        "management_ingress_cidrs":
            spec["network"]["managementIngressCidrs"],
        "private_cidr": spec["network"]["privateCidr"],
        "public_ipv4": spec["network"]["publicIPv4"],
        "public_ipv6": spec["network"]["publicIPv6"],
        "server_type": spec["topology"]["serverType"],
        "worker_count": spec["topology"]["workerCount"],
        "workload_ingress_cidrs":
            spec["network"]["workloadIngressCidrs"],
    })
}

fn lifecycle_environment(
    document: &Value,
    project_root: &Path,
    environment_source: &dyn EnvironmentSource,
) -> Result<(BTreeMap<String, String>, Vec<String>), AinfraError> {
    let mut environment = child_environment();
    let mut secrets = Vec::new();
    let provider = &document["spec"]["provider"]["projectTokenRef"];
    resolve_reference(
        provider,
        project_root,
        "HCLOUD_TOKEN",
        &mut environment,
        &mut secrets,
        environment_source,
    )?;
    let state = &document["spec"]["state"];
    if state["mode"] == "remote" && state["backendConfigRef"]["type"] == "environment" {
        let name = state["backendConfigRef"]["name"]
            .as_str()
            .ok_or_else(|| AinfraError::input_contract("invalid backend reference"))?;
        let value = environment_source.get(name).ok_or_else(|| {
            AinfraError::dependency(format!("referenced environment variable is unset: {name}"))
        })?;
        environment.insert(name.to_owned(), value.clone());
        secrets.push(value);
    }
    Ok((environment, secrets))
}

fn resolve_reference(
    reference: &Value,
    project_root: &Path,
    target_name: &str,
    environment: &mut BTreeMap<String, String>,
    secrets: &mut Vec<String>,
    environment_source: &dyn EnvironmentSource,
) -> Result<(), AinfraError> {
    let name = reference["name"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("invalid credential reference"))?;
    let value = if reference["type"] == "environment" {
        environment_source.get(name).ok_or_else(|| {
            AinfraError::dependency(format!("referenced environment variable is unset: {name}"))
        })?
    } else {
        let path = project_root.join(name);
        let value = fs::read_to_string(&path).map_err(|_| {
            AinfraError::dependency(format!(
                "provider credential file is missing: {}",
                path.display()
            ))
        })?;
        let value = value.trim().to_owned();
        if value.is_empty() {
            return Err(AinfraError::dependency(format!(
                "provider credential file is empty: {}",
                path.display()
            )));
        }
        value
    };
    environment.insert(target_name.to_owned(), value.clone());
    secrets.push(value);
    Ok(())
}

fn backend_args(
    document: &Value,
    project_root: &Path,
    environment_source: &dyn EnvironmentSource,
) -> Result<Vec<String>, AinfraError> {
    let state = &document["spec"]["state"];
    if state["mode"] != "remote" {
        return Ok(Vec::new());
    }
    let reference = &state["backendConfigRef"];
    let path = if reference["type"] == "environment" {
        let name = reference["name"]
            .as_str()
            .ok_or_else(|| AinfraError::input_contract("invalid backend reference"))?;
        PathBuf::from(environment_source.get(name).ok_or_else(|| {
            AinfraError::dependency(format!(
                "backend configuration path environment reference is unset: {name}"
            ))
        })?)
    } else {
        project_root.join(
            reference["name"]
                .as_str()
                .ok_or_else(|| AinfraError::input_contract("invalid backend reference"))?,
        )
    };
    let resolved = path.canonicalize().map_err(|_| {
        AinfraError::dependency(format!(
            "backend configuration file is missing: {}",
            path.display()
        ))
    })?;
    if reference["type"] == "local-file"
        && !resolved.starts_with(
            project_root
                .join(".ainfra")
                .canonicalize()
                .map_err(|error| AinfraError::guard(format!("cannot resolve .ainfra: {error}")))?,
        )
    {
        return Err(AinfraError::guard(
            "backend configuration escaped the .ainfra directory",
        ));
    }
    if !resolved.is_file() {
        return Err(AinfraError::dependency(format!(
            "backend configuration file is missing: {}",
            resolved.display()
        )));
    }
    Ok(vec![format!("-backend-config={}", resolved.display())])
}
