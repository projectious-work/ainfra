//! Guarded `OpenTofu` plan orchestration.

use std::collections::BTreeMap;
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

use serde_json::{Value, json};

use crate::contracts::validate_path;
use crate::error::AinfraError;
use crate::plan_record::{
    ExpectedBindings, Operation, PlanRecord, ProjectBinding, file_hash, template_hash,
};
use crate::policy::validate_policy;
use crate::process::{ProcessRequest, ProcessResult, Runner, child_environment, require_success};
use crate::project::Project;
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
    plan_bound(
        runner,
        environment_source,
        project_root,
        template_name,
        input_path,
        operation,
        None,
    )
}

/// Create one project-bound reviewed plan for an exact configured environment.
///
/// # Errors
///
/// Rejects missing, changed, or ambiguous project inputs before tool execution.
pub fn plan_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    operation: Operation,
) -> Result<PlanRecord, AinfraError> {
    let selection = project_selection(project_root, environment_name)?;
    plan_bound(
        runner,
        environment_source,
        &selection.binding.root,
        &selection.template,
        &selection.input,
        operation,
        Some(&selection.binding),
    )
}

fn plan_bound(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    input_path: &Path,
    operation: Operation,
    project: Option<&ProjectBinding>,
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
    if let Some(binding) = project {
        record.bind_project(binding)?;
    }
    let run = project_root.join(".ainfra/runs").join(&record.id);
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
    let backend = backend_binding(&document, &project_root, environment_source)?;
    record.backend_config_path = backend
        .as_ref()
        .map(|binding| binding.path.display().to_string());
    record.backend_config_sha256 = backend.as_ref().map(|binding| binding.sha256.clone());
    let (environment, secrets) =
        lifecycle_environment(&document, &project_root, environment_source)?;
    let tofu_root = workspace.join(
        template.manifest["spec"]["engines"]["tofu"]["workingDirectory"]
            .as_str()
            .ok_or_else(|| AinfraError::input_contract("missing tofu working directory"))?,
    );
    require_project_binding(&project_root, project)?;
    initialize(runner, &tofu_root, &environment, &secrets, backend.as_ref())?;
    require_project_binding(&project_root, project)?;
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
    require_project_binding(&project_root, project)?;
    record.bind_plan(&project_root)?;
    record.write(&project_root)?;
    Ok(record)
}

/// Apply one exact reviewed apply plan.
///
/// # Errors
///
/// Returns a guard error before process execution when any binding changed.
pub fn apply(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    input_path: &Path,
    approval: &str,
) -> Result<ProcessResult, AinfraError> {
    execute(
        runner,
        environment_source,
        project_root,
        template_name,
        input_path,
        approval,
        Operation::Apply,
        None,
    )
}

/// Apply one exact reviewed destroy plan.
///
/// # Errors
///
/// Returns a guard error before process execution when any binding changed.
pub fn destroy(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    input_path: &Path,
    approval: &str,
) -> Result<ProcessResult, AinfraError> {
    execute(
        runner,
        environment_source,
        project_root,
        template_name,
        input_path,
        approval,
        Operation::Destroy,
        None,
    )
}

/// Apply an exact project-bound reviewed apply plan.
///
/// # Errors
///
/// Rejects changed project files or an environment mismatch before execution.
pub fn apply_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    approval: &str,
) -> Result<ProcessResult, AinfraError> {
    execute_project(
        runner,
        environment_source,
        project_root,
        environment_name,
        approval,
        Operation::Apply,
    )
}

/// Apply an exact project-bound reviewed destroy plan.
///
/// # Errors
///
/// Rejects changed project files or an environment mismatch before execution.
pub fn destroy_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    approval: &str,
) -> Result<ProcessResult, AinfraError> {
    execute_project(
        runner,
        environment_source,
        project_root,
        environment_name,
        approval,
        Operation::Destroy,
    )
}

fn execute_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    approval: &str,
    operation: Operation,
) -> Result<ProcessResult, AinfraError> {
    let selection = project_selection(project_root, environment_name)?;
    execute(
        runner,
        environment_source,
        &selection.binding.root,
        &selection.template,
        &selection.input,
        approval,
        operation,
        Some(&selection.binding),
    )
}

