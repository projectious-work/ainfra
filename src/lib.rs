//! Library boundary for the ainfra command-line product.

pub mod cli;
pub mod error;

use clap::Parser;

use crate::cli::Cli;
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
    Err(AinfraError::dependency(format!(
        "the Rust preview does not implement `{}` yet; use the Python CLI",
        cli.command.name()
    )))
}
