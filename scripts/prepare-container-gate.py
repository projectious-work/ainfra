#!/usr/bin/env python3
"""Prepare an immutable input snapshot for the Phase 1 host container gate."""

from __future__ import annotations

import datetime as dt
import hashlib
import json
import os
import secrets
import shutil
import subprocess
import sys
from pathlib import Path

ACTIVE_RUN_DIR: Path | None = None


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


def main() -> int:
    global ACTIVE_RUN_DIR
    if len(sys.argv) != 1:
        fail("this command accepts no arguments")

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
    run_dir = repo / "tmp" / "container-gate" / run_id
    input_dir = run_dir / "input"
    context_dir = input_dir / "context"
    platform_dir = context_dir / "dist" / "linux"
    run_dir.mkdir(parents=True, mode=0o700)
    ACTIVE_RUN_DIR = run_dir
    input_dir.mkdir(mode=0o700)
    platform_dir.mkdir(parents=True, mode=0o700)

    commit = run(["git", "rev-parse", "HEAD"], cwd=repo).decode().strip()
    branch = run(["git", "branch", "--show-current"], cwd=repo).decode().strip()
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
    }
    (input_dir / "provenance.json").write_text(
        json.dumps(provenance, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    shutil.copyfile(repo / "Dockerfile", input_dir / "Dockerfile")

    base_env = os.environ.copy()
    for architecture in ("amd64", "arm64"):
        output = platform_dir / architecture / "ainfra"
        output.parent.mkdir(mode=0o700)
        build_env = base_env | {
            "CGO_ENABLED": "0",
            "GOOS": "linux",
            "GOARCH": architecture,
        }
        run(
            [
                "go",
                "build",
                "-trimpath",
                "-o",
                str(output),
                "./cmd/ainfra",
            ],
            cwd=repo,
            env=build_env,
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
        path.chmod(0o555 if path.is_dir() else 0o444)
    input_dir.chmod(0o555)

    ACTIVE_RUN_DIR = None
    print(run_dir)
    print(f"run id: {run_id}", file=sys.stderr)
    print(f"source commit: {commit}", file=sys.stderr)
    print(f"dirty worktree: {str(bool(status)).lower()}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except BaseException:
        cleanup_incomplete_run()
        raise
