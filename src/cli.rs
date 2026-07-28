//! Command-line syntax shared by the Rust product shell.

use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueEnum};
use serde_json::json;

use crate::contracts::validate_path;
use crate::doctor;
use crate::error::AinfraError;
use crate::inventory::write_inventory;
use crate::lifecycle::{self, OsEnvironment};
use crate::plan_record::Operation;
use crate::policy::validate_policy;
use crate::process::SubprocessRunner;
use crate::project::{Project, initialize};
use crate::template::discover_builtin;

/// Validate and operate explicit infrastructure templates.
#[derive(Debug, Parser)]
#[command(name = "ainfra", version, about)]
pub struct Cli {
    /// Lifecycle command to execute.
    #[command(subcommand)]
    pub command: Command,
}

/// Initial compatibility command surface.
#[derive(Debug, Subcommand)]
pub enum Command {
    /// Initialize a project without overwriting configuration.
    Init {
        /// Project name; defaults to the current directory name.
        #[arg(long)]
        name: Option<String>,
        /// Built-in template to lock.
        #[arg(long, default_value = "hetzner-kubernetes-baseline")]
        template: String,
        /// Example environment name.
        #[arg(long, default_value = "development")]
        environment: String,
        /// Output format.
        #[arg(long, value_enum, default_value_t)]
        format: TextFormat,
    },
    /// Validate a manifest, input, or standardized output.
    Validate {
        /// Document path or template name; omit to validate the project.
        target: Option<String>,
        /// Input document to validate with the selected template.
        #[arg(long)]
        input: Option<PathBuf>,
        /// Output format.
        #[arg(long, value_enum, default_value_t)]
        format: TextFormat,
    },
    /// Report local dependency readiness without mutation.
    Doctor {
        /// Output format.
        #[arg(long, value_enum, default_value_t)]
        format: TextFormat,
        /// Validate backend readiness for this environment input.
        #[arg(long)]
        input: Option<PathBuf>,
        /// Project environment to check.
        #[arg(long)]
        environment: Option<String>,
    },
    /// Create a reviewable plan.
    Plan {
        /// Template name for explicit compatibility mode.
        template: Option<String>,
        /// Environment input for explicit compatibility mode.
        #[arg(long)]
        input: Option<PathBuf>,
        /// Configured project environment.
        #[arg(long)]
        environment: Option<String>,
        /// Create a reviewed destroy plan without applying it.
        #[arg(long)]
        destroy: bool,
        /// Output format.
        #[arg(long, value_enum, default_value_t)]
        format: TextFormat,
    },
    /// Apply an exact reviewed plan.
    Apply {
        /// Template name for explicit compatibility mode.
        template: Option<String>,
        /// Environment input for explicit compatibility mode.
        #[arg(long)]
        input: Option<PathBuf>,
        /// Configured project environment.
        #[arg(long)]
        environment: Option<String>,
        /// Exact reviewed plan identifier.
        #[arg(long, value_name = "PLAN_ID")]
        approve: String,
    },
    /// Destroy an exact reviewed plan.
    Destroy {
        /// Template name for explicit compatibility mode.
        template: Option<String>,
        /// Environment input for explicit compatibility mode.
        #[arg(long)]
        input: Option<PathBuf>,
        /// Configured project environment.
        #[arg(long)]
        environment: Option<String>,
        /// Exact reviewed destroy-plan identifier.
        #[arg(long, value_name = "PLAN_ID")]
        approve_destroy: String,
    },
    /// Read the sanitized standardized output.
    Outputs {
        /// Template name for explicit compatibility mode.
        template: Option<String>,
        /// Configured project environment.
        #[arg(long)]
        environment: Option<String>,
        /// Exact applied run identifier.
        #[arg(long, value_name = "PLAN_ID")]
        run: String,
        /// Output serialization.
        #[arg(long, value_enum, default_value_t)]
        format: DocumentFormat,
    },
    /// Configure hosts from one exact applied run.
    Configure {
        /// Template name for explicit compatibility mode.
        template: Option<String>,
        /// Configured project environment.
        #[arg(long)]
        environment: Option<String>,
        /// Exact applied run identifier.
        #[arg(long, value_name = "PLAN_ID")]
        run: String,
        /// Independently verified SSH known-hosts file.
        #[arg(long)]
        known_hosts: PathBuf,
        /// Run Ansible in check mode with a diff.
        #[arg(long)]
        check: bool,
    },
    /// Generate Ansible inventory from standardized output.
    Inventory {
        /// Standardized infrastructure output.
        #[arg(long)]
        output: PathBuf,
        /// Inventory destination.
        #[arg(long)]
        destination: PathBuf,
    },
}

