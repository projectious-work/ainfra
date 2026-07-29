from __future__ import annotations

import hashlib
import json
import os
import platform
import subprocess
import tarfile
from pathlib import Path

ROOT = Path(__file__).parents[1]


def run(
    *args: str, env: dict[str, str] | None = None
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        args,
        cwd=ROOT,
        env=env,
        text=True,
        capture_output=True,
        check=False,
    )


def test_package_is_deterministic_and_has_canonical_contract(
    tmp_path: Path,
) -> None:
    binary = tmp_path / "ainfra"
    binary.write_bytes((ROOT / "target/debug/ainfra").read_bytes())
    binary.chmod(0o755)
    machine = platform.machine()
    target = {
        "aarch64": "aarch64-unknown-linux-gnu",
        "x86_64": "x86_64-unknown-linux-gnu",
    }[machine]

    command = (
        str(ROOT / "scripts/maintain.sh"),
        "package",
        "0.1.0",
        target,
        str(binary),
    )
    first = run(*command)
    assert first.returncode == 0, first.stderr
    archive = ROOT / f"dist/ainfra-v0.1.0-{target}.tar.gz"
    first_digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    second = run(*command)
    assert second.returncode == 0, second.stderr
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == first_digest
    assert archive.with_suffix(archive.suffix + ".sha256").read_text() == (
        f"{first_digest}\n"
    )
    with tarfile.open(archive) as bundle:
        assert bundle.getnames() == ["LICENSE", "ainfra"]
        binary_member = bundle.getmember("ainfra")
        assert binary_member.isfile()
        assert binary_member.mode == 0o755
        assert binary_member.uid == binary_member.gid == 0
    audit = run(
        str(ROOT / "scripts/maintain.sh"),
        "audit-release",
        "0.1.0",
        target,
    )
    assert audit.returncode == 0, audit.stderr
    assert "verified 1 release artifact(s)" in audit.stdout

    checksum = archive.with_suffix(archive.suffix + ".sha256")
    checksum.write_text(f"{'0' * 64}\n")
    rejected = run(
        str(ROOT / "scripts/maintain.sh"),
        "audit-release",
        "0.1.0",
        target,
    )
    assert rejected.returncode != 0
    assert "checksum verification failed" in rejected.stderr


def test_package_rejects_invalid_version_target_and_binary(
    tmp_path: Path,
) -> None:
    binary = tmp_path / "ainfra"
    binary.write_text("not executable")
    script = str(ROOT / "scripts/maintain.sh")
    for version, target in (
        ("v0.1.0", "x86_64-unknown-linux-gnu"),
        ("1..2", "x86_64-unknown-linux-gnu"),
        ("0.1.0", "armv7-unknown-linux-gnu"),
    ):
        result = run(script, "package", version, target, str(binary))
        assert result.returncode != 0


def test_installer_rejects_root_and_malformed_versions(tmp_path: Path) -> None:
    fake_bin = tmp_path / "bin"
    fake_bin.mkdir()
    fake_id = fake_bin / "id"
    fake_id.write_text("#!/bin/sh\nprintf '0\\n'\n")
    fake_id.chmod(0o755)
    env = os.environ | {
        "PATH": f"{fake_bin}:{os.environ['PATH']}",
        "HOME": str(tmp_path),
        "AINFRA_VERSION": "0.1.0;touch-pwned",
    }
    malformed = run(str(ROOT / "install.sh"), env=env)
    assert malformed.returncode != 0
    assert not (ROOT / "touch-pwned").exists()

    env["AINFRA_VERSION"] = "0.1.0"
    root = run(str(ROOT / "install.sh"), env=env)
    assert root.returncode != 0
    assert "refusing to install as root" in root.stderr


def test_installer_verifies_and_installs_release(tmp_path: Path) -> None:
    fake_bin = tmp_path / "fake-bin"
    assets = tmp_path / "assets"
    install_dir = tmp_path / "installed"
    fake_bin.mkdir()
    assets.mkdir()
    payload = tmp_path / "payload"
    payload.mkdir()
    executable = payload / "ainfra"
    executable.write_text(
        "#!/bin/sh\n"
        'case "$1" in --version) echo "ainfra 0.1.0";; '
        "--help) echo help;; esac\n"
    )
    executable.chmod(0o755)
    (payload / "LICENSE").write_text("MIT\n")
    name = "ainfra-v0.1.0-aarch64-unknown-linux-gnu.tar.gz"
    archive = assets / name
    subprocess.run(
        ["tar", "-czf", archive, "-C", payload, "LICENSE", "ainfra"],
        check=True,
    )
    (assets / f"{name}.sha256").write_text(
        f"{hashlib.sha256(archive.read_bytes()).hexdigest()}\n"
    )
    helpers = {
        "id": "#!/bin/sh\necho 1000\n",
        "uname": (
            '#!/bin/sh\ncase "$1" in -s) echo Linux;; -m) echo aarch64;; esac\n'
        ),
        "curl": (
            "#!/bin/sh\n"
            'while [ "$#" -gt 0 ]; do\n'
            '  if [ "$1" = --output ]; then output="$2"; shift 2; '
            'else url="$1"; shift; fi\n'
            "done\n"
            'cp "$TEST_ASSETS/${url##*/}" "$output"\n'
        ),
    }
    for name_, source in helpers.items():
        helper = fake_bin / name_
        helper.write_text(source)
        helper.chmod(0o755)
    env = os.environ | {
        "PATH": f"{fake_bin}:{os.environ['PATH']}",
        "HOME": str(tmp_path),
        "AINFRA_VERSION": "0.1.0",
        "AINFRA_INSTALL_DIR": str(install_dir),
        "TEST_ASSETS": str(assets),
    }
    result = run(str(ROOT / "install.sh"), env=env)
    assert result.returncode == 0, result.stderr
    assert (install_dir / "ainfra").stat().st_mode & 0o777 == 0o755
    assert "ainfra 0.1.0" in result.stdout

    (assets / f"{name}.sha256").write_text(f"{'0' * 64}\n")
    mismatch = run(str(ROOT / "install.sh"), env=env)
    assert mismatch.returncode != 0
    assert "checksum verification failed" in mismatch.stderr


