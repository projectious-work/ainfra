#!/usr/bin/env python3
"""Prepare an immutable input snapshot for the Phase 1 host container gate."""

from __future__ import annotations

import datetime as dt
import hashlib
import json
import os
import re
import secrets
import shutil
import subprocess
import sys
import tarfile
from pathlib import Path

ACTIVE_RUN_DIR: Path | None = None
SEMVER = re.compile(
    r"^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"
    r"(?:-((?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)"
    r"(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?"
    r"(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$"
)


def cleanup_incomplete_run() -> None:
    global ACTIVE_RUN_DIR
    if ACTIVE_RUN_DIR is not None:
        shutil.rmtree(ACTIVE_RUN_DIR)
        ACTIVE_RUN_DIR = None


def fail(message: str) -> None:
    cleanup_incomplete_run()
    raise SystemExit(f"container-gate preparation failed: {message}")


def run(
    argv: list[str], *, cwd: Path, env: dict[str, str] | None = None
) -> bytes:
    try:
        completed = subprocess.run(
            argv,
            cwd=cwd,
            env=env,
            check=False,
            capture_output=True,
        )
    except FileNotFoundError:
        fail(f"required executable is unavailable: {argv[0]}")
    if completed.returncode != 0:
        detail = completed.stderr.decode("utf-8", "replace").strip()
        fail(f"{' '.join(argv)} failed: {detail}")
    return completed.stdout


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def load_release_checksums(release_dir: Path) -> dict[str, str]:
    manifest = release_dir / "checksums.sha256"
    if not manifest.is_file() or manifest.is_symlink():
        fail(f"missing or unsafe release checksum manifest: {manifest}")
    checksums: dict[str, str] = {}
    for line in manifest.read_text(encoding="ascii").splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  ([^/]+)", line)
        if match is None or match.group(2) in checksums:
            fail("release checksum manifest is malformed")
        checksums[match.group(2)] = match.group(1)
    return checksums


def validate_source_branch(branch: str) -> str:
    """Require provenance that the host gate can validate.

    ``git branch --show-current`` returns an empty string for detached HEAD.
    Such a checkout can identify an exact commit but cannot produce the
    non-empty branch provenance required by the host gate.
    """

    if not branch:
        fail("source checkout must be on a named branch")
    return branch


def copy_packaged_binary(
    release_dir: Path,
    checksums: dict[str, str],
    release_version: str,
    operating_system: str,
    architecture: str,
    output: Path,
) -> None:
    base = f"ainfra_{release_version}_{operating_system}_{architecture}"
    archive = release_dir / f"{base}.tar.gz"
    expected_digest = checksums.get(archive.name)
    if (
        expected_digest is None
        or not archive.is_file()
        or archive.is_symlink()
        or sha256(archive) != expected_digest
    ):
        fail(f"missing, unsafe, or checksum-invalid release archive: {archive}")
    member_name = f"{base}/ainfra"
    try:
        with tarfile.open(archive, "r:gz") as bundle:
            members = bundle.getmembers()
            if any(
                member.issym()
                or member.islnk()
                or member.name.startswith("/")
                or ".." in Path(member.name).parts
                for member in members
            ):
                fail(f"release archive contains an unsafe member: {archive}")
            member = bundle.getmember(member_name)
            if not member.isfile():
                fail(f"release archive lacks a regular binary: {archive}")
            source = bundle.extractfile(member)
            if source is None:
                fail(f"release archive binary cannot be read: {archive}")
            output.parent.mkdir(parents=True, mode=0o700, exist_ok=True)
            with source, output.open("wb") as destination:
                shutil.copyfileobj(source, destination)
    except (KeyError, tarfile.TarError, OSError) as error:
        fail(f"release archive could not be consumed: {archive}: {error}")