impl Command {
    /// Return the stable command spelling.
    #[must_use]
    pub const fn name(&self) -> &'static str {
        match self {
            Self::Init { .. } => "init",
            Self::Validate { .. } => "validate",
            Self::Doctor { .. } => "doctor",
            Self::Plan { .. } => "plan",
            Self::Apply { .. } => "apply",
            Self::Destroy { .. } => "destroy",
            Self::Outputs { .. } => "outputs",
            Self::Configure { .. } => "configure",
            Self::Inventory { .. } => "inventory",
        }
    }
}

/// Text or JSON command output.
#[derive(Clone, Copy, Debug, Default, ValueEnum)]
pub enum TextFormat {
    /// Human-readable output.
    #[default]
    Text,
    /// Machine-readable JSON.
    Json,
}

/// Validate one contract document through the Rust implementation.
///
/// # Errors
///
/// Returns a stable contract or policy error when validation fails.
pub fn run_validate(
    target: Option<&str>,
    input: Option<&std::path::Path>,
    format: TextFormat,
) -> Result<(), AinfraError> {
    let Some(target) = target else {
        if input.is_some() {
            return Err(AinfraError::input_contract(
                "--input requires a template target",
            ));
        }
        let root =
            std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
        let project = Project::discover(&root)?;
        match format {
            TextFormat::Text => {
                println!("valid AinfraProject: {}", project.root.display());
            }
            TextFormat::Json => println!(
                "{}",
                serde_json::to_string(&json!({
                    "apiVersion": project.config.api_version,
                    "environments": project.inputs.keys().collect::<Vec<_>>(),
                    "kind": project.config.kind,
                    "ok": true,
                    "projectRoot": project.root,
                    "template": {
                        "name": project.lock.template.name,
                        "sha256": project.lock.template.sha256,
                        "source": project.lock.template.source,
                        "version": project.lock.template.version,
                    },
                }))
                .map_err(|error| AinfraError::dependency(error.to_string()))?
            ),
        }
        return Ok(());
    };
    let path = std::path::Path::new(target);
    if !path.exists() {
        let template = discover_builtin(target)?;
        let mut validated = vec![template.manifest_path.to_owned()];
        if let Some(input) = input {
            let document = validate_path(input)?;
            let root = std::env::current_dir()
                .map_err(|error| AinfraError::dependency(error.to_string()))?;
            validate_policy(&document, input, Some(template.name), &root)?;
            validated.push(input.display().to_string());
        }
        match format {
            TextFormat::Text => {
                println!("valid InfrastructureTemplate: {target}");
            }
            TextFormat::Json => println!(
                "{}",
                serde_json::to_string(&json!({
                    "apiVersion": template.manifest["apiVersion"],
                    "kind": template.manifest["kind"],
                    "ok": true,
                    "template": template.name,
                    "validated": validated,
                }))
                .map_err(|error| AinfraError::dependency(error.to_string()))?
            ),
        }
        return Ok(());
    }
    let document = validate_path(path)?;
    let root =
        std::env::current_dir().map_err(|error| AinfraError::dependency(error.to_string()))?;
    validate_policy(&document, path, None, &root)?;
    let kind = document
        .get("kind")
        .and_then(serde_json::Value::as_str)
        .unwrap_or("<unknown>");
    match format {
        TextFormat::Text => println!("valid {kind}: {target}"),
        TextFormat::Json => println!(
            "{}",
            serde_json::to_string(&json!({
                "apiVersion": document["apiVersion"],
                "kind": kind,
                "ok": true,
                "path": target,
            }))
            .map_err(|error| AinfraError::dependency(error.to_string()))?
        ),
    }
    Ok(())
}

/// Initialize the current directory as an ainfra project.
///
/// # Errors
///
/// Refuses invalid names, unsupported templates, and all primary-file
/// overwrites.
pub fn run_init(
    name: Option<&str>,
    template: &str,
    environment: &str,
    format: TextFormat,
) -> Result<(), AinfraError> {
    let root = std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
    let result = initialize(&root, name, template, environment)?;
    match format {
        TextFormat::Json => println!("{}", crate::project::init_json(&result)?),
        TextFormat::Text => {
            println!("initialized ainfra project: {}", result.project_root);
            println!(
                "template: {} {}",
                result.template.name, result.template.version
            );
            for path in &result.created {
                println!("created: {path}");
            }
            for path in &result.updated {
                println!("updated: {path}");
            }
            for next in result.next {
                println!("next: {next}");
            }
        }
    }
    Ok(())
}