def test_no_github_actions_release_workflow() -> None:
    workflows = ROOT / ".github/workflows"
    assert not workflows.exists() or not any(workflows.iterdir())


def test_macos_release_uses_system_tar_and_tag_bound_inputs() -> None:
    maintain = (ROOT / "scripts/maintain.sh").read_text()
    release_lib = (ROOT / "scripts/release-lib.sh").read_text()
    assert "command -v gtar" not in maintain
    assert "AINFRA_TAR=gtar" not in maintain
    assert "COPYFILE_DISABLE=1" in release_lib
    assert "--format ustar" in release_lib
    for release_input in (
        "Cargo.toml",
        "Cargo.lock",
        "LICENSE",
        "install.sh",
        "schemas",
        "src",
        "templates",
    ):
        assert release_input in maintain


def test_release_phase0_is_candidate_bound_and_documented() -> None:
    maintain = (ROOT / "scripts/maintain.sh").read_text()
    evidence = (ROOT / "scripts/release-evidence.sh").read_text()
    assert "release VERSION --steps phase0" in maintain
    assert "release VERSION --steps checks" in maintain
    assert 'release-evidence.sh" phase0' in maintain
    assert "set_version" in maintain
    assert "must be prepared and merged by PR" in maintain
    for report in (
        "RELEASE-STATE.md",
        "RELEASE-DOCTORS.md",
        "RELEASE-DOCS.md",
        "PHASE0.sha256",
    ):
        assert report in evidence
    for heading in (
        "Added",
        "Changed",
        "Fixed",
        "Known limitations",
        "Install",
        "Rollback",
    ):
        assert f"## {heading}" in (ROOT / "release-notes/v0.1.0.md").read_text()


def test_release_checks_are_bounded_bound_and_tamper_aware() -> None:
    checks = (ROOT / "scripts/release-checks.sh").read_text()
    for gate in (
        "rust-quality",
        "rust-tests",
        "repository",
        "infrastructure",
    ):
        assert gate in checks
    assert "AINFRA_RELEASE_JOBS:-2" in checks
    assert "candidate_commit=" in checks
    assert "cargo_lock_sha256=" in checks
    assert "rustc_fingerprint=" in checks
    assert "evidence_reusable" in checks
    assert "CHECKS.sha256" in checks


def test_live_hetzner_smoke_is_opt_in_and_cleanup_bound() -> None:
    smoke = (ROOT / "scripts/live-hetzner-release-smoke").read_text()
    assert "cost-and-destroy-approved" in smoke
    assert "trap cleanup EXIT INT TERM" in smoke
    assert "workerCount: 0" in smoke
    assert "serverType: cx23" in smoke
    assert "prepare_destroy_plan" in smoke
    assert 'state list)" ]]' in smoke
    assert "--selector" in smoke
    assert "cleanup_complete=true" in smoke


def test_release_metadata_binds_archive_and_locked_source() -> None:
    target = "aarch64-unknown-linux-gnu"
    archive = ROOT / f"dist/ainfra-v0.1.0-{target}.tar.gz"
    archive.parent.mkdir(exist_ok=True)
    archive.write_bytes(b"release artifact fixture")
    source_commit = "0" * 40
    env = os.environ | {
        "SOURCE_DATE_EPOCH": "0",
        "AINFRA_RELEASE_SOURCE_COMMIT": source_commit,
    }
    generated = run(
        str(ROOT / "scripts/generate-release-metadata.py"),
        "--root",
        str(ROOT),
        "--version",
        "0.1.0",
        "--target",
        target,
        env=env,
    )
    assert generated.returncode == 0, generated.stderr
    sbom = json.loads((ROOT / "dist/ainfra-v0.1.0.spdx.json").read_text())
    assert sbom["spdxVersion"] == "SPDX-2.3"
    assert sbom["creationInfo"]["created"] == "1970-01-01T00:00:00Z"
    provenance = json.loads(
        (ROOT / f"dist/ainfra-v0.1.0-{target}.provenance.json").read_text()
    )
    assert provenance["predicateType"] == "https://slsa.dev/provenance/v1"
    dependencies = provenance["predicate"]["buildDefinition"][
        "resolvedDependencies"
    ]
    assert dependencies[0]["digest"]["gitCommit"] == source_commit
    subject = provenance["subject"][0]
    assert subject["name"] == archive.name
    assert (
        subject["digest"]["sha256"]
        == hashlib.sha256(archive.read_bytes()).hexdigest()
    )
