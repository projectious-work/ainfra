#!/usr/bin/env python3
"""Enforce an ordered, commit-bound release integrity state machine."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any

SEMVER = re.compile(
    r"^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"
    r"(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$"
)
STAGES = ("frozen", "packaged", "gated", "signed", "published")
LANES = ("v1.x-dev", "v1.x-pre-release", "v1.x-release")


class IntegrityError(RuntimeError):
    """A release integrity invariant was not satisfied."""


def run(repo: Path, *argv: str) -> str:
    completed = subprocess.run(
        list(argv), cwd=repo, check=False, capture_output=True, text=True
    )
    if completed.returncode != 0:
        detail = completed.stderr.strip() or completed.stdout.strip()
        raise IntegrityError(f"{' '.join(argv)} failed: {detail}")
    return completed.stdout.strip()


def digest(path: Path) -> str:
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def state_path(repo: Path, version: str) -> Path:
    return repo / "tmp" / "release" / version / "release-state.json"


def load_state(repo: Path, version: str) -> dict[str, Any]:
    path = state_path(repo, version)
    if not path.is_file() or path.is_symlink():
        raise IntegrityError(f"release is not frozen: {path}")
    try:
        state = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise IntegrityError(f"release state is invalid: {error}") from error
    if state.get("schemaVersion") != 1 or state.get("version") != version:
        raise IntegrityError("release state identity is invalid")
    return state


def write_state(repo: Path, version: str, state: dict[str, Any]) -> None:
    path = state_path(repo, version)
    path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(
        json.dumps(state, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    os.replace(temporary, path)


def git_identity(repo: Path) -> tuple[str, str, dict[str, str]]:
    head = run(repo, "git", "rev-parse", "HEAD")
    tree = run(repo, "git", "rev-parse", "HEAD^{tree}")
    lanes = {
        lane: run(repo, "git", "rev-parse", f"origin/{lane}") for lane in LANES
    }
    return head, tree, lanes


def require_clean(repo: Path) -> None:
    if run(repo, "git", "status", "--porcelain"):
        raise IntegrityError("release worktree is not clean")


def validate_frozen(repo: Path, state: dict[str, Any]) -> None:
    require_clean(repo)
    head, tree, lanes = git_identity(repo)
    if head != state.get("commit") or tree != state.get("tree"):
        raise IntegrityError(
            "HEAD no longer matches the frozen release identity"
        )
    if lanes != state.get("lanes") or any(
        value != head for value in lanes.values()
    ):
        raise IntegrityError("release lanes no longer match the frozen commit")


def verify_checksums(release_dir: Path) -> str:
    manifest = release_dir / "checksums.sha256"
    if not manifest.is_file() or manifest.is_symlink():
        raise IntegrityError("release checksum manifest is missing or unsafe")
    seen: set[str] = set()
    for line in manifest.read_text(encoding="ascii").splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  ([^/]+)", line)
        if match is None or match.group(2) in seen:
            raise IntegrityError("release checksum manifest is malformed")
        seen.add(match.group(2))
        candidate = release_dir / match.group(2)
        if not candidate.is_file() or candidate.is_symlink():
            raise IntegrityError(
                f"release artifact is missing or unsafe: {candidate}"
            )
        if digest(candidate) != match.group(1):
            raise IntegrityError(
                f"release artifact checksum mismatch: {candidate}"
            )
    if not seen:
        raise IntegrityError("release checksum manifest is empty")
    return digest(manifest)


def find_gate(repo: Path, version: str, commit: str) -> tuple[str, str]:
    version_dir = repo / "tmp" / "container-gate" / version
    candidates = sorted(
        version_dir.glob("*/evidence/result.json"), reverse=True
    )
    for result_path in candidates:
        try:
            result = json.loads(result_path.read_text(encoding="utf-8"))
            provenance = result["provenance"]
            grype = json.loads((result_path.parent / "grype.json").read_text())
        except (OSError, KeyError, json.JSONDecodeError, TypeError):
            continue
        if (
            result.get("ok") is True
            and result.get("cleanupOk") is True
            and provenance.get("sourceCommit") == commit
            and provenance.get("dirtyWorktree") is False
            and not grype.get("matches")
        ):
            return result.get("runId", result_path.parent.parent.name), digest(
                result_path
            )
    raise IntegrityError("no successful host gate matches the frozen commit")


def freeze(repo: Path, version: str) -> None:
    require_clean(repo)
    head, tree, lanes = git_identity(repo)
    if any(value != head for value in lanes.values()):
        detail = ", ".join(
            f"{name}={value[:12]}" for name, value in lanes.items()
        )
        raise IntegrityError(
            f"release lanes are not aligned with HEAD: {detail}"
        )
    notes = repo / "docs" / "releases" / f"v{version}.md"
    changelog = repo / "CHANGELOG.md"
    if not notes.is_file() or notes.is_symlink():
        raise IntegrityError(f"release notes are missing or unsafe: {notes}")
    if f"## [{version}]" not in changelog.read_text(encoding="utf-8"):
        raise IntegrityError("changelog lacks the release version")
    path = state_path(repo, version)
    if path.exists():
        raise IntegrityError(f"release is already frozen: {path}")
    now = (
        dt.datetime.now(dt.UTC)
        .replace(microsecond=0)
        .isoformat()
        .replace("+00:00", "Z")
    )
    write_state(
        repo,
        version,
        {
            "schemaVersion": 1,
            "version": version,
            "commit": head,
            "tree": tree,
            "lanes": lanes,
            "frozenAt": now,
            "stages": {"frozen": {"recordedAt": now}},
        },
    )
    print(f"release frozen: v{version} at {head}")


def record(repo: Path, version: str, stage: str) -> None:
    state = load_state(repo, version)
    validate_frozen(repo, state)
    index = STAGES.index(stage)
    if stage in state["stages"]:
        require(repo, version, stage)
        print(f"release stage already recorded: v{version} {stage}")
        return
    later = [name for name in STAGES[index + 1 :] if name in state["stages"]]
    if later:
        raise IntegrityError(
            f"release stage {stage} cannot follow completed stage {later[0]}"
        )
    for prerequisite in STAGES[:index]:
        if prerequisite not in state["stages"]:
            raise IntegrityError(
                f"release stage {stage} requires {prerequisite}"
            )
    evidence: dict[str, Any] = {
        "recordedAt": dt.datetime.now(dt.UTC)
        .replace(microsecond=0)
        .isoformat()
        .replace("+00:00", "Z")
    }
    release_dir = repo / "dist" / "release" / version
    if stage in {"packaged", "gated", "signed", "published"}:
        evidence["checksumsSha256"] = verify_checksums(release_dir)
    if stage in {"gated", "signed", "published"}:
        run_id, result_digest = find_gate(repo, version, state["commit"])
        evidence.update(
            {"gateRunId": run_id, "gateResultSha256": result_digest}
        )
    if stage in {"signed", "published"}:
        bundle = release_dir / "checksums.sha256.sigstore.json"
        if (
            not bundle.is_file()
            or bundle.is_symlink()
            or not bundle.stat().st_size
        ):
            raise IntegrityError(
                "release signature bundle is missing or unsafe"
            )
        evidence["signatureBundleSha256"] = digest(bundle)
    state["stages"][stage] = evidence
    write_state(repo, version, state)
    print(f"release stage recorded: v{version} {stage}")


def require(repo: Path, version: str, stage: str) -> None:
    state = load_state(repo, version)
    validate_frozen(repo, state)
    if stage not in state.get("stages", {}):
        raise IntegrityError(f"release stage has not completed: {stage}")
    expected = state["stages"][stage]
    release_dir = repo / "dist" / "release" / version
    if stage != "frozen":
        current = verify_checksums(release_dir)
        if current != expected.get("checksumsSha256"):
            raise IntegrityError(
                "release checksum manifest changed after stage completion"
            )
    if stage in {"gated", "signed", "published"}:
        run_id, result_digest = find_gate(repo, version, state["commit"])
        if run_id != expected.get("gateRunId") or result_digest != expected.get(
            "gateResultSha256"
        ):
            raise IntegrityError(
                "host gate evidence changed after stage completion"
            )
    if stage in {"signed", "published"}:
        bundle = release_dir / "checksums.sha256.sigstore.json"
        if digest(bundle) != expected.get("signatureBundleSha256"):
            raise IntegrityError(
                "signature bundle changed after stage completion"
            )
    print(f"release integrity verified: v{version} {stage}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=("freeze", "record", "require"))
    parser.add_argument("--version", required=True)
    parser.add_argument("--stage", choices=STAGES)
    parser.add_argument("--repo", type=Path)
    args = parser.parse_args()
    args.version = args.version.removeprefix("v")
    if not SEMVER.fullmatch(args.version):
        parser.error("version must be strict SemVer")
    if args.command != "freeze" and args.stage is None:
        parser.error("--stage is required for record and require")
    return args


def main() -> int:
    args = parse_args()
    repo = (args.repo or Path(__file__).resolve().parent.parent).resolve()
    try:
        if args.command == "freeze":
            freeze(repo, args.version)
        elif args.command == "record":
            record(repo, args.version, args.stage)
        else:
            require(repo, args.version, args.stage)
    except IntegrityError as error:
        print(f"release integrity failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