/// Collect and validate standardized output for one exact apply run.
///
/// # Errors
///
/// Returns a guard error for destroy plans or changed bindings, and a contract
/// error for malformed, sensitive, or inconsistent engine output.
pub fn collect_output(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    run_id: &str,
) -> Result<Value, AinfraError> {
    collect_output_bound(
        runner,
        environment_source,
        project_root,
        template_name,
        run_id,
        None,
        None,
    )
}

/// Collect output for one exact project-bound applied run.
///
/// # Errors
///
/// Rejects changed project files or an environment mismatch before `tofu`.
pub fn collect_output_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    run_id: &str,
) -> Result<Value, AinfraError> {
    let selection = project_selection(project_root, environment_name)?;
    collect_output_bound(
        runner,
        environment_source,
        &selection.binding.root,
        &selection.template,
        run_id,
        Some(&selection.binding),
        Some(environment_name),
    )
}

fn collect_output_bound(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    run_id: &str,
    project: Option<&ProjectBinding>,
    selected_environment: Option<&str>,
) -> Result<Value, AinfraError> {
    let project_root = project_root.canonicalize().map_err(|error| {
        AinfraError::guard(format!(
            "cannot resolve project root {}: {error}",
            project_root.display()
        ))
    })?;
    let record = PlanRecord::load(&project_root, run_id)?;
    if record.operation != Operation::Apply {
        return Err(AinfraError::guard(
            "standardized output requires an applied plan",
        ));
    }
    if selected_environment.is_some_and(|selected| selected != record.environment) {
        return Err(AinfraError::guard(
            "reviewed plan belongs to another environment",
        ));
    }
    let input_path = PathBuf::from(&record.input_path);
    let template = discover_builtin(template_name)?;
    let document = validate_path(&input_path)?;
    validate_policy(&document, &input_path, Some(template.name), &project_root)?;
    let run = project_root.join(".ainfra/runs").join(run_id);
    let workspace = run.join("workspace");
    let backend = backend_binding(&document, &project_root, environment_source)?;
    record.verify(
        &project_root,
        &ExpectedBindings {
            template: template.name,
            template_version: template.manifest["metadata"]["version"]
                .as_str()
                .ok_or_else(|| AinfraError::input_contract("missing template version"))?,
            environment: document["metadata"]["environment"]
                .as_str()
                .ok_or_else(|| AinfraError::input_contract("missing environment"))?,
            input_path: &input_path,
            template_root: &workspace,
            operation: Operation::Apply,
            backend_config_path: backend.as_ref().and_then(|binding| binding.path.to_str()),
            backend_config_sha256: backend.as_ref().map(|binding| binding.sha256.as_str()),
            project,
        },
    )?;
    require_applied_marker(&run, &record)?;
    require_project_binding(&project_root, project)?;
    let (environment, secrets) =
        lifecycle_environment(&document, &project_root, environment_source)?;
    let tofu_root = workspace.join("tofu");
    let result = require_success(runner.run(&ProcessRequest {
        argv: vec!["tofu".to_owned(), "output".to_owned(), "-json".to_owned()],
        cwd: tofu_root,
        environment,
        secrets,
    })?)?;
    require_project_binding(&project_root, project)?;
    let raw: Value = serde_json::from_str(&result.stdout).map_err(|error| {
        AinfraError::output_contract(format!("cannot parse OpenTofu output: {error}"))
    })?;
    let output = standardized_output(&record, &document, &raw)?;
    let output_path = run.join("output.json");
    crate::contracts::validate_document(&output, &output_path)?;
    validate_policy(&output, &output_path, Some(template.name), &project_root)?;
    persist_immutable_json(&output_path, &output)?;
    Ok(output)
}

/// Configure hosts for one exact apply run using validated output.
///
/// # Errors
///
/// Returns before Ansible execution when output, inventory, workspace, or
/// host-key evidence is missing or changed.
pub fn configure(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    run_id: &str,
    known_hosts: &Path,
    check: bool,
) -> Result<ProcessResult, AinfraError> {
    configure_bound(
        runner,
        environment_source,
        project_root,
        template_name,
        run_id,
        known_hosts,
        check,
        None,
        None,
    )
}

