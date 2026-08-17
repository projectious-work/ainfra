"""Tests for the commit-bound release integrity state machine."""

from __future__ import annotations

import hashlib
import importlib.util
import json
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "release_integrity", ROOT / "scripts" / "release-integrity.py"
)
assert SPEC is not None and SPEC.loader is not None
integrity = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(integrity)


class ReleaseIntegrityTest(unittest.TestCase):
    """Exercise immutable identity and ordered evidence transitions."""

    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.repo = Path(self.temporary.name)
        self.version = "9.8.7-test"
        self._git("init", "-q")
        self._git("config", "user.name", "Release Test")
        self._git("config", "user.email", "release@example.invalid")
        (self.repo / "docs" / "releases").mkdir(parents=True)
        (self.repo / "docs" / "releases" / f"v{self.version}.md").write_text(
            "# Release\n", encoding="utf-8"
        )
        (self.repo / "CHANGELOG.md").write_text(
            f"## [{self.version}] - 2099-01-01\n", encoding="utf-8"
        )
        (self.repo / ".gitignore").write_text("/dist/\n/tmp/\n")
        self._git("add", ".")
        self._git("commit", "-qm", "release fixture")
        self._git("commit", "--allow-empty", "-qm", "canonical release")
        self.commit = self._git("rev-parse", "HEAD")
        self._align_lanes(self.commit)

    def _git(self, *args: str) -> str:
        return subprocess.check_output(
            ["git", *args], cwd=self.repo, text=True
        ).strip()

    def _align_lanes(self, commit: str) -> None:
        for lane in integrity.LANES:
            self._git("update-ref", f"refs/remotes/origin/{lane}", commit)

    def _package(self) -> None:
        release = self.repo / "dist" / "release" / self.version
        release.mkdir(parents=True)
        artifact = release / "artifact.tar.gz"
        artifact.write_bytes(b"artifact")
        checksum = hashlib.sha256(artifact.read_bytes()).hexdigest()
        (release / "checksums.sha256").write_text(
            f"{checksum}  {artifact.name}\n", encoding="ascii"
        )

    def _gate(self) -> None:
        evidence = (
            self.repo
            / "tmp"
            / "container-gate"
            / self.version
            / "run-1"
            / "evidence"
        )
        evidence.mkdir(parents=True)
        result = {
            "ok": True,
            "cleanupOk": True,
            "runId": "run-1",
            "provenance": {
                "sourceCommit": self.commit,
                "dirtyWorktree": False,
            },
        }
        (evidence / "result.json").write_text(json.dumps(result))
        (evidence / "grype.json").write_text('{"matches": []}\n')

    def test_freeze_rejects_misaligned_release_lane(self) -> None:
        self._git("update-ref", "refs/remotes/origin/v1.x-release", "HEAD^")

        with self.assertRaisesRegex(integrity.IntegrityError, "not aligned"):
            integrity.freeze(self.repo, self.version)

    def test_stages_require_ordered_commit_bound_evidence(self) -> None:
        integrity.freeze(self.repo, self.version)
        self._package()

        with self.assertRaisesRegex(
            integrity.IntegrityError, "requires packaged"
        ):
            integrity.record(self.repo, self.version, "gated")

        integrity.record(self.repo, self.version, "packaged")
        integrity.record(self.repo, self.version, "packaged")
        self._gate()
        integrity.record(self.repo, self.version, "gated")
        integrity.require(self.repo, self.version, "gated")

        state = json.loads(
            integrity.state_path(self.repo, self.version).read_text()
        )
        self.assertEqual(state["commit"], self.commit)
        self.assertEqual(state["stages"]["gated"]["gateRunId"], "run-1")

    def test_artifact_mutation_invalidates_completed_stage(self) -> None:
        integrity.freeze(self.repo, self.version)
        self._package()
        integrity.record(self.repo, self.version, "packaged")
        artifact = (
            self.repo / "dist" / "release" / self.version / "artifact.tar.gz"
        )
        artifact.write_bytes(b"changed")

        with self.assertRaisesRegex(
            integrity.IntegrityError, "checksum mismatch"
        ):
            integrity.require(self.repo, self.version, "packaged")

    def test_publication_preflights_and_supports_safe_resume(self) -> None:
        source = (ROOT / "scripts" / "publish-release.sh").read_text()

        self.assertLess(
            source.index("+ publication credential preflight"),
            source.index('git tag --annotate "$tag"'),
        )
        self.assertIn('git rev-list -n 1 "$tag"', source)
        self.assertIn(
            'gh release upload "$tag" "${assets[@]}" --clobber', source
        )
        self.assertIn("--stage=published", source)


if __name__ == "__main__":
    unittest.main()
