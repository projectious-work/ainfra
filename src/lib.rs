//! Library boundary for the ainfra command-line product.

pub mod cli;
pub mod contracts;
pub mod doctor;
pub mod error;
pub mod inventory;
pub mod legacy;
pub mod lifecycle;
pub mod plan_record;
pub mod policy;
pub mod process;
pub mod project;
pub mod run_record;
pub mod status;
pub mod template;

use clap::Parser;

use crate::cli::{Cli, Command, LegacyCommand};
use crate::error::AinfraError;

/// Parse and execute ainfra using the process argument vector.
///
/// The production command surface is implemented entirely in Rust. Frozen
/// compatibility fixtures preserve the former prototype's external contract.
///
/// # Errors
///
/// Returns a stable typed error when command validation or execution fails.
pub fn run() -> Result<(), AinfraError> {
    run_from(&Cli::parse())
}

/// Execute an already parsed command.
///
/// # Errors
///
/// Returns a stable typed error when the selected command fails.
pub fn run_from(cli: &Cli) -> Result<(), AinfraError> {
    match &cli.command {
        Command::Validate {
            target,
            input,
            format,
        } => cli::run_validate(target.as_deref(), input.as_deref(), *format),
        Command::Init {
            name,
            template,
            environment,
            format,
        } => cli::run_init(name.as_deref(), template, environment, *format),
        Command::Inventory {
            output,
            destination,
        } => cli::run_inventory(output, destination),
        Command::Status {
            environment,
            format,
        } => cli::run_status(environment, *format),
        Command::Legacy { command } => match command {
            LegacyCommand::Inspect { root, format } => cli::run_legacy_inspect(root, *format),
        },
        Command::Doctor {
            format,
            input,
            environment,
        } => cli::run_doctor(*format, input.as_deref(), environment.as_deref()),
        Command::Plan {
            template,
            input,
            environment,
            destroy,
            format,
        } => cli::run_plan(
            template.as_deref(),
            input.as_deref(),
            environment.as_deref(),
            *destroy,
            *format,
        ),
        Command::Apply {
            template,
            input,
            environment,
            approve,
        } => cli::run_execute(
            template.as_deref(),
            input.as_deref(),
            environment.as_deref(),
            approve,
            crate::plan_record::Operation::Apply,
        ),
        Command::Up {
            environment,
            approve,
            known_hosts,
            format,
        } => cli::run_up(environment, approve, known_hosts, *format),
        Command::Destroy {
            template,
            input,
            environment,
            approve_destroy,
        } => cli::run_execute(
            template.as_deref(),
            input.as_deref(),
            environment.as_deref(),
            approve_destroy,
            crate::plan_record::Operation::Destroy,
        ),
        Command::Down {
            environment,
            approve_destroy,
        } => cli::run_down(environment, approve_destroy),
        Command::Outputs {
            template,
            environment,
            run,
            format,
        } => cli::run_outputs(template.as_deref(), environment.as_deref(), run, *format),
        Command::Configure {
            template,
            environment,
            run,
            known_hosts,
            check,
        } => cli::run_configure(
            template.as_deref(),
            environment.as_deref(),
            run,
            known_hosts,
            *check,
        ),
    }
}
