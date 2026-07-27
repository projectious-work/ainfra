//! Thin binary entry point for ainfra.

use std::process::ExitCode;

fn main() -> ExitCode {
    match ainfra::run() {
        Ok(()) => ExitCode::SUCCESS,
        Err(error) => {
            eprintln!("{error}");
            ExitCode::from(error.exit_code())
        }
    }
}