def main() -> int:
    global ACTIVE_RUN_DIR
    if len(sys.argv) != 2 or not sys.argv[1].startswith("--version="):
        fail("usage: prepare-container-gate.py --version=SEMVER")
    release_version = sys.argv[1].removeprefix("--version=").removeprefix("v")
    if not SEMVER.fullmatch(release_version):
        fail("release version is not strict SemVer")

    repo = Path(__file__).resolve().parent.parent
    discovered_repo = Path(
        run(["git", "rev-parse", "--show-toplevel"], cwd=repo)
        .decode()
        .strip()
    ).resolve()
    if discovered_repo != repo:
        fail("repository root could not be resolved from the script path")

    now = dt.datetime.now(dt.UTC).replace(microsecond=0)
    run_id = now.strftime("%Y%m%dT%H%M%SZ-") + secrets.token_hex(16)
    version_dir = repo / "tmp" / "container-gate" / release_version
    version_dir.mkdir(parents=True, mode=0o700, exist_ok=True)
    run_dir = version_dir / run_id
    input_dir = run_dir / "input"
    context_dir = input_dir / "context"
    run_dir.mkdir(parents=True, mode=0o700)
    ACTIVE_RUN_DIR = run_dir
    input_dir.mkdir(mode=0o700)
    context_dir.mkdir(parents=True, mode=0o700)

    commit = run(["git", "rev-parse", "HEAD"], cwd=repo).decode().strip()
    branch = validate_source_branch(
        run(["git", "branch", "--show-current"], cwd=repo).decode().strip()
    )
    status = run(["git", "status", "--porcelain=v1", "-z"], cwd=repo)
    diff = run(["git", "diff", "--binary", "HEAD"], cwd=repo)
    untracked = run(
        ["git", "ls-files", "--others", "--exclude-standard", "-z"],
        cwd=repo,
    )
    dirty_identity = hashlib.sha256(status + b"\0" + diff)
    for raw_path in sorted(path for path in untracked.split(b"\0") if path):
        relative = Path(os.fsdecode(raw_path))
        candidate = repo / relative
        if not candidate.is_file() or candidate.is_symlink():
            fail(f"untracked source is not a regular file: {relative}")
        dirty_identity.update(b"\0untracked\0" + raw_path + b"\0")
        dirty_identity.update(bytes.fromhex(sha256(candidate)))
    diff_digest = dirty_identity.hexdigest()

    provenance = {
        "schemaVersion": 1,
        "sourceCommit": commit,
        "branch": branch,
        "dirtyWorktree": bool(status),
        "diffSha256": diff_digest,
        "preparedAt": now.isoformat().replace("+00:00", "Z"),
        "releaseVersion": release_version,
    }
    (input_dir / "provenance.json").write_text(
        json.dumps(provenance, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    shutil.copyfile(repo / "Dockerfile", input_dir / "Dockerfile")

    release_dir = repo / "dist" / "release" / release_version
    checksums = load_release_checksums(release_dir)
    for operating_system in ("linux", "darwin"):
        for architecture in ("amd64", "arm64"):
            output = (
                context_dir
                / "dist"
                / operating_system
                / architecture
                / "ainfra"
            )
            copy_packaged_binary(
                release_dir,
                checksums,
                release_version,
                operating_system,
                architecture,
                output,
            )

    checked_files = [input_dir / "Dockerfile"] + sorted(
        path for path in context_dir.rglob("*") if path.is_file()
    )
    manifest_lines = [
        f"{sha256(path)}  {path.relative_to(input_dir).as_posix()}"
        for path in checked_files
    ]
    (input_dir / "checksums.sha256").write_text(
        "\n".join(manifest_lines) + "\n", encoding="ascii"
    )

    for path in sorted(input_dir.rglob("*"), reverse=True):
        if path.is_dir():
            mode = 0o555
        elif path.name == "ainfra" and "dist" in path.parts:
            mode = 0o555
        else:
            mode = 0o444
        path.chmod(mode)
    input_dir.chmod(0o555)

    ACTIVE_RUN_DIR = None
    print(run_dir)
    print(f"run id: {run_id}", file=sys.stderr)
    print(f"release version: {release_version}", file=sys.stderr)
    print(f"source commit: {commit}", file=sys.stderr)
    print(f"dirty worktree: {str(bool(status)).lower()}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except BaseException:
        cleanup_incomplete_run()
        raise
