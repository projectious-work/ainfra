//! Command-line syntax shared by the Rust product shell.

use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueEnum};

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
    /// Validate a manifest, input, or standardized output.
    Validate {
        /// Document path or template name beneath the template catalog.
        target: String,
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
    },
    /// Create a reviewable plan.
    Plan {
        /// Template name.
        template: String,
        /// Environment input document.
        #[arg(long)]
        input: PathBuf,
        /// Create a reviewed destroy plan without applying it.
        #[arg(long)]
        destroy: bool,
        /// Output format.
        #[arg(long, value_enum, default_value_t)]
        format: TextFormat,
    },
    /// Apply an exact reviewed plan.
    Apply {
        /// Template name.
        template: String,
        /// Environment input document.
        #[arg(long)]
        input: PathBuf,
        /// Exact reviewed plan identifier.
        #[arg(long, value_name = "PLAN_ID")]
        approve: String,
    },
    /// Destroy an exact reviewed plan.
    Destroy {
        /// Template name.
        template: String,
        /// Environment input document.
        #[arg(long)]
        input: PathBuf,
        /// Exact reviewed destroy-plan identifier.
        #[arg(long, value_name = "PLAN_ID")]
        approve_destroy: String,
    },
    /// Read the sanitized standardized output.
    Outputs {
        /// Template name.
        template: String,
        /// Output serialization.
        #[arg(long, value_enum, default_value_t)]
        format: DocumentFormat,
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
            Self::Validate { .. } => "validate",
            Self::Doctor { .. } => "doctor",
            Self::Plan { .. } => "plan",
            Self::Apply { .. } => "apply",
            Self::Destroy { .. } => "destroy",
            Self::Outputs { .. } => "outputs",
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

/// JSON or YAML document output.
#[derive(Clone, Copy, Debug, Default, ValueEnum)]
pub enum DocumentFormat {
    /// JSON document.
    #[default]
    Json,
    /// YAML document.
    Yaml,
}
