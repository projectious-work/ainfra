//! Library boundary for the ainfra command-line product.

pub mod cli;
pub mod contracts;
pub mod doctor;
pub mod error;
pub mod inventory;
pub mod lifecycle;
pub mod plan_record;
pub mod policy;
pub mod process;
pub mod project;
pub mod template;

use clap::Parser;

use crate::cli::{Cli, Command};
use crate::error::AinfraError;

/// Parse and execute ainfra using the process argument vector.
///
/// The first Rust milestone intentionally exposes only the product shell.
/// Python remains the behavioral oracle until compatibility fixtures cover
/// each command.
///
/// # Errors
///
/// Returns a stable preview error when a parsed command reaches the
/// not-yet-ported execution boundary.
pub fn run() -> Result<(), AinfraError> {
    run_from(&Cli::parse())
}

/// Execute an already parsed command.
///
/// # Errors
///
/// Returns a stable preview error until the selected command has passed its
/// compatibility gate and is implemented in Rust.
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
        Command::Doctor { format, input } => cli::run_doctor(*format, input.as_deref()),
        Command::Plan {
            template,
            input,
            destroy,
            format,
        } => cli::run_plan(template, input, *destroy, *format),
        Command::Apply {
            template,
            input,
            approve,
        } => cli::run_execute(
            template,
            input,
            approve,
            crate::plan_record::Operation::Apply,
        ),
        Command::Destroy {
            template,
            input,
            approve_destroy,
        } => cli::run_execute(
            template,
            input,
            approve_destroy,
            crate::plan_record::Operation::Destroy,
        ),
        Command::Outputs {
            template,
            run,
            format,
        } => cli::run_outputs(template, run, *format),
        Command::Configure {
            template,
            run,
            known_hosts,
            check,
        } => cli::run_configure(template, run, known_hosts, *check),
    }
}
