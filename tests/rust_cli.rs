//! Installed-binary-style tests for the Rust product shell.

#![allow(clippy::unwrap_used)]

use assert_cmd::Command;
use predicates::prelude::*;

const COMMANDS: [&str; 7] = [
    "validate",
    "doctor",
    "plan",
    "apply",
    "destroy",
    "outputs",
    "inventory",
];

#[test]
fn help_exposes_the_compatibility_surface() {
    let mut command = Command::cargo_bin("ainfra").unwrap();
    let assertion = command.arg("--help").assert().success();
    let stdout = String::from_utf8_lossy(&assertion.get_output().stdout);
    for name in COMMANDS {
        assert!(stdout.contains(name), "help did not contain {name}");
    }
}

#[test]
fn version_uses_the_package_version() {
    Command::cargo_bin("ainfra")
        .unwrap()
        .arg("--version")
        .assert()
        .success()
        .stdout(predicate::str::contains(env!("CARGO_PKG_VERSION")));
}

#[test]
fn help_works_outside_the_source_checkout() {
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(std::env::temp_dir())
        .arg("--help")
        .assert()
        .success();
}

#[test]
fn compatibility_command_shapes_refuse_safely_in_the_preview() {
    let cases: &[&[&str]] = &[
        &["validate", "fixture.json", "--format", "json"],
        &["doctor", "--format", "json", "--input", "fixture.json"],
        &[
            "plan",
            "fixture-template",
            "--input",
            "fixture.json",
            "--destroy",
            "--format",
            "json",
        ],
        &[
            "apply",
            "fixture-template",
            "--input",
            "fixture.json",
            "--approve",
            "fixture-plan",
        ],
        &[
            "destroy",
            "fixture-template",
            "--input",
            "fixture.json",
            "--approve-destroy",
            "fixture-plan",
        ],
        &["outputs", "fixture-template", "--format", "yaml"],
        &[
            "inventory",
            "--output",
            "fixture-output.json",
            "--destination",
            "inventory.yaml",
        ],
    ];

    for arguments in cases {
        Command::cargo_bin("ainfra")
            .unwrap()
            .args(*arguments)
            .assert()
            .code(5)
            .stderr(predicate::str::contains("AINFRA-E500"));
    }
}
