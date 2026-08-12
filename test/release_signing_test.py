"""Tests for the non-publishing and keyless release-signing boundaries."""

from __future__ import annotations

import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


REPO = Path(__file__).resolve().parents[1]
SCRIPT = REPO / "scripts" / "release-sign-checksums"
MAINTAIN = REPO / "scripts" / "maintain.sh"


class ReleaseSigningTest(unittest.TestCase):
    """Exercise the shell entrypoint without contacting Sigstore."""

    def setUp(self) -> None:
        """Create an isolated fake Cosign and version-specific manifest."""
        self.version = "99.98.97-test"
        self.release_dir = REPO / "dist" / "release" / self.version
        self.release_dir.mkdir(parents=True)
        self.manifest = self.release_dir / "checksums.sha256"
        self.manifest.write_text("0" * 64 + "  artifact.tar.gz\n")
        self.temp_dir = Path(tempfile.mkdtemp(prefix="ainfra-sign-test-"))
        self.log = self.temp_dir / "cosign.log"
        fake = self.temp_dir / "cosign"
        fake.write_text(
            "#!/bin/sh\n"
            "printf '%s\\n' \"$*\" >> \"$COSIGN_TEST_LOG\"\n"
            "if [ \"$1\" = sign-blob ]; then\n"
            "  while [ \"$#\" -gt 0 ]; do\n"
            "    if [ \"$1\" = --bundle ]; then\n"
            "      shift\n"
            "      printf '{}\\n' > \"$1\"\n"
            "    fi\n"
            "    shift\n"
            "  done\n"
            "fi\n"
        )
        fake.chmod(0o755)
        self.env = {
            **os.environ,
            "PATH": f"{self.temp_dir}:{os.environ['PATH']}",
            "COSIGN_TEST_LOG": str(self.log),
        }

    def tearDown(self) -> None:
        """Remove only the test's unique ignored output and temporary tools."""
        shutil.rmtree(self.release_dir)
        shutil.rmtree(self.temp_dir)

    def run_script(self, *extra: str) -> subprocess.CompletedProcess[str]:
        """Run the signing entrypoint for the isolated test version."""
        return subprocess.run(
            [str(SCRIPT), f"--version={self.version}", *extra],
            cwd=REPO,
            env=self.env,
            check=False,
            capture_output=True,
            text=True,
        )

    def test_dry_run_does_not_invoke_cosign_or_create_bundle(self) -> None:
        """A dry-run validates inputs without creating a public signature."""
        completed = self.run_script("--dry-run")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertFalse(self.log.exists())
        self.assertFalse(
            (self.release_dir / "checksums.sha256.sigstore.json").exists()
        )

    def test_signs_then_verifies_with_pinned_identity(self) -> None:
        """Normal execution signs once and immediately verifies fixed claims."""
        completed = self.run_script()

        self.assertEqual(completed.returncode, 0, completed.stderr)
        calls = self.log.read_text().splitlines()
        self.assertEqual(len(calls), 2)
        self.assertTrue(calls[0].startswith("sign-blob "))
        self.assertIn("verify-blob", calls[1])
        self.assertIn("--certificate-identity info@projectious.work", calls[1])
        self.assertIn(
            "--certificate-oidc-issuer https://github.com/login/oauth",
            calls[1],
        )

    def test_maintain_dispatches_non_signing_dry_run(self) -> None:
        """The public maintainer command exposes the safe signing preflight."""
        completed = subprocess.run(
            [
                str(MAINTAIN),
                "release-sign",
                f"--version={self.version}",
                "--dry-run",
            ],
            cwd=REPO,
            env=self.env,
            check=False,
            capture_output=True,
            text=True,
        )

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertIn("no signature or transparency entry", completed.stdout)
        self.assertFalse(self.log.exists())


if __name__ == "__main__":
    unittest.main()
