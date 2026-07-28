//! Library boundary for the ainfra command-line product.

pub mod cli;
pub mod contracts;
pub mod doctor;
pub mod error;
pub mod inventory;
pub mod plan_record;
pub mod policy;
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
        } => cli::run_validate(target, input.as_deref(), *format),
        Command::Inventory {
            output,
            destination,
        } => cli::run_inventory(output, destination),
        Command::Doctor { format, input } => cli::run_doctor(*format, input.as_deref()),
        command => Err(AinfraError::dependency(format!(
            "the Rust preview does not implement `{}` yet; use the Python CLI",
            command.name()
        ))),
    }
}