/// Configure one exact project-bound applied environment.
///
/// # Errors
///
/// Rejects changed project files or an environment mismatch before Ansible.
#[allow(clippy::too_many_arguments)]
pub fn configure_project(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    environment_name: &str,
    run_id: &str,
    known_hosts: &Path,
    check: bool,
) -> Result<ProcessResult, AinfraError> {
    let selection = project_selection(project_root, environment_name)?;
    configure_bound(
        runner,
        environment_source,
        &selection.binding.root,
        &selection.template,
        run_id,
        known_hosts,
        check,
        Some(&selection.binding),
        Some(environment_name),
    )
}

#[allow(clippy::too_many_arguments)]
fn configure_bound(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    run_id: &str,
    known_hosts: &Path,
    check: bool,
    project: Option<&ProjectBinding>,
    selected_environment: Option<&str>,
) -> Result<ProcessResult, AinfraError> {
    let project_root = project_root.canonicalize().map_err(|error| {
        AinfraError::guard(format!(
            "cannot resolve project root {}: {error}",
            project_root.display()
        ))
    })?;
    let known_hosts = verified_known_hosts(known_hosts)?;
    let reviewed_record = PlanRecord::load(&project_root, run_id)?;
    collect_output_bound(
        runner,
        environment_source,
        &project_root,
        template_name,
        run_id,
        project,
        selected_environment,
    )?;
    if PlanRecord::load(&project_root, run_id)? != reviewed_record {
        return Err(AinfraError::guard(
            "reviewed plan record changed during output collection",
        ));
    }
    require_project_binding(&project_root, project)?;
    let run = project_root.join(".ainfra/runs").join(run_id);
    let output = run.join("output.json");
    let inventory = run.join("inventory.yml");
    crate::inventory::write_inventory(&output, &inventory, &project_root)?;
    let ansible_root = run.join("workspace/ansible");
    let config = ansible_root
        .join("ansible.cfg")
        .canonicalize()
        .map_err(|_| AinfraError::guard("retained Ansible configuration is missing"))?;
    let ansible_home = run.join("ansible-home");
    let local_temp = ansible_home.join("tmp");
    fs::create_dir_all(&local_temp).map_err(|error| {
        AinfraError::dependency(format!("cannot create {}: {error}", local_temp.display()))
    })?;
    let mut environment = child_environment();
    environment.insert("ANSIBLE_CONFIG".to_owned(), config.display().to_string());
    environment.insert(
        "ANSIBLE_HOME".to_owned(),
        ansible_home.display().to_string(),
    );
    environment.insert(
        "ANSIBLE_LOCAL_TEMP".to_owned(),
        local_temp.display().to_string(),
    );
    environment.insert(
        "ANSIBLE_REMOTE_TEMP".to_owned(),
        format!("/tmp/ainfra-{run_id}"),
    );
    environment.insert("ANSIBLE_HOST_KEY_CHECKING".to_owned(), "True".to_owned());
    environment.insert("ANSIBLE_RETRY_FILES_ENABLED".to_owned(), "False".to_owned());
    environment.insert(
        "ANSIBLE_SSH_ARGS".to_owned(),
        format!(
            "-o UserKnownHostsFile={} -o StrictHostKeyChecking=yes",
            known_hosts.display()
        ),
    );
    if let Some(socket) = environment_source.get("SSH_AUTH_SOCK") {
        environment.insert("SSH_AUTH_SOCK".to_owned(), socket);
    }
    let mut argv = vec![
        "ansible-playbook".to_owned(),
        "--inventory".to_owned(),
        inventory.display().to_string(),
    ];
    if check {
        argv.extend(["--check".to_owned(), "--diff".to_owned()]);
    }
    argv.push("site.yml".to_owned());
    require_project_binding(&project_root, project)?;
    require_success(runner.run(&ProcessRequest {
        argv,
        cwd: ansible_root,
        environment,
        secrets: Vec::new(),
    })?)
}

