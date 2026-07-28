//! Installed-binary-style tests for the Rust product shell.

#![allow(clippy::unwrap_used)]

use assert_cmd::Command;
use predicates::prelude::*;

const COMMANDS: [&str; 13] = [
    "init",
    "validate",
    "doctor",
    "plan",
    "apply",
    "up",
    "destroy",
    "down",
    "outputs",
    "configure",
    "status",
    "legacy",
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
fn init_and_project_validation_work_outside_the_checkout() {
    let project = tempfile::tempdir().unwrap();
    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .args([
            "init",
            "--name",
            "example-infrastructure",
            "--format",
            "json",
        ])
        .output()
        .unwrap();
    assert!(output.status.success());
    let result: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert_eq!(result["apiVersion"], "ainfra.init/v1alpha1");
    assert_eq!(result["template"]["source"], "builtin");
    assert!(!project.path().join(".ainfra").exists());

    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path().join("environments"))
        .args(["validate", "--format", "json"])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"kind\":\"AinfraProject\""))
        .stdout(predicate::str::contains("\"development\""));
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path().join("environments"))
        .args(["status", "--environment", "development", "--format", "json"])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"state\":\"none\""))
        .stdout(predicate::str::contains(
            "ainfra plan --environment development",
        ));
}

#[test]
fn init_refuses_overwrite_without_partial_changes() {
    let project = tempfile::tempdir().unwrap();
    std::fs::write(project.path().join("ainfra.lock"), "user-owned\n").unwrap();

    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .args(["init", "--name", "example-infrastructure"])
        .assert()
        .failure()
        .stderr(predicate::str::contains("AINFRA-E600"))
        .stderr(predicate::str::contains("ainfra.lock"));

    assert!(!project.path().join("ainfra.yaml").exists());
    assert!(!project.path().join("environments").exists());
}

#[cfg(unix)]
#[test]
#[allow(clippy::too_many_lines)]
fn project_lifecycle_works_from_a_nested_directory() {
    use std::fs;
    use std::os::unix::fs::PermissionsExt;

    let project = tempfile::tempdir().unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .args(["init", "--name", "example-infrastructure"])
        .assert()
        .success();
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
             done\nif [ \"$1\" = version ]; then printf 'OpenTofu v1.10.0'; fi\n\
             if [ \"$1\" = apply ]; then printf applied; fi\n\
             if [ \"$1\" = output ]; then printf '%s' '{raw_output}'; fi\n",
        ),
    )
    .unwrap();
    fs::set_permissions(&tofu, fs::Permissions::from_mode(0o700)).unwrap();
    let ansible = bin.join("ansible-playbook");
    fs::write(
        &ansible,
        "#!/bin/sh\nif [ \"$1\" = --version ]; then \
         printf 'ansible-playbook core 2.16.0'; \
         elif printf '%s\\n' \"$@\" | grep -q '^--check$'; then \
         printf 'PLAY RECAP ****\\nainfra-development-control-01 : \
         ok=3 changed=0 unreachable=0 failed=0 \
         skipped=0 rescued=0 ignored=0\\n'; else printf configured; fi\n",
    )
    .unwrap();
    fs::set_permissions(&ansible, fs::Permissions::from_mode(0o700)).unwrap();
    let path = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let nested = project.path().join("work/deep");
    fs::create_dir_all(&nested).unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .args(["doctor", "--environment", "development", "--format", "json"])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"id\":\"tofu\""));
    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args(["plan", "--environment", "development", "--format", "json"])
        .output()
        .unwrap();
    assert!(output.status.success());
    let record: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert_eq!(record["format_version"], "ainfra.plan/v1alpha2");
    let plan_id = record["id"].as_str().unwrap();

    let known_hosts = project.path().join("known_hosts");
    fs::write(&known_hosts, "host ssh-ed25519 fixture\n").unwrap();
    fs::set_permissions(&known_hosts, fs::Permissions::from_mode(0o600)).unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "up",
            "--environment",
            "development",
            "--approve",
            plan_id,
            "--known-hosts",
            known_hosts.to_str().unwrap(),
            "--format",
            "json",
        ])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"InfrastructureOutput\""))
        .stdout(predicate::str::contains("fixture-secret").not());
    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args(["status", "--environment", "development", "--format", "json"])
        .output()
        .unwrap();
    assert!(output.status.success());
    let text = String::from_utf8(output.stdout).unwrap();
    assert!(text.contains("\"state\":\"verified\""));
    assert!(!text.contains("fixture-secret"));
    assert!(!text.contains("HCLOUD_TOKEN"));

    let output = Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "plan",
            "--environment",
            "development",
            "--destroy",
            "--format",
            "json",
        ])
        .output()
        .unwrap();
    assert!(output.status.success());
    let destroy_record: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    let destroy_id = destroy_record["id"].as_str().unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .env("HCLOUD_TOKEN", "fixture-secret")
        .args([
            "down",
            "--environment",
            "development",
            "--approve-destroy",
            destroy_id,
        ])
        .assert()
        .success()
        .stdout(predicate::str::contains("destroyed and verified"));
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(&nested)
        .env("PATH", &path)
        .args(["status", "--environment", "development", "--format", "json"])
        .assert()
        .success()
        .stdout(predicate::str::contains("\"state\":\"destroyed\""));
}

#[test]
fn lifecycle_selectors_are_mutually_exclusive() {
    let project = tempfile::tempdir().unwrap();
    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(project.path())
        .args([
            "plan",
            "hetzner-kubernetes-baseline",
            "--input",
            "input.yaml",
            "--environment",
            "development",
        ])
        .assert()
        .failure()
        .stderr(predicate::str::contains("AINFRA-E200"))
        .stderr(predicate::str::contains(
            "either TEMPLATE --input INPUT or --environment ENV",
        ));
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

#[test]
fn legacy_inspection_is_read_only_and_checkout_independent() {
    let source = tempfile::tempdir().unwrap();

    Command::cargo_bin("ainfra")
        .unwrap()
        .current_dir(std::env::temp_dir())
        .args([
            "legacy",
            "inspect",
            "--root",
            source.path().to_str().unwrap(),
            "--format",
            "json",
        ])
        .assert()
        .success()
        .stdout(predicate::str::contains(
            "\"apiVersion\":\"ainfra.legacy-inspection/v1alpha1\"",
        ))
        .stdout(predicate::str::contains("\"state\":\"none\""));

    assert!(!source.path().join(".ainfra").exists());
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