/// Generate a validated Ansible inventory.
///
/// # Errors
///
/// Returns a stable contract, policy, guard, or filesystem error.
pub fn run_inventory(
    output: &std::path::Path,
    destination: &std::path::Path,
) -> Result<(), AinfraError> {
    let root =
        std::env::current_dir().map_err(|error| AinfraError::dependency(error.to_string()))?;
    write_inventory(output, destination, &root)?;
    println!("inventory written: {}", destination.display());
    Ok(())
}

/// Report local readiness without changing the host.
///
/// # Errors
///
/// Returns a dependency error when any check fails.
pub fn run_doctor(
    format: TextFormat,
    input: Option<&std::path::Path>,
    environment: Option<&str>,
) -> Result<(), AinfraError> {
    let current =
        std::env::current_dir().map_err(|error| AinfraError::dependency(error.to_string()))?;
    let checks = match (input, environment) {
        (Some(_), Some(_)) => {
            return Err(AinfraError::input_contract(
                "choose either --input or --environment",
            ));
        }
        (input, None) => doctor::run_doctor(&current, input),
        (None, Some(environment)) => {
            let project = Project::discover(&current)?;
            let input = project.inputs.get(environment).ok_or_else(|| {
                AinfraError::input_contract(format!(
                    "environment {environment:?} is not declared in ainfra.yaml"
                ))
            })?;
            doctor::run_project_doctor(&project.root, input)
        }
    };
    match format {
        TextFormat::Json => println!(
            "{}",
            serde_json::to_string(&checks)
                .map_err(|error| AinfraError::dependency(error.to_string()))?
        ),
        TextFormat::Text => {
            for check in &checks {
                println!("{}: {}", check.status, check.id);
            }
        }
    }
    if checks.iter().any(|check| check.status == "fail") {
        return Err(AinfraError::dependency(
            "one or more readiness checks failed",
        ));
    }
    Ok(())
}

/// Create one isolated reviewed `OpenTofu` plan.
///
/// # Errors
///
/// Returns a stable validation, dependency, or guard error.
pub fn run_plan(
    template: Option<&str>,
    input: Option<&std::path::Path>,
    environment: Option<&str>,
    destroy: bool,
    format: TextFormat,
) -> Result<(), AinfraError> {
    let current = std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
    let operation = if destroy {
        Operation::Destroy
    } else {
        Operation::Apply
    };
    let runner = SubprocessRunner::default();
    let record = match lifecycle_mode(template, input, environment)? {
        LifecycleMode::Explicit { template, input } => lifecycle::plan(
            &runner,
            &OsEnvironment,
            &current,
            template,
            input,
            operation,
        )?,
        LifecycleMode::Project { environment } => {
            let project = Project::discover(&current)?;
            lifecycle::plan_project(
                &runner,
                &OsEnvironment,
                &project.root,
                environment,
                operation,
            )?
        }
    };
    match format {
        TextFormat::Json => println!(
            "{}",
            serde_json::to_string(&record)
                .map_err(|error| AinfraError::dependency(error.to_string()))?
        ),
        TextFormat::Text => {
            println!("plan {}", record.id);
            println!("{} approval: {}", operation.as_str(), record.id);
        }
    }
    Ok(())
}

/// Execute one exact reviewed apply or destroy plan.
///
/// # Errors
///
/// Returns a stable guard error before mutation when a binding changed.
pub fn run_execute(
    template: Option<&str>,
    input: Option<&std::path::Path>,
    environment: Option<&str>,
    approval: &str,
    operation: Operation,
) -> Result<(), AinfraError> {
    let current = std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
    let runner = SubprocessRunner::default();
    let result = match (lifecycle_mode(template, input, environment)?, operation) {
        (LifecycleMode::Explicit { template, input }, Operation::Apply) => {
            lifecycle::apply(&runner, &OsEnvironment, &current, template, input, approval)
        }
        (LifecycleMode::Explicit { template, input }, Operation::Destroy) => {
            lifecycle::destroy(&runner, &OsEnvironment, &current, template, input, approval)
        }
        (LifecycleMode::Project { environment }, Operation::Apply) => {
            let project = Project::discover(&current)?;
            lifecycle::apply_project(
                &runner,
                &OsEnvironment,
                &project.root,
                environment,
                approval,
            )
        }
        (LifecycleMode::Project { environment }, Operation::Destroy) => {
            let project = Project::discover(&current)?;
            lifecycle::destroy_project(
                &runner,
                &OsEnvironment,
                &project.root,
                environment,
                approval,
            )
        }
    }?;
    print!("{}", result.stdout);
    Ok(())
}