#[allow(clippy::too_many_arguments)]
fn execute(
    runner: &dyn Runner,
    environment_source: &dyn EnvironmentSource,
    project_root: &Path,
    template_name: &str,
    input_path: &Path,
    approval: &str,
    operation: Operation,
    project: Option<&ProjectBinding>,
) -> Result<ProcessResult, AinfraError> {
    if approval.is_empty() {
        return Err(AinfraError::guard(format!(
            "{} requires an exact reviewed plan ID",
            operation.as_str()
        )));
    }
    let project_root = project_root.canonicalize().map_err(|error| {
        AinfraError::guard(format!(
            "cannot resolve project root {}: {error}",
            project_root.display()
        ))
    })?;
    let record = PlanRecord::load(&project_root, approval)?;
    let template = discover_builtin(template_name)?;
    let document = validate_path(input_path)?;
    validate_policy(&document, input_path, Some(template.name), &project_root)?;
    let environment_name = document["metadata"]["environment"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("missing environment"))?;
    let template_version = template.manifest["metadata"]["version"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("missing template version"))?;
    let run = project_root.join(".ainfra/runs").join(approval);
    let workspace = run.join("workspace");
    let backend = backend_binding(&document, &project_root, environment_source)?;
    record.verify(
        &project_root,
        &ExpectedBindings {
            template: template.name,
            template_version,
            environment: environment_name,
            input_path,
            template_root: &workspace,
            operation,
            backend_config_path: backend.as_ref().and_then(|binding| binding.path.to_str()),
            backend_config_sha256: backend.as_ref().map(|binding| binding.sha256.as_str()),
            project,
        },
    )?;
    let (environment, secrets) =
        lifecycle_environment(&document, &project_root, environment_source)?;
    let tofu_root = workspace.join(
        template.manifest["spec"]["engines"]["tofu"]["workingDirectory"]
            .as_str()
            .ok_or_else(|| AinfraError::input_contract("missing tofu working directory"))?,
    );
    require_project_binding(&project_root, project)?;
    initialize(runner, &tofu_root, &environment, &secrets, backend.as_ref())?;
    require_project_binding(&project_root, project)?;
    let result = require_success(runner.run(&ProcessRequest {
        argv: vec![
            "tofu".to_owned(),
            "apply".to_owned(),
            "-input=false".to_owned(),
            record.plan_path.clone(),
        ],
        cwd: tofu_root,
        environment,
        secrets,
    })?)?;
    persist_operation_marker(&run, &record)?;
    Ok(result)
}

struct ProjectSelection {
    binding: ProjectBinding,
    template: String,
    input: PathBuf,
}

fn project_selection(
    project_root: &Path,
    environment_name: &str,
) -> Result<ProjectSelection, AinfraError> {
    let before = ProjectBinding::capture(project_root)?;
    let project = Project::load(&before.root)?;
    let after = ProjectBinding::capture(&before.root)?;
    if before != after {
        return Err(AinfraError::guard(
            "project configuration changed during resolution",
        ));
    }
    let input = project
        .inputs
        .get(environment_name)
        .cloned()
        .ok_or_else(|| {
            AinfraError::input_contract(format!(
                "environment {environment_name:?} is not declared in ainfra.yaml"
            ))
        })?;
    Ok(ProjectSelection {
        binding: after,
        template: project.config.spec.template,
        input,
    })
}

fn require_project_binding(
    project_root: &Path,
    expected: Option<&ProjectBinding>,
) -> Result<(), AinfraError> {
    if let Some(expected) = expected
        && ProjectBinding::capture(project_root)? != *expected
    {
        return Err(AinfraError::guard(
            "project configuration changed before process execution",
        ));
    }
    Ok(())
}

/// Read-only environment boundary used for secret references.
pub trait EnvironmentSource {
    /// Return one environment value without exposing it through diagnostics.
    fn get(&self, name: &str) -> Option<String>;
}

/// Real process-environment source.
pub struct OsEnvironment;

struct BackendBinding {
    path: PathBuf,
    sha256: String,
}

impl EnvironmentSource for OsEnvironment {
    fn get(&self, name: &str) -> Option<String> {
        std::env::var(name).ok()
    }
}

