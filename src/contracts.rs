//! Embedded `v1alpha1` contract validation.

use std::fs;
use std::net::{Ipv4Addr, Ipv6Addr};
use std::path::Path;

use jsonschema::Draft;
use serde_json::Value;

use crate::error::AinfraError;

const MANIFEST_SCHEMA: &str = include_str!("../schemas/template-manifest.v1alpha1.json");
const INPUT_SCHEMA: &str = include_str!("../schemas/template-input.v1alpha1.json");
const OUTPUT_SCHEMA: &str = include_str!("../schemas/template-output.v1alpha1.json");

/// Load and validate one JSON or YAML contract document.
///
/// # Errors
///
/// Returns a stable input/output contract error for read, parse, dispatch, or
/// schema validation failures.
pub fn validate_path(path: &Path) -> Result<Value, AinfraError> {
    let content = fs::read_to_string(path).map_err(|error| {
        AinfraError::input_contract(format!("cannot read {}: {error}", path.display()))
    })?;
    let document = if path.extension().is_some_and(|suffix| suffix == "json") {
        serde_json::from_str(&content).map_err(|error| {
            AinfraError::input_contract(format!("cannot parse {}: {error}", path.display()))
        })?
    } else {
        serde_yaml::from_str(&content).map_err(|error| {
            AinfraError::input_contract(format!("cannot parse {}: {error}", path.display()))
        })?
    };
    validate_document(&document, path)?;
    Ok(document)
}

/// Validate an already parsed contract document.
///
/// # Errors
///
/// Returns a stable input/output contract error when the kind or schema is
/// invalid.
pub fn validate_document(document: &Value, source: &Path) -> Result<(), AinfraError> {
    let object = document.as_object().ok_or_else(|| {
        AinfraError::input_contract(format!("{}: document must be an object", source.display()))
    })?;
    let kind = object.get("kind").and_then(Value::as_str).ok_or_else(|| {
        AinfraError::input_contract(format!(
            "{}: /kind is required and must be a string",
            source.display()
        ))
    })?;
    let schema_text = match kind {
        "InfrastructureTemplate" => MANIFEST_SCHEMA,
        "TemplateInput" => INPUT_SCHEMA,
        "InfrastructureOutput" => OUTPUT_SCHEMA,
        other => {
            return Err(AinfraError::input_contract(format!(
                "unsupported kind {other:?}; expected InfrastructureOutput, \
                 InfrastructureTemplate, TemplateInput"
            )));
        }
    };
    let schema: Value = serde_json::from_str(schema_text)
        .map_err(|error| AinfraError::dependency(format!("embedded schema is invalid: {error}")))?;
    let validator = jsonschema::options()
        .with_draft(Draft::Draft202012)
        .should_validate_formats(true)
        .with_format("ipv4-network", is_ipv4_network)
        .with_format("ipv6-network", is_ipv6_network)
        .build(&schema)
        .map_err(|error| AinfraError::dependency(format!("embedded schema is invalid: {error}")))?;
    let mut errors: Vec<_> = validator.iter_errors(document).collect();
    errors.sort_by_key(|error| error.instance_path().to_string());
    if let Some(error) = errors.first() {
        let raw_pointer = error.instance_path().to_string();
        let pointer = if raw_pointer.is_empty() {
            "<root>"
        } else {
            raw_pointer.as_str()
        };
        let message = format!("{}: {pointer}: {error}", source.display());
        return Err(if kind == "InfrastructureOutput" {
            AinfraError::output_contract(message)
        } else {
            AinfraError::input_contract(message)
        });
    }
    Ok(())
}

fn is_ipv4_network(value: &str) -> bool {
    strict_network(value, true)
}

fn is_ipv6_network(value: &str) -> bool {
    strict_network(value, false)
}

fn strict_network(value: &str, ipv4: bool) -> bool {
    let Some((address, prefix)) = value.split_once('/') else {
        return false;
    };
    let Ok(prefix) = prefix.parse::<u8>() else {
        return false;
    };
    if ipv4 {
        let Ok(address) = address.parse::<Ipv4Addr>() else {
            return false;
        };
        if prefix > 32 {
            return false;
        }
        let numeric = u32::from(address);
        let host_bits = 32_u8.saturating_sub(prefix);
        let mask = if host_bits == 32 {
            0
        } else {
            u32::MAX << host_bits
        };
        numeric & mask == numeric
    } else {
        let Ok(address) = address.parse::<Ipv6Addr>() else {
            return false;
        };
        if prefix > 128 {
            return false;
        }
        let numeric = u128::from(address);
        let host_bits = 128_u8.saturating_sub(prefix);
        let mask = if host_bits == 128 {
            0
        } else {
            u128::MAX << host_bits
        };
        numeric & mask == numeric
    }
}
