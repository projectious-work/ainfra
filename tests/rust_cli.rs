//! Installed-binary-style tests for the Rust product shell.

#![allow(clippy::unwrap_used)]

use assert_cmd::Command;
use predicates::prelude::*;

const COMMANDS: [&str; 8] = [
    "validate",
    "doctor",
    "plan",
    "apply",
    "destroy",
    "outputs",
    "configure",
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
fn doctor_emits_machine_readable_checks() {
    Command::cargo_bin("ainfra")
        .unwrap()
        .args(["doctor", "--format", "json"])
        .assert()
        .failure()
        .stdout(predicate::str::contains("\"id\":\"python\""))
        .stdout(predicate::str::contains("\"id\":\"github-workflows\""))
        .stderr(predicate::str::contains("AINFRA-E500"));
}

#[cfg(unix)]
#[test]
#[allow(clippy::too_many_lines)]
fn plan_works_outside_the_checkout_with_an_embedded_template() {
    use std::fs;
    use std::os::unix::fs::PermissionsExt;
    use std::path::Path;

    let project = tempfile::tempdir().unwrap();
    let bin = project.path().join("bin");
    fs::create_dir(&bin).unwrap();
    let tofu = bin.join("tofu");
    let raw_output = serde_json::json!({
        "inventory_nodes": {
            "sensitive": false,
            "value": [{
                "name": "ainfra-development-control-01",
                "role": "control-plane-capable",
                "private_ipv4": "10.42.0.10",
                "public_ipv4": null,
                "public_ipv6": null,
                "image": "debian-13",
            }],
        },
        "ownership": {
            "sensitive": false,
            "value": {
                "managed-by": "ainfra",
                "template": "hetzner-kubernetes-baseline",
                "environment": "development",
            },
        },
    })
    .to_string();
    fs::write(
        &tofu,
        format!(
            "#!/bin/sh\nfor value in \"$@\"; do\n\
         case \"$value\" in -out=*) : > \"${{value#-out=}}\";; esac\n\
         done\nif [ \"$1\" = apply ]; then printf applied; fi\n\
         if [ \"$1\" = output ]; then printf '%s' '{raw_output}'; fi\n",
        ),
    )
    .unwrap();
    fs::set_permissions(&tofu, fs::Permissions::from_mode(0o700)).unwrap();
    let ansible = bin.join("ansible-playbook");
    fs::write(&ansible, "#!/bin/sh\nprintf configured\n").unwrap();
    fs::set_permissions(&ansible, fs::Permissions::from_mode(0o700)).unwrap();
    let path = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let input = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/contracts/v1alpha1/valid/template-input.json");
    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "plan",
            "hetzner-kubernetes-baseline",
            "--input",
            input.to_str().unwrap(),
            "--format",
            "json",
        ])
        .output()
        .unwrap();
    assert!(output.status.success());
    let record: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert_eq!(record["format_version"], "ainfra.plan/v1alpha1");
    let plan_id = record["id"].as_str().unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "apply",
            "hetzner-kubernetes-baseline",
            "--input",
            input.to_str().unwrap(),
            "--approve",
            plan_id,
        ])
        .assert()
        .success()
        .stdout("applied");
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "outputs",
            "hetzner-kubernetes-baseline",
            "--run",
            plan_id,
            "--format",
            "json",
        ])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"InfrastructureOutput\""));
    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o600)).unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "configure",
            "hetzner-kubernetes-baseline",
            "--run",
            plan_id,
            "--known-hosts",
            known_hosts.to_str().unwrap(),
        ])
        .assert()
        .success()
        .stdout("configured");

    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "plan",
            "hetzner-kubernetes-baseline",
            "--input",
            input.to_str().unwrap(),
            "--destroy",
            "--format",
            "json",
        ])
        .output()
        .unwrap();
    assert!(output.status.success());
    let record: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    let destroy_id = record["id"].as_str().unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "destroy",
            "hetzner-kubernetes-baseline",
            "--input",
            input.to_str().unwrap(),
            "--approve-destroy",
            destroy_id,
        ])
        .assert()
        .success()
        .stdout("applied");
    let runs = project.path().join(".ainfra/runs");
    assert!(
        fs::read_dir(runs)
            .unwrap()
            .filter_map(Result::ok)
            .any(|entry| entry.path().join("plan.json").is_file())
    );
}
