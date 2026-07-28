//! Language-neutral CLI compatibility corpus executed by the Rust product.

#![allow(clippy::unwrap_used)]

use std::path::PathBuf;

use assert_cmd::Command;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct Case {
    name: String,
    args: Vec<String>,
    expect: Expectation,
}

#[derive(Debug, Deserialize)]
struct Expectation {
    exit: i32,
    stdout_contains: Vec<String>,
    stderr_contains: Vec<String>,
}

#[test]
fn rust_cli_satisfies_the_frozen_compatibility_corpus() {
    let root = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let cases: Vec<Case> = serde_yaml::from_str(include_str!("compat/cli/cases.yaml")).unwrap();
    for case in cases {
        let args: Vec<_> = case
            .args
            .iter()
            .map(|value| value.replace("<ROOT>", &root.display().to_string()))
            .collect();
        let output = Command::cargo_bin("ainfra")
            .unwrap()
            .current_dir(&root)
            .args(args)
            .output()
            .unwrap();
        assert_eq!(
            output.status.code(),
            Some(case.expect.exit),
            "{} used an unexpected exit code",
            case.name
        );
        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);
        for fragment in &case.expect.stdout_contains {
            assert!(
                stdout.contains(fragment),
                "{} stdout did not contain {fragment:?}: {stdout}",
                case.name
            );
        }
        for fragment in &case.expect.stderr_contains {
            assert!(
                stderr.contains(fragment),
                "{} stderr did not contain {fragment:?}: {stderr}",
                case.name
            );
        }
    }
}
