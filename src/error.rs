//! Stable public error envelope for the Rust implementation.

use thiserror::Error;

/// An expected ainfra failure with a stable identifier and exit code.
#[derive(Debug, Error)]
#[error("{code}: {message}")]
pub struct AinfraError {
    code: &'static str,
    message: String,
    exit_code: u8,
}

impl AinfraError {
    /// Construct a dependency or underlying-tool failure.
    #[must_use]
    pub fn dependency(message: impl Into<String>) -> Self {
        Self {
            code: "AINFRA-E500",
            message: message.into(),
            exit_code: 5,
        }
    }

    /// Return the stable public error identifier.
    #[must_use]
    pub const fn code(&self) -> &'static str {
        self.code
    }

    /// Return the process exit code.
    #[must_use]
    pub const fn exit_code(&self) -> u8 {
        self.exit_code
    }
}
