//! Read-only local readiness checks.

use std::collections::BTreeMap;
use std::path::{Path, PathBuf};
use std::process::Command;

use regex::Regex;
use serde::Serialize;

use crate::contracts::validate_path;
use crate::policy::validate_policy;
use crate::template::discover_builtin;

/// One stable, machine-readable readiness result.
#[derive(Clone, Debug, Serialize)]
pub struct Check {
    /// Stable check identifier.
    pub id: String,
    /// `pass` or `fail`.
    pub status: String,
    /// Detected value, when available.
    pub found: Option<String>,
    /// Required state.
    pub required: String,
    /// Operator remediation.
    pub remediation: String,
}

impl Check {
    fn new(
        id: &str,
        passed: bool,
        found: Option<String>,
        required: impl Into<String>,
        remediation: impl Into<String>,
    ) -> Self {
        Self {
            id: id.to_owned(),
            status: if passed { "pass" } else { "fail" }.to_owned(),
            found,
            required: required.into(),
            remediation: remediation.into(),
        }
    }
}

/// Run the compatibility doctor against an explicit project root.
#[must_use]
pub fn run_doctor(root: &Path, input: Option<&Path>) -> Vec<Check> {
    let mut checks = Vec::new();
    let pins = load_tool_pins(&root.join("tools.lock"));
    for (id, argv) in [
        ("tofu", &["tofu", "version"][..]),
        ("ansible-playbook", &["ansible-playbook", "--version"][..]),
        ("checkov", &["checkov", "--version"][..]),
        ("gitleaks", &["gitleaks", "version"][..]),
    ] {
        checks.push(command_check(id, argv, pins.get(id)));
    }
    for (name, id) in [("tools.lock", "tool-lock"), (".gitignore", "gitignore")] {
        checks.push(file_check(&root.join(name), id));
    }
    checks.push(template_check());
    checks.push(backend_policy_check(root, input));
    checks.push(Check::new(
        "github-workflows",
        !root.join(".github/workflows").exists(),
        None,
        "no GitHub workflows",
        "Remove .github/workflows; all automation is local.",
    ));
    checks
}

/// Run installed-product readiness checks for one validated project input.
#[must_use]
pub fn run_project_doctor(root: &Path, input: &Path) -> Vec<Check> {
    let tofu = ("1.10.0".to_owned(), "minimum".to_owned());
    let ansible = ("2.16.0".to_owned(), "minimum".to_owned());
    vec![
        command_check("tofu", &["tofu", "version"], Some(&tofu)),
        command_check(
            "ansible-playbook",
            &["ansible-playbook", "--version"],
            Some(&ansible),
        ),
        template_check(),
        backend_policy_check(root, Some(input)),
    ]
}

fn command_check(id: &str, argv: &[&str], pin: Option<&(String, String)>) -> Check {
    let (required_version, policy) = pin
        .cloned()
        .unwrap_or_else(|| ("unknown".to_owned(), "exact".to_owned()));
    let required = format!("{policy} {required_version}");
    let line = command_first_line(argv);
    let found = line.as_deref().and_then(extract_version).map(str::to_owned);
    let passed = found
        .as_deref()
        .is_some_and(|value| version_matches(value, &required_version, &policy));
    Check::new(
        id,
        passed,
        found.or(line),
        required.clone(),
        format!("Install {} at {required}.", argv[0]),
    )
}

fn command_first_line(argv: &[&str]) -> Option<String> {
    let output = Command::new(argv.first()?).args(&argv[1..]).output().ok()?;
    if !output.status.success() {
        return None;
    }
    let text = if output.stdout.is_empty() {
        &output.stderr
    } else {
        &output.stdout
    };
    String::from_utf8_lossy(text)
        .lines()
        .next()
        .map(str::to_owned)
}

fn extract_version(value: &str) -> Option<&str> {
    let pattern = Regex::new(r"\d+\.\d+(?:\.\d+)?").ok()?;
    pattern.find(value).map(|found| found.as_str())
}

fn version_matches(found: &str, required: &str, policy: &str) -> bool {
    let parts = |value: &str| {
        value
            .split('.')
            .map(str::parse::<u64>)
            .collect::<Result<Vec<_>, _>>()
            .ok()
    };
    match (parts(found), parts(required), policy) {
        (Some(found), Some(required), "exact") => found == required,
        (Some(found), Some(required), "minimum") => found >= required,
        _ => false,
    }
}

fn load_tool_pins(path: &Path) -> BTreeMap<String, (String, String)> {
    let Ok(document) = std::fs::read_to_string(path) else {
        return BTreeMap::new();
    };
    let mut result = BTreeMap::new();
    let mut current = None;
    let mut version = None;
    let mut policy = None;
    for line in document.lines().map(str::trim) {
        if let Some(name) = line
            .strip_prefix("[tools.")
            .and_then(|value| value.strip_suffix(']'))
        {
            if let (Some(name), Some(version), Some(policy)) =
                (current.take(), version.take(), policy.take())
            {
                result.insert(name, (version, policy));
            }
            current = Some(name.to_owned());
        } else if let Some(value) = assignment(line, "version") {
            version = Some(value);
        } else if let Some(value) = assignment(line, "policy") {
            policy = Some(value);
        }
    }
    if let (Some(name), Some(version), Some(policy)) = (current, version, policy) {
        result.insert(name, (version, policy));
    }
    result
}

fn assignment(line: &str, key: &str) -> Option<String> {
    line.strip_prefix(key)?
        .trim_start()
        .strip_prefix('=')?
        .trim()
        .strip_prefix('"')?
        .strip_suffix('"')
        .map(str::to_owned)
}

fn file_check(path: &Path, id: &str) -> Check {
    Check::new(
        id,
        path.is_file(),
        path.is_file().then(|| path.display().to_string()),
        format!(
            "{} present",
            path.file_name().unwrap_or_default().to_string_lossy()
        ),
        format!(
            "Restore the committed {} file.",
            path.file_name().unwrap_or_default().to_string_lossy()
        ),
    )
}

fn template_check() -> Check {
    match discover_builtin("hetzner-kubernetes-baseline") {
        Ok(template) => Check::new(
            "template-contract",
            true,
            Some(template.manifest_path.to_owned()),
            "reference template resolves safely",
            "Repair the reference manifest and its contained paths.",
        ),
        Err(error) => Check::new(
            "template-contract",
            false,
            Some(error.to_string()),
            "reference template resolves safely",
            "Repair the reference manifest and its contained paths.",
        ),
    }
}

fn backend_policy_check(root: &Path, input: Option<&Path>) -> Check {
    let default = root.join("templates/hetzner-kubernetes-baseline/inputs/example.input.yaml");
    let path: PathBuf = input.map_or(default, Path::to_path_buf);
    let result = validate_path(&path).and_then(|document| {
        validate_policy(&document, &path, Some("hetzner-kubernetes-baseline"), root)
    });
    match result {
        Ok(()) => Check::new(
            "backend-policy",
            true,
            Some(path.display().to_string()),
            "example input satisfies state policy",
            "Repair the example state mode and backend capabilities.",
        ),
        Err(error) => Check::new(
            "backend-policy",
            false,
            Some(error.to_string()),
            "example input satisfies state policy",
            "Repair the example state mode and backend capabilities.",
        ),
    }
}