/// Emit validated standardized output for one exact apply run.
///
/// # Errors
///
/// Returns a stable output-contract, dependency, or guard error.
pub fn run_outputs(
    template: Option<&str>,
    environment: Option<&str>,
    run: &str,
    format: DocumentFormat,
) -> Result<(), AinfraError> {
    let current = std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
    let runner = SubprocessRunner::default();
    let output = match template_mode(template, environment)? {
        TemplateMode::Explicit(template) => {
            lifecycle::collect_output(&runner, &OsEnvironment, &current, template, run)?
        }
        TemplateMode::Project(environment) => {
            let project = Project::discover(&current)?;
            lifecycle::collect_output_project(
                &runner,
                &OsEnvironment,
                &project.root,
                environment,
                run,
            )?
        }
    };
    match format {
        DocumentFormat::Json => println!(
            "{}",
            serde_json::to_string_pretty(&output)
                .map_err(|error| AinfraError::dependency(error.to_string()))?
        ),
        DocumentFormat::Yaml => print!(
            "{}",
            serde_yaml::to_string(&output)
                .map_err(|error| AinfraError::dependency(error.to_string()))?
        ),
    }
    Ok(())
}

/// Configure hosts from validated output and verified host keys.
///
/// # Errors
///
/// Returns before Ansible when any run binding or host-key control fails.
pub fn run_configure(
    template: Option<&str>,
    environment: Option<&str>,
    run: &str,
    known_hosts: &std::path::Path,
    check: bool,
) -> Result<(), AinfraError> {
    let current = std::env::current_dir().map_err(|error| AinfraError::guard(error.to_string()))?;
    let runner = SubprocessRunner::default();
    let result = match template_mode(template, environment)? {
        TemplateMode::Explicit(template) => lifecycle::configure(
            &runner,
            &OsEnvironment,
            &current,
            template,
            run,
            known_hosts,
            check,
        )?,
        TemplateMode::Project(environment) => {
            let project = Project::discover(&current)?;
            lifecycle::configure_project(
                &runner,
                &OsEnvironment,
                &project.root,
                environment,
                run,
                known_hosts,
                check,
            )?
        }
    };
    print!("{}", result.stdout);
    Ok(())
}

enum LifecycleMode<'a> {
    Explicit {
        template: &'a str,
        input: &'a std::path::Path,
    },
    Project {
        environment: &'a str,
    },
}

fn lifecycle_mode<'a>(
    template: Option<&'a str>,
    input: Option<&'a std::path::Path>,
    environment: Option<&'a str>,
) -> Result<LifecycleMode<'a>, AinfraError> {
    match (template, input, environment) {
        (Some(template), Some(input), None) => Ok(LifecycleMode::Explicit { template, input }),
        (None, None, Some(environment)) => Ok(LifecycleMode::Project { environment }),
        _ => Err(AinfraError::input_contract(
            "use either TEMPLATE --input INPUT or --environment ENV",
        )),
    }
}

enum TemplateMode<'a> {
    Explicit(&'a str),
    Project(&'a str),
}

fn template_mode<'a>(
    template: Option<&'a str>,
    environment: Option<&'a str>,
) -> Result<TemplateMode<'a>, AinfraError> {
    match (template, environment) {
        (Some(template), None) => Ok(TemplateMode::Explicit(template)),
        (None, Some(environment)) => Ok(TemplateMode::Project(environment)),
        _ => Err(AinfraError::input_contract(
            "use either TEMPLATE or --environment ENV",
        )),
    }
}

/// JSON or YAML document output.
#[derive(Clone, Copy, Debug, Default, ValueEnum)]
pub enum DocumentFormat {
    /// JSON document.
    #[default]
    Json,
    /// YAML document.
    Yaml,
}
