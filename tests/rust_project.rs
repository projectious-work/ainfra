//! Project initialization, discovery, and lock-boundary coverage.

#![allow(clippy::unwrap_used)]

use std::fs;
use std::path::Path;

use ainfra::project::{Project, initialize};

fn initialize_fixture(root: &Path) {
    initialize(
        root,
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
}

#[test]
fn init_creates_a_valid_checkout_independent_project() {
    let directory = tempfile::tempdir().unwrap();
    let result = initialize(
        directory.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();

    assert_eq!(result.api_version, "ainfra.init/v1alpha1");
    assert_eq!(result.template.source, "builtin");
    assert_eq!(result.template.version, "0.1.0");
    assert_eq!(result.template.sha256.len(), 64);
    assert!(directory.path().join("ainfra.yaml").is_file());
    assert!(directory.path().join("ainfra.lock").is_file());
    assert!(
        directory
            .path()
            .join("environments/development.yaml")
            .is_file()
    );
    assert!(
        fs::read_to_string(directory.path().join(".gitignore"))
            .unwrap()
            .lines()
            .any(|line| line == ".ainfra/")
    );
    assert!(!directory.path().join(".ainfra").exists());

    let project = Project::load(directory.path()).unwrap();
    assert_eq!(project.config.metadata.name, "example-infrastructure");
    assert!(project.inputs.contains_key("development"));
}

#[test]
fn init_preflights_conflicts_without_partial_writes() {
    let directory = tempfile::tempdir().unwrap();
    fs::write(directory.path().join("ainfra.lock"), "owned by user\n").unwrap();

    let error = initialize(
        directory.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("ainfra.lock"));
    assert!(!directory.path().join("ainfra.yaml").exists());
    assert!(!directory.path().join("environments").exists());
    assert_eq!(
        fs::read_to_string(directory.path().join("ainfra.lock")).unwrap(),
        "owned by user\n"
    );
}

#[test]
fn init_appends_but_never_duplicates_the_ignore_rule() {
    let directory = tempfile::tempdir().unwrap();
    fs::write(directory.path().join(".gitignore"), "owned/\n").unwrap();
    initialize_fixture(directory.path());
    let content = fs::read_to_string(directory.path().join(".gitignore")).unwrap();
    assert!(content.starts_with("owned/\n"));
    assert_eq!(
        content.lines().filter(|line| *line == ".ainfra/").count(),
        1
    );
}

#[test]
fn discovery_chooses_the_nearest_ancestor_project() {
    let outer = tempfile::tempdir().unwrap();
    initialize_fixture(outer.path());
    let inner = outer.path().join("nested/project");
    fs::create_dir_all(&inner).unwrap();
    initialize(
        &inner,
        Some("inner-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap();
    let child = inner.join("work/deep");
    fs::create_dir_all(&child).unwrap();

    let project = Project::discover(&child).unwrap();

    assert_eq!(project.root, inner.canonicalize().unwrap());
    assert_eq!(project.config.metadata.name, "inner-infrastructure");
}

#[test]
fn discovery_without_a_marker_has_a_stable_guard_error() {
    let directory = tempfile::tempdir().unwrap();
    let error = Project::discover(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("no ainfra.yaml found"));
}

#[test]
fn stale_lock_and_escaped_inputs_are_rejected() {
    let directory = tempfile::tempdir().unwrap();
    initialize_fixture(directory.path());
    let lock_path = directory.path().join("ainfra.lock");
    let mut lock: serde_yaml::Value =
        serde_yaml::from_str(&fs::read_to_string(&lock_path).unwrap()).unwrap();
    lock["template"]["sha256"] = serde_yaml::Value::String("0".repeat(64));
    fs::write(&lock_path, serde_yaml::to_string(&lock).unwrap()).unwrap();
    let error = Project::load(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("ainfra.lock does not match"));
}

#[test]
fn parent_traversal_input_is_rejected() {
    let directory = tempfile::tempdir().unwrap();
    initialize_fixture(directory.path());
    let config_path = directory.path().join("ainfra.yaml");
    let mut config: serde_yaml::Value =
        serde_yaml::from_str(&fs::read_to_string(&config_path).unwrap()).unwrap();
    config["spec"]["environments"]["development"]["input"] =
        serde_yaml::Value::String("../outside.yaml".to_owned());
    fs::write(&config_path, serde_yaml::to_string(&config).unwrap()).unwrap();

    let error = Project::load(directory.path()).unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("contained relative path"));
}

#[cfg(unix)]
#[test]
fn symlinked_project_files_and_inputs_are_rejected() {
    use std::os::unix::fs::symlink;

    let directory = tempfile::tempdir().unwrap();
    initialize_fixture(directory.path());
    let config = directory.path().join("ainfra.yaml");
    let real_config = directory.path().join("real-project.yaml");
    fs::rename(&config, &real_config).unwrap();
    symlink(&real_config, &config).unwrap();
    let error = Project::load(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("regular non-symlink"));

    fs::remove_file(&config).unwrap();
    fs::rename(&real_config, &config).unwrap();
    let input = directory.path().join("environments/development.yaml");
    let real_input = directory.path().join("environments/real.yaml");
    fs::rename(&input, &real_input).unwrap();
    symlink(&real_input, &input).unwrap();
    let error = Project::load(directory.path()).unwrap_err();
    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("regular non-symlink"));
}

#[cfg(unix)]
#[test]
fn init_rejects_redirected_write_paths_before_creating_files() {
    use std::os::unix::fs::symlink;

    let project = tempfile::tempdir().unwrap();
    let outside = tempfile::tempdir().unwrap();
    symlink(outside.path(), project.path().join("environments")).unwrap();

    let error = initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains("environments"));
    assert!(!project.path().join("ainfra.yaml").exists());
    assert!(!project.path().join("ainfra.lock").exists());
    assert!(!outside.path().join("development.yaml").exists());

    fs::remove_file(project.path().join("environments")).unwrap();
    let outside_ignore = outside.path().join("external-ignore");
    fs::write(&outside_ignore, "owned/\n").unwrap();
    symlink(&outside_ignore, project.path().join(".gitignore")).unwrap();
    let error = initialize(
        project.path(),
        Some("example-infrastructure"),
        "hetzner-kubernetes-baseline",
        "development",
    )
    .unwrap_err();

    assert_eq!(error.code(), "AINFRA-E600");
    assert!(error.to_string().contains(".gitignore"));
    assert!(!project.path().join("ainfra.yaml").exists());
    assert!(!project.path().join("ainfra.lock").exists());
    assert_eq!(fs::read_to_string(outside_ignore).unwrap(), "owned/\n");
}