fn initialize(
    runner: &dyn Runner,
    tofu_root: &Path,
    environment: &BTreeMap<String, String>,
    secrets: &[String],
    backend: Option<&BackendBinding>,
) -> Result<(), AinfraError> {
    let mut argv = vec![
        "tofu".to_owned(),
        "init".to_owned(),
        "-input=false".to_owned(),
        "-lockfile=readonly".to_owned(),
        "-reconfigure".to_owned(),
    ];
    if let Some(binding) = backend {
        argv.push(format!("-backend-config={}", binding.path.display()));
    }
    require_success(runner.run(&ProcessRequest {
        argv,
        cwd: tofu_root.to_path_buf(),
        environment: environment.clone(),
        secrets: secrets.to_vec(),
    })?)?;
    Ok(())
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

fn persist_immutable_json(path: &Path, document: &Value) -> Result<(), AinfraError> {
    let content = serde_json::to_string_pretty(document)
        .map_err(|error| AinfraError::dependency(error.to_string()))?;
    if path.exists() {
        let existing = fs::read_to_string(path).map_err(|error| {
            AinfraError::guard(format!("cannot read {}: {error}", path.display()))
        })?;
        if existing == content {
            return Ok(());
        }
        return Err(AinfraError::guard(
            "standardized output changed for an immutable apply run",
        ));
    }
    let temporary = path.with_extension("json.tmp");
    let mut file = OpenOptions::new()
        .create_new(true)
        .write(true)
        .open(&temporary)
        .map_err(|error| {
            AinfraError::guard(format!("cannot create {}: {error}", temporary.display()))
        })?;
    file.write_all(content.as_bytes()).map_err(|error| {
        AinfraError::guard(format!("cannot write {}: {error}", temporary.display()))
    })?;
    fs::rename(&temporary, path)
        .map_err(|error| AinfraError::guard(format!("cannot persist {}: {error}", path.display())))
}

fn persist_operation_marker(run: &Path, record: &PlanRecord) -> Result<(), AinfraError> {
    let marker = json!({
        "format_version": "ainfra.run/v1alpha1",
        "operation": record.operation.as_str(),
        "plan_id": record.id,
        "plan_sha256": record.plan_sha256,
    });
    persist_immutable_json(
        &run.join(format!("{}.json", record.operation.as_str())),
        &marker,
    )
}

fn require_applied_marker(run: &Path, record: &PlanRecord) -> Result<(), AinfraError> {
    let path = run.join("apply.json");
    let document: Value = serde_json::from_str(&fs::read_to_string(&path).map_err(|_| {
        AinfraError::guard("standardized output requires a successfully applied run")
    })?)
    .map_err(|_| AinfraError::guard("applied-run marker is invalid"))?;
    if document["format_version"] != "ainfra.run/v1alpha1"
        || document["operation"] != "apply"
        || document["plan_id"] != record.id
        || document["plan_sha256"] != record.plan_sha256
    {
        return Err(AinfraError::guard(
            "applied-run marker does not match the reviewed plan",
        ));
    }
    Ok(())
}

fn standardized_output(
    record: &PlanRecord,
    input: &Value,
    raw: &Value,
) -> Result<Value, AinfraError> {
    let object = raw
        .as_object()
        .ok_or_else(|| AinfraError::output_contract("OpenTofu output must be an object"))?;
    let expected = ["inventory_nodes", "ownership"];
    if object.len() != expected.len() || expected.iter().any(|name| !object.contains_key(*name)) {
        return Err(AinfraError::output_contract(
            "OpenTofu output must contain only inventory_nodes and ownership",
        ));
    }
    let value = |name: &str| -> Result<&Value, AinfraError> {
        let wrapper = &object[name];
        if wrapper["sensitive"] != false {
            return Err(AinfraError::output_contract(format!(
                "OpenTofu output {name} must be non-sensitive"
            )));
        }
        wrapper.get("value").ok_or_else(|| {
            AinfraError::output_contract(format!("OpenTofu output {name} has no value"))
        })
    };
    let ownership = value("ownership")?
        .as_object()
        .ok_or_else(|| AinfraError::output_contract("ownership output must be an object"))?;
    let environment = input["metadata"]["environment"]
        .as_str()
        .ok_or_else(|| AinfraError::input_contract("missing environment"))?;
    for (name, expected_value) in [
        ("managed-by", "ainfra"),
        ("template", record.template.as_str()),
        ("environment", environment),
    ] {
        if ownership.get(name).and_then(Value::as_str) != Some(expected_value) {
            return Err(AinfraError::output_contract(format!(
                "ownership output does not match {name}"
            )));
        }
    }
    let (host_aliases, nodes) = standardized_nodes(value("inventory_nodes")?, input)?;
    Ok(json!({
        "apiVersion": "ainfra.projectious.work/v1alpha1",
        "kind": "InfrastructureOutput",
        "metadata": {
            "template": record.template,
            "templateVersion": record.template_version,
            "environment": environment,
            "runId": record.id,
        },
        "spec": {
            "target": {
                "type": "kubernetes-ready",
                "name": format!("ainfra-{environment}"),
                "provider": "hetzner-cloud",
            },
            "access": {
                "ssh": {
                    "hostAliases": host_aliases,
                    "transport": "private-network",
                },
            },
            "network": {
                "privateCidrs": [
                    input["spec"]["network"]["privateCidr"].clone()
                ],
            },
            "nodes": nodes,
        },
    }))
}

fn standardized_nodes(
    value: &Value,
    input: &Value,
) -> Result<(Vec<Value>, Vec<Value>), AinfraError> {
    let raw_nodes = value
        .as_array()
        .ok_or_else(|| AinfraError::output_contract("inventory_nodes output must be an array"))?;
    let expected_count = input["spec"]["topology"]["controlPlaneCount"]
        .as_u64()
        .unwrap_or_default()
        + input["spec"]["topology"]["workerCount"]
            .as_u64()
            .unwrap_or_default();
    if u64::try_from(raw_nodes.len()).unwrap_or(u64::MAX) != expected_count {
        return Err(AinfraError::output_contract(
            "inventory node count does not match the reviewed input",
        ));
    }
    let mut names = std::collections::BTreeSet::new();
    let mut nodes = Vec::new();
    for raw_node in raw_nodes {
        let name = raw_node["name"]
            .as_str()
            .ok_or_else(|| AinfraError::output_contract("inventory node has no name"))?;
        if !names.insert(name.to_owned()) {
            return Err(AinfraError::output_contract(
                "inventory node names must be unique",
            ));
        }
        let mut node = serde_json::Map::from_iter([
            ("name".to_owned(), raw_node["name"].clone()),
            ("role".to_owned(), raw_node["role"].clone()),
            ("privateIPv4".to_owned(), raw_node["private_ipv4"].clone()),
            ("image".to_owned(), raw_node["image"].clone()),
        ]);
        for (source, destination) in [("public_ipv4", "publicIPv4"), ("public_ipv6", "publicIPv6")]
        {
            if let Some(value) = raw_node.get(source).filter(|value| !value.is_null()) {
                node.insert(destination.to_owned(), value.clone());
            }
        }
        nodes.push(Value::Object(node));
    }
    nodes.sort_by_key(|node| node["name"].as_str().unwrap_or_default().to_owned());
    Ok((names.into_iter().map(Value::String).collect(), nodes))
}

fn verified_known_hosts(path: &Path) -> Result<PathBuf, AinfraError> {
    if path
        .as_os_str()
        .to_string_lossy()
        .chars()
        .any(char::is_whitespace)
    {
        return Err(AinfraError::guard(
            "known_hosts path must not contain whitespace",
        ));
    }
    let resolved = path.canonicalize().map_err(|_| {
        AinfraError::guard(format!(
            "verified known_hosts file is missing: {}",
            path.display()
        ))
    })?;
    let metadata =
        fs::metadata(&resolved).map_err(|error| AinfraError::guard(error.to_string()))?;
    if !metadata.is_file() {
        return Err(AinfraError::guard(
            "verified known_hosts path is not a regular file",
        ));
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;

        if metadata.permissions().mode() & 0o022 != 0 {
            return Err(AinfraError::guard(
                "verified known_hosts file must not be group/world writable",
            ));
        }
    }
    Ok(resolved)
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

fn backend_binding(
    document: &Value,
    project_root: &Path,
    environment_source: &dyn EnvironmentSource,
) -> Result<Option<BackendBinding>, AinfraError> {
    let state = &document["spec"]["state"];
    if state["mode"] != "remote" {
        return Ok(None);
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
    let sha256 = file_hash(&resolved)?;
    Ok(Some(BackendBinding {
        path: resolved,
        sha256,
    }))
}
