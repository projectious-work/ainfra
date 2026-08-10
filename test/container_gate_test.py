from __future__ import annotations

import hashlib
import importlib.util
import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "container_gate_host", ROOT / "scripts" / "container-gate-host.py"
)
assert SPEC is not None and SPEC.loader is not None
host = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(host)


class ContainerGateValidationTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.gate_root = Path(self.temporary.name) / "tmp" / "container-gate"
        self.run_dir = (
            self.gate_root / "20260808T010203Z-0123456789abcdef0123456789abcdef"
        )
        self.input_dir = self.run_dir / "input"
        context = self.input_dir / "context" / "dist" / "linux" / "arm64"
        context.mkdir(parents=True)
        (self.input_dir / "Dockerfile").write_text(
            "FROM scratch\n", encoding="utf-8"
        )
        (context / "ainfra").write_bytes(b"binary")
        provenance = {
            "schemaVersion": 1,
            "sourceCommit": "a" * 40,
            "branch": "v1.x-dev",
            "dirtyWorktree": True,
            "diffSha256": "b" * 64,
            "preparedAt": "2026-08-08T01:02:03Z",
        }
        (self.input_dir / "provenance.json").write_text(
            json.dumps(provenance), encoding="utf-8"
        )
        checked = [
            self.input_dir / "Dockerfile",
            context / "ainfra",
        ]
        manifest = "".join(
            f"{hashlib.sha256(path.read_bytes()).hexdigest()}  "
            f"{path.relative_to(self.input_dir).as_posix()}\n"
            for path in checked
        )
        (self.input_dir / "checksums.sha256").write_text(
            manifest, encoding="ascii"
        )
        (self.run_dir / "runtime").mkdir(mode=0o700)
        self._make_immutable()

    def _make_immutable(self) -> None:
        for path in sorted(self.input_dir.rglob("*"), reverse=True):
            if path.is_symlink():
                continue
            path.chmod(0o555 if path.is_dir() else 0o444)
        self.input_dir.chmod(0o555)
        self.run_dir.chmod(0o700)

    def _make_mutable(self) -> None:
        for path in [self.input_dir, *self.input_dir.rglob("*")]:
            path.chmod(0o755 if path.is_dir() else 0o644)

    def _run_main_with_fake_tools(
        self, *, fail_build: bool = False
    ) -> tuple[dict[str, object], list[list[str]]]:
        script = (
            Path(self.temporary.name) / "scripts" / "container-gate-host.py"
        )
        script.parent.mkdir()
        script.write_text("test fixture\n", encoding="utf-8")
        commands: list[list[str]] = []
        image_exists = False

        def fake_execute(
            log: object,
            argv: list[str],
            *,
            env: dict[str, str],
            output: Path,
        ) -> subprocess.CompletedProcess[bytes]:
            nonlocal image_exists
            del log, env
            commands.append(argv)
            stdout = b""
            if len(argv) == 2 and argv[1] == "version":
                stdout = b"fixture version\n"
            elif argv[1:3] == ["image", "ls"]:
                stdout = b"image-id\n" if image_exists else b""
            elif argv[1] == "build":
                image_exists = True
                if fail_build:
                    output.write_bytes(b"fixture build failure\n")
                    raise host.GateError("fixture build failure")
            elif argv[1:3] == ["image", "rm"]:
                image_exists = False
            elif argv[1:3] == ["image", "inspect"]:
                stdout = b"65532\n" if "--format" in argv else b"{}\n"
            elif argv[1] == "run":
                stdout = b"{}\n"
            elif Path(argv[0]).name == "syft":
                sbom = Path(argv[argv.index("-o") + 1].split("=", 1)[1])
                sbom.write_text("{}\n", encoding="utf-8")
            elif Path(argv[0]).name == "grype":
                stdout = b"{}\n"
            output.write_bytes(stdout)
            return subprocess.CompletedProcess(argv, 0, stdout)

        def fake_tool(name: str) -> Path:
            return Path("/usr/bin") / name

        argv = [str(script), str(self.run_dir)]
        with (
            mock.patch.object(host, "__file__", str(script)),
            mock.patch.object(host, "resolve_tool", side_effect=fake_tool),
            mock.patch.object(
                host,
                "validate_bootstrap",
                return_value={"invocation": "fixture uv invocation"},
            ),
            mock.patch.object(host, "execute", side_effect=fake_execute),
            mock.patch.object(host.platform, "machine", return_value="arm64"),
            mock.patch.object(host.sys, "argv", argv),
        ):
            if fail_build:
                with self.assertRaisesRegex(host.GateError, "build failure"):
                    host.main()
            else:
                self.assertEqual(host.main(), 0)

        result = json.loads(
            (self.run_dir / "evidence" / "result.json").read_text(
                encoding="utf-8"
            )
        )
        return result, commands

    def test_accepts_complete_immutable_snapshot(self) -> None:
        input_dir, provenance = host.validate_tree(
            self.run_dir.resolve(), self.gate_root.resolve()
        )
        self.assertEqual(input_dir, self.input_dir)
        self.assertTrue(provenance["dirtyWorktree"])

    def test_rejects_existing_evidence(self) -> None:
        (self.run_dir / "evidence").mkdir()
        with self.assertRaisesRegex(host.GateError, "exactly input"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_accepts_empty_private_runtime(self) -> None:
        host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_python_tree_validation_defers_runtime_contents(self) -> None:
        runtime = self.run_dir / "runtime"
        (runtime / "unexpected").write_text("state\n", encoding="utf-8")
        host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_bootstrap_validation_rejects_nonprivate_runtime(self) -> None:
        runtime = self.run_dir / "runtime"
        runtime.chmod(0o755)
        with self.assertRaisesRegex(host.GateError, "only bootstrap"):
            host.validate_bootstrap(self.run_dir.resolve(), ROOT)

    def test_rejects_runtime_symlink(self) -> None:
        target = Path(self.temporary.name) / "runtime-target"
        target.mkdir(mode=0o700)
        (self.run_dir / "runtime").rmdir()
        (self.run_dir / "runtime").symlink_to(target)
        with self.assertRaises((host.GateError, NotADirectoryError)):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_incomplete_checksum_manifest(self) -> None:
        self._make_mutable()
        manifest = self.input_dir / "checksums.sha256"
        manifest.write_text(
            manifest.read_text(encoding="ascii").splitlines()[0] + "\n",
            encoding="ascii",
        )
        self._make_immutable()
        with self.assertRaisesRegex(host.GateError, "complete build input"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_symlink(self) -> None:
        self._make_mutable()
        link = self.input_dir / "context" / "escape"
        link.symlink_to("/tmp")
        self._make_immutable()
        with self.assertRaisesRegex(host.GateError, "symlink rejected"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_hard_link(self) -> None:
        self._make_mutable()
        binary = (
            self.input_dir / "context" / "dist" / "linux" / "arm64" / "ainfra"
        )
        os.link(binary, self.input_dir / "context" / "duplicate")
        self._make_immutable()
        with self.assertRaisesRegex(host.GateError, "hard-linked"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_world_writable_input(self) -> None:
        target = self.input_dir / "Dockerfile"
        target.chmod(0o446)
        with self.assertRaisesRegex(host.GateError, "world-writable"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_owner_writable_input(self) -> None:
        target = self.input_dir / "Dockerfile"
        target.chmod(0o644)
        with self.assertRaisesRegex(host.GateError, "must be immutable"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_unknown_provenance_field(self) -> None:
        self._make_mutable()
        path = self.input_dir / "provenance.json"
        provenance = json.loads(path.read_text(encoding="utf-8"))
        provenance["unexpected"] = True
        path.write_text(json.dumps(provenance), encoding="utf-8")
        self._make_immutable()
        with self.assertRaisesRegex(host.GateError, "unknown fields"):
            host.validate_tree(self.run_dir.resolve(), self.gate_root.resolve())

    def test_rejects_non_direct_child(self) -> None:
        nested = self.gate_root / "nested" / self.run_dir.name
        nested.mkdir(parents=True)
        with self.assertRaisesRegex(host.GateError, "direct child"):
            host.validate_tree(nested.resolve(), self.gate_root.resolve())

    def test_host_script_uses_no_shell_or_privileged_escape(self) -> None:
        source = (ROOT / "scripts" / "container-gate-host.py").read_text(
            encoding="utf-8"
        )
        launcher = (ROOT / "scripts" / "container-gate-host").read_text(
            encoding="utf-8"
        )
        self.assertTrue(source.startswith("#!/usr/bin/env python3"))
        self.assertIn('readonly PYTHON_REQUEST="3.13.14"', launcher)
        self.assertIn("python install", launcher)
        self.assertIn("acquisition=uv-managed", launcher)
        self.assertNotIn("approved-python.sha256", launcher)
        self.assertIn("--no-project", launcher)
        self.assertIn("Darwin) state_root=", launcher)
        self.assertIn("Linux) state_root=", launcher)
        self.assertIn("/usr/bin/shasum -a 256", launcher)
        self.assertIn("/usr/bin/sha256sum", launcher)
        self.assertIn("/home/linuxbrew/.linuxbrew/bin/uv", launcher)
        forbidden = (
            "shell=True",
            "os.system",
            "--privileged",
            "--pid=host",
            "--network=host",
            "/var/run/docker.sock",
            "docker push",
            "docker system prune",
        )
        for token in forbidden:
            with self.subTest(token=token):
                self.assertNotIn(token, source + launcher)

    def test_missing_tools_offer_macos_homebrew_guidance(self) -> None:
        def fake_resolve(name: str) -> Path:
            if name in {"syft", "grype"}:
                raise host.GateError(f"missing fixture tool: {name}")
            return Path("/opt/homebrew/bin") / name

        with (
            mock.patch.object(host, "resolve_tool", side_effect=fake_resolve),
            self.assertRaises(host.GateError) as raised,
        ):
            host.resolve_required_tools("Darwin")
        message = str(raised.exception)
        self.assertIn("Detected macOS", message)
        self.assertIn("brew install syft grype", message)
        self.assertIn("missing fixture tool: syft", message)

    def test_missing_tools_offer_linux_guidance(self) -> None:
        """Linux failures provide guidance without running an installer."""

        with (
            mock.patch.object(
                host,
                "resolve_tool",
                side_effect=host.GateError("fixture tool unavailable"),
            ),
            self.assertRaises(host.GateError) as raised,
        ):
            host.resolve_required_tools("Linux")
        message = str(raised.exception)
        self.assertIn("Detected Linux", message)
        self.assertIn("official repository", message)
        self.assertIn("/home/linuxbrew/.linuxbrew/bin", message)

    def test_missing_tool_preflight_records_failed_evidence(self) -> None:
        script = (
            Path(self.temporary.name) / "scripts" / "container-gate-host.py"
        )
        script.parent.mkdir()
        script.write_text("test fixture\n", encoding="utf-8")
        argv = [str(script), str(self.run_dir)]
        unavailable = host.GateError("fixture tools unavailable")
        with (
            mock.patch.object(host, "__file__", str(script)),
            mock.patch.object(
                host,
                "validate_bootstrap",
                return_value={"invocation": "fixture uv invocation"},
            ),
            mock.patch.object(
                host, "resolve_required_tools", side_effect=unavailable
            ),
            mock.patch.object(host.sys, "argv", argv),
            self.assertRaisesRegex(host.GateError, "tools unavailable"),
        ):
            host.main()
        result = json.loads(
            (self.run_dir / "evidence" / "result.json").read_text()
        )
        self.assertFalse(result["ok"])
        self.assertIn("tools unavailable", result["error"])

    def test_success_records_result_and_removes_exact_image(self) -> None:
        result, commands = self._run_main_with_fake_tools()
        self.assertTrue(result["ok"])
        self.assertTrue(result["cleanupOk"])
        self.assertEqual(result["pythonDependencies"], [])
        self.assertIn("pythonExecutable", result)
        self.assertTrue((self.run_dir / "runtime" / "tool-state").is_dir())
        self.assertFalse((self.run_dir / "evidence" / "tool-state").exists())
        removals = [argv for argv in commands if argv[1:3] == ["image", "rm"]]
        self.assertEqual(len(removals), 1)
        self.assertEqual(
            removals[0][-1],
            f"ainfra-container-gate:{self.run_dir.name.lower()}",
        )

    def test_failed_build_records_failure_and_still_cleans_up(self) -> None:
        result, commands = self._run_main_with_fake_tools(fail_build=True)
        self.assertFalse(result["ok"])
        self.assertTrue(result["cleanupOk"])
        self.assertIn("fixture build failure", result["error"])
        self.assertTrue(any(argv[1:3] == ["image", "rm"] for argv in commands))


if __name__ == "__main__":
    unittest.main()
