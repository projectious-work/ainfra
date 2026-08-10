#!/usr/bin/env python3
"""Run the owner-reviewed, host-only Phase 1 container validation gate.

The reviewed ``container-gate-host`` launcher provides an exact, uv-managed
Python before this entrypoint starts. This script is
the narrow trust boundary between an
immutable build-input snapshot prepared in the development harness and
container tooling installed on an owner's host. It validates the snapshot
before invoking Docker, Syft, or Grype, records every command and result below
the run directory, and removes only the image tag derived from that run's
identifier.

The script deliberately accepts no configuration other than one run-directory
path.  In particular, it does not inherit commands, image names, build
arguments, mounts, or tool-state locations from the caller.  Keep those
constraints intact when maintaining this file.
"""

from __future__ import annotations

import datetime as dt
import hashlib
import json
import os
import platform
import re
import shlex
import stat
import subprocess
import sys
from pathlib import Path
from typing import IO, Any

# A timestamp makes runs recognizable to humans; 128 random bits prevent two
# preparations in the same second from sharing an evidence directory or tag.
RUN_ID = re.compile(r"^[0-9]{8}T[0-9]{6}Z-[0-9a-f]{32}$")
HEX_64 = re.compile(r"^[0-9a-f]{64}$")
HEX_40 = re.compile(r"^[0-9a-f]{40}$")

# Tool lookup is intentionally independent of PATH.  This prevents a caller
# from injecting a different executable through its environment.  Both common
# Linux locations and the two Homebrew locations used on macOS are explicit.
TOOL_DIRS = (
    Path("/home/linuxbrew/.linuxbrew/bin"),
    Path("/usr/local/bin"),
    Path("/usr/bin"),
    Path("/bin"),
    Path("/opt/homebrew/bin"),
)


class GateError(RuntimeError):
    """Report a validation or execution condition that fails the gate."""


def sha256(path: Path) -> str:
    """Return the lowercase SHA-256 digest of a regular input file.

    Args:
        path: File whose bytes are hashed.

    Returns:
        The file digest as 64 lowercase hexadecimal characters.

    Raises:
        OSError: If the file cannot be opened or read.
    """

    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def resolve_tool(name: str) -> Path:
    """Resolve one required executable from the fixed trusted directories.

    Symlinks in standard installation paths are allowed, but their final
    target must be a non-world-writable, executable regular file.

    Args:
        name: Basename of the required executable.

    Returns:
        The canonical executable path.

    Raises:
        GateError: If no acceptable executable is available.
        OSError: If candidate metadata cannot be inspected.
    """

    for directory in TOOL_DIRS:
        candidate = directory / name
        if not candidate.exists():
            continue
        resolved = candidate.resolve(strict=True)
        metadata = resolved.stat()
        if not stat.S_ISREG(metadata.st_mode):
            raise GateError(f"{name} is not a regular file: {resolved}")
        if metadata.st_mode & stat.S_IWOTH:
            raise GateError(f"{name} is world-writable: {resolved}")
        if not os.access(resolved, os.X_OK):
            raise GateError(f"{name} is not executable: {resolved}")
        return resolved
    raise GateError(f"required tool is unavailable in approved paths: {name}")


def installation_guidance(missing: list[str], system: str) -> str:
    """Return non-executing, OS-specific installation guidance.

    The text is intentionally advisory: the trusted gate never invokes a
    package manager or elevates privileges on behalf of the operator.

    Args:
        missing: Required executable basenames that were not resolved.
        system: Host operating-system name returned by ``platform.system``.

    Returns:
        A human-readable error suffix with one installation possibility.
    """

    names = ", ".join(missing)
    lines = [f"missing required host tools: {names}"]
    if system == "Darwin":
        lines.append(
            "Detected macOS. One installation possibility is Homebrew:"
        )
        if "uv" in missing:
            lines.append("  brew install uv")
        if "docker" in missing:
            lines.extend(
                (
                    "  brew install --cask docker",
                    "  Then start Docker Desktop, install its CLI tools into",
                    "  /usr/local/bin, and enable the default Docker socket.",
                )
            )
        scanners = [name for name in ("syft", "grype") if name in missing]
        if scanners:
            lines.append(f"  brew install {' '.join(scanners)}")
    elif system == "Linux":
        lines.extend(
            (
                "Detected Linux. Install Docker Engine using your vendor's",
                "official repository and install uv, Syft, and Grype into",
                "one of: /home/linuxbrew/.linuxbrew/bin, /usr/local/bin,",
                "/usr/bin, /bin.",
            )
        )
    else:
        lines.append(
            f"Detected {system or 'an unknown OS'}; consult the official "
            "installation documentation for each missing tool."
        )
    return "\n".join(lines)


def resolve_required_tools(system: str) -> dict[str, Path]:
    """Resolve every host tool before creating an evidence directory.

    Args:
        system: Host operating-system name returned by ``platform.system``.

    Returns:
        Canonical paths keyed by executable basename.

    Raises:
        GateError: If one or more required tools are unavailable or unsafe.
    """

    tools: dict[str, Path] = {}
    errors: dict[str, str] = {}
    for name in ("docker", "syft", "grype"):
        try:
            tools[name] = resolve_tool(name)
        except GateError as error:
            errors[name] = str(error)
    if errors:
        guidance = installation_guidance(list(errors), system)
        detail = "\n".join(f"  {name}: {errors[name]}" for name in errors)
        raise GateError(f"{guidance}\nResolution details:\n{detail}")
    return tools


def parse_bootstrap_manifest(path: Path) -> dict[str, str]:
    """Parse the launcher's closed, line-oriented bootstrap manifest.

    Args:
        path: Canonical ``runtime/bootstrap/manifest.env`` path.

    Returns:
        Manifest values keyed by their fixed field names.

    Raises:
        GateError: If a line, key set, or value is malformed.
        OSError: If the manifest cannot be read.
    """

    expected = {
        "schemaVersion", "hostSystem", "launcherPath", "launcherSha256", "uvPath",
        "uvSha256", "uvVersion", "pythonRequest", "pythonPath",
        "pythonSha256", "pythonIdentity", "pythonEntry",
        "pythonEntrySha256", "uvCacheDir", "pythonInstallDir", "venv",
        "acquisition", "dependencies", "invocation",
    }
    values: dict[str, str] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        key, separator, value = line.partition("=")
        if not separator or key not in expected or key in values or not value:
            raise GateError("bootstrap manifest is malformed")
        values[key] = value
    if set(values) != expected or values["schemaVersion"] != "1":
        raise GateError("bootstrap manifest has missing or unknown fields")
    return values


def validate_bootstrap(run_dir: Path, repo: Path) -> dict[str, Any]:
    """Validate the first-stage runtime and return evidence-safe identity.

    Args:
        run_dir: Canonical prepared run directory.
        repo: Canonical repository root containing both reviewed stages.

    Returns:
        A summary of the verified launcher, uv, Python, and invocation.

    Raises:
        GateError: If runtime layout, hashes, environment, or interpreter
            identity differs from the launcher's manifest.
        OSError: If bootstrap files or executable identities cannot be read.
    """

    runtime = run_dir / "runtime"
    bootstrap = runtime / "bootstrap"
    if set(path.name for path in runtime.iterdir()) != {"bootstrap"}:
        raise GateError("runtime must contain only bootstrap/ at entry")
    if stat.S_IMODE(runtime.lstat().st_mode) != 0o700:
        raise GateError("runtime directory must use mode 0700")
    if set(path.name for path in bootstrap.iterdir()) != {
        "bootstrap.log", "manifest.env", "venv",
    }:
        raise GateError("bootstrap directory has unexpected entries")
    manifest_path = bootstrap / "manifest.env"
    log_path = bootstrap / "bootstrap.log"
    for path in (manifest_path, log_path):
        metadata = path.lstat()
        if not stat.S_ISREG(metadata.st_mode) or metadata.st_nlink != 1:
            raise GateError(f"unsafe bootstrap file: {path.name}")
        if stat.S_IMODE(metadata.st_mode) != 0o600:
            raise GateError(f"bootstrap file must use mode 0600: {path.name}")

    values = parse_bootstrap_manifest(manifest_path)
    launcher = (repo / "scripts" / "container-gate-host").resolve(strict=True)
    entry = Path(__file__).resolve(strict=True)
    required = {
        "launcherPath": str(launcher),
        "launcherSha256": sha256(launcher),
        "pythonEntry": str(entry),
        "pythonEntrySha256": sha256(entry),
        "pythonPath": str(Path(sys.executable).resolve(strict=True)),
        "pythonSha256": sha256(Path(sys.executable).resolve(strict=True)),
        "venv": sys.prefix,
        "acquisition": "uv-managed",
        "dependencies": "stdlib-only",
        "hostSystem": platform.system(),
    }
    for key, expected in required.items():
        if values[key] != expected:
            raise GateError(f"bootstrap manifest mismatch: {key}")
    if sys.prefix == sys.base_prefix:
        raise GateError("bootstrap did not provide an isolated environment")
    approved_uv_prefixes = (
        str(Path.home() / ".local" / "bin") + "/",
        "/opt/homebrew/bin/",
        "/home/linuxbrew/.linuxbrew/bin/",
        "/usr/local/bin/",
        "/usr/bin/",
    )
    if not values["uvPath"].startswith(approved_uv_prefixes):
        raise GateError("bootstrap uv path is outside approved locations")
    if sha256(Path(values["uvPath"]).resolve(strict=True)) != values["uvSha256"]:
        raise GateError("bootstrap uv digest mismatch")

    expected_env = {
        "UV_CACHE_DIR": values["uvCacheDir"],
        "UV_PYTHON_INSTALL_DIR": values["pythonInstallDir"],
        "VIRTUAL_ENV": values["venv"],
        "AINFRA_CONTAINER_GATE_BOOTSTRAP": str(manifest_path),
    }
    for name, expected in expected_env.items():
        if os.environ.get(name) != expected:
            raise GateError(f"bootstrap environment mismatch: {name}")
    return {
        **values,
        "manifestSha256": sha256(manifest_path),
        "logSha256": sha256(log_path),
        "logPath": str(log_path),
    }


def verify_immutable_input(input_dir: Path) -> None:
    """Reject mutable or aliased entries anywhere in the input snapshot.

    ``lstat`` is important here: following a symlink before checking its type
    would allow input to escape the prepared tree.  Regular files must also
    have one link so another pathname cannot mutate the same inode.

    Args:
        input_dir: Canonical ``input/`` directory to inspect recursively.

    Raises:
        GateError: If an entry is writable, linked, or not a plain file or
            directory.
        OSError: If filesystem metadata cannot be read.
    """

    for path in [input_dir, *input_dir.rglob("*")]:
        metadata = path.lstat()
        if stat.S_ISLNK(metadata.st_mode):
            raise GateError(f"symlink rejected: {path}")
        if not (
            stat.S_ISDIR(metadata.st_mode) or stat.S_ISREG(metadata.st_mode)
        ):
            raise GateError(f"special file rejected: {path}")
        if metadata.st_mode & stat.S_IWOTH:
            raise GateError(f"world-writable input rejected: {path}")
        if metadata.st_mode & 0o222:
            raise GateError(f"input must be immutable: {path}")
        if stat.S_ISREG(metadata.st_mode) and metadata.st_nlink != 1:
            raise GateError(f"hard-linked regular file rejected: {path}")


def verify_manifest(input_dir: Path) -> dict[str, str]:
    """Validate complete checksum coverage and every recorded file digest.

    The accepted manifest grammar names only the Dockerfile and files below
    ``context/``.  Equality between the manifest and discovered file sets
    makes omission as fatal as a checksum mismatch.

    Args:
        input_dir: Immutable input directory containing the manifest.

    Returns:
        A mapping from validated relative paths to SHA-256 digests.

    Raises:
        GateError: If the tree is mutable or the manifest is unsafe,
            incomplete, duplicated, malformed, or incorrect.
        OSError: If input files cannot be read.
    """

    # Recheck immutability on every manifest verification.  The second call at
    # the end of the gate detects ordinary accidental changes during the run.
    verify_immutable_input(input_dir)
    manifest_path = input_dir / "checksums.sha256"
    manifest: dict[str, str] = {}
    for line in manifest_path.read_text(encoding="ascii").splitlines():
        match = re.fullmatch(
            r"([0-9a-f]{64})  (Dockerfile|context/[A-Za-z0-9._/-]+)", line
        )
        if match is None or ".." in Path(match.group(2)).parts:
            raise GateError(
                "checksums.sha256 contains an unsafe or malformed line"
            )
        if match.group(2) in manifest:
            raise GateError("checksums.sha256 contains a duplicate path")
        manifest[match.group(2)] = match.group(1)

    # Build the actual file set independently rather than trusting the list in
    # the manifest to define what should be checked.
    actual = {"Dockerfile"}
    actual.update(
        path.relative_to(input_dir).as_posix()
        for path in (input_dir / "context").rglob("*")
        if path.is_file()
    )
    if set(manifest) != actual:
        raise GateError(
            "checksum manifest does not cover the complete build input"
        )
    for relative, expected in manifest.items():
        if sha256(input_dir / relative) != expected:
            raise GateError(f"checksum mismatch: {relative}")
    return manifest


def validate_tree(
    run_dir: Path, gate_root: Path
) -> tuple[Path, dict[str, Any]]:
    """Validate the run layout and return trusted input and provenance.

    Both arguments must already be canonical paths.  The direct-parent check
    prevents traversal and ensures evidence can only be created beneath the
    repository's dedicated ``tmp/container-gate/`` root.

    Args:
        run_dir: Canonical candidate run directory supplied by the operator.
        gate_root: Canonical repository container-gate directory.

    Returns:
        The validated input directory and parsed provenance object.

    Raises:
        GateError: If layout, permissions, provenance, or checksums fail.
        OSError: If filesystem content or metadata cannot be inspected.
    """

    # Exact set comparisons fail closed on both missing and extra entries.
    if run_dir.parent != gate_root:
        raise GateError(
            "run directory must be a direct child of tmp/container-gate"
        )
    if not RUN_ID.fullmatch(run_dir.name):
        raise GateError(
            "run directory basename is not a collision-resistant run ID"
        )
    run_entries = set(path.name for path in run_dir.iterdir())
    if run_entries != {"input", "runtime"}:
        raise GateError(
            "Python stage requires exactly input/ and launcher-created "
            "runtime/"
        )
    runtime_metadata = (run_dir / "runtime").lstat()
    if not stat.S_ISDIR(runtime_metadata.st_mode):
        raise GateError("runtime path is not a directory")

    input_dir = run_dir / "input"
    if set(path.name for path in input_dir.iterdir()) != {
        "provenance.json",
        "checksums.sha256",
        "Dockerfile",
        "context",
    }:
        raise GateError("input/ has missing or unexpected entries")

    run_metadata = run_dir.lstat()
    if not stat.S_ISDIR(run_metadata.st_mode):
        raise GateError("run path is not a directory")
    if run_metadata.st_mode & stat.S_IWOTH:
        raise GateError(f"world-writable run directory rejected: {run_dir}")
    verify_immutable_input(input_dir)

    # Provenance is data for review, never executable configuration.  Its
    # schema is intentionally closed so later fields cannot silently influence
    # this script without an explicit code review.
    provenance_path = input_dir / "provenance.json"
    try:
        provenance = json.loads(provenance_path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as error:
        raise GateError(f"invalid provenance.json: {error}") from error
    expected_keys = {
        "schemaVersion",
        "sourceCommit",
        "branch",
        "dirtyWorktree",
        "diffSha256",
        "preparedAt",
    }
    if not isinstance(provenance, dict) or set(provenance) != expected_keys:
        raise GateError("provenance.json has missing or unknown fields")
    if provenance["schemaVersion"] != 1:
        raise GateError("unsupported provenance schemaVersion")
    if not isinstance(provenance["sourceCommit"], str) or not HEX_40.fullmatch(
        provenance["sourceCommit"]
    ):
        raise GateError("invalid provenance sourceCommit")
    if not isinstance(provenance["branch"], str) or not provenance["branch"]:
        raise GateError("invalid provenance branch")
    if not isinstance(provenance["dirtyWorktree"], bool):
        raise GateError("invalid provenance dirtyWorktree")
    if not isinstance(provenance["diffSha256"], str) or not HEX_64.fullmatch(
        provenance["diffSha256"]
    ):
        raise GateError("invalid provenance diffSha256")
    if not isinstance(provenance["preparedAt"], str):
        raise GateError("invalid provenance preparedAt")
    try:
        dt.datetime.fromisoformat(
            provenance["preparedAt"].replace("Z", "+00:00")
        )
    except ValueError as error:
        raise GateError("invalid provenance preparedAt") from error

    verify_manifest(input_dir)
    return input_dir, provenance


def log_command(log: IO[str], argv: list[str]) -> None:
    """Print and durably append one exact argument vector before execution.

    Args:
        log: Append-mode command log owned by this run's evidence directory.
        argv: Complete command argument vector about to be executed.

    Raises:
        OSError: If the command cannot be written and synced to the log.
    """

    rendered = shlex.join(argv)
    print(f"+ {rendered}")
    log.write(f"+ {rendered}\n")
    log.flush()
    os.fsync(log.fileno())


def execute(
    log: IO[str],
    argv: list[str],
    *,
    env: dict[str, str],
    output: Path,
) -> subprocess.CompletedProcess[bytes]:
    """Execute a fixed argument vector and retain its combined output.

    No shell participates in execution.  Standard input is closed, the caller
    supplies a controlled environment, and the process runs in the evidence
    directory rather than inheriting a meaningful caller working directory.

    Args:
        log: Append-mode log receiving the exact command first.
        argv: Complete executable and argument vector.
        env: Script-controlled process environment.
        output: Evidence file receiving combined stdout and stderr.

    Returns:
        The successful completed-process record, including captured output.

    Raises:
        GateError: If the command returns a nonzero status.
        OSError: If execution or evidence writing fails.
    """

    log_command(log, argv)
    completed = subprocess.run(
        argv,
        cwd=output.parent,
        env=env,
        stdin=subprocess.DEVNULL,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        check=False,
    )
    output.write_bytes(completed.stdout)
    if completed.returncode != 0:
        raise GateError(
            f"command failed with status {completed.returncode}: {argv[0]}"
        )
    return completed


def main() -> int:
    """Validate one prepared run and produce container-gate evidence.

    Returns:
        Zero after all checks and exact-image cleanup succeed.

    Raises:
        GateError: If validation, a required check, or cleanup fails.
        OSError: If required filesystem or process operations fail.
        UnicodeError: If required text evidence cannot be decoded.
        json.JSONDecodeError: If machine-readable evidence is malformed.
    """

    # Options and extra positional arguments are forbidden.  All operational
    # values below are derived from the validated run or fixed in this script.
    if len(sys.argv) != 2 or sys.argv[1].startswith("-"):
        raise GateError("usage: scripts/container-gate-host RUN_DIRECTORY")

    # Resolve from this reviewed script, not the operator's current directory.
    repo = Path(__file__).resolve(strict=True).parent.parent
    gate_root = (repo / "tmp" / "container-gate").resolve(strict=True)
    supplied = Path(sys.argv[1])
    if supplied.is_symlink():
        raise GateError("run-directory symlinks are rejected")
    run_dir = supplied.resolve(strict=True)
    input_dir, provenance = validate_tree(run_dir, gate_root)
    bootstrap = validate_bootstrap(run_dir, repo)

    machine = platform.machine().lower()

    # Runtime is mutable execution infrastructure, never authoritative
    # evidence. The launcher already created and validated bootstrap/.
    runtime = run_dir / "runtime"

    # validate_tree requires evidence to be absent.  mkdir and touch are both
    # exclusive, so an existing or raced path fails instead of being reused.
    evidence = run_dir / "evidence"
    evidence.mkdir(mode=0o700)
    command_log_path = evidence / "commands.log"
    command_log_path.touch(mode=0o600, exist_ok=False)
    run_id = run_dir.name
    image_tag = f"ainfra-container-gate:{run_id.lower()}"
    print(f"container gate start: {run_id}")
    print(f"input: {input_dir}")
    print(f"evidence: {evidence}")

    # Isolate every writable home, cache, Docker config, scanner database, and
    # temporary directory below runtime/. Authoritative results stay separate
    # in evidence/. The child environment is constructed from scratch and
    # therefore cannot inherit caller-controlled tool configuration.
    state = runtime / "tool-state"
    for directory in (
        state / "docker",
        state / "syft",
        state / "grype",
        state / "home",
        state / "cache",
        state / "tmp",
    ):
        directory.mkdir(parents=True, mode=0o700, exist_ok=True)
    trusted_paths = {
        "Darwin": "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin",
        "Linux": (
            "/home/linuxbrew/.linuxbrew/bin:/usr/local/bin:/usr/bin:/bin"
        ),
    }
    system = platform.system()
    if system not in trusted_paths:
        raise GateError(f"unsupported host OS: {system}")
    env = {
        "HOME": str(state / "home"),
        "PATH": trusted_paths[system],
        "DOCKER_CONFIG": str(state / "docker"),
        "SYFT_CACHE_DIR": str(state / "syft"),
        "GRYPE_DB_CACHE_DIR": str(state / "grype"),
        "XDG_CACHE_HOME": str(state / "cache"),
        "TMPDIR": str(state / "tmp"),
        "LANG": "C",
        "LC_ALL": "C",
    }

    # Start metadata and verified source records are retained even when tool
    # discovery or the first command fails.
    metadata = {
        "runId": run_id,
        "input": str(input_dir),
        "evidence": str(evidence),
        "runtime": str(runtime),
        "hostArchitecture": machine,
        "bootstrap": bootstrap,
        "pythonExecutable": str(Path(sys.executable).resolve(strict=True)),
        "pythonVersion": platform.python_version(),
        "pythonImplementation": platform.python_implementation(),
        "pythonEnvironment": sys.prefix,
        "pythonBaseEnvironment": sys.base_prefix,
        "pythonDependencies": [],
        "scriptSha256": sha256(Path(__file__).resolve(strict=True)),
        "invocation": {
            "declared": bootstrap["invocation"],
            "pythonArgv": sys.orig_argv,
        },
        "startedAt": dt.datetime.now(dt.UTC).isoformat(),
        "provenance": provenance,
    }
    (evidence / "verified-checksums.sha256").write_bytes(
        (input_dir / "checksums.sha256").read_bytes()
    )
    (evidence / "verified-provenance.json").write_bytes(
        (input_dir / "provenance.json").read_bytes()
    )
    # These flags distinguish a pre-build failure (nothing to remove) from a
    # failed build that may nevertheless have created the temporary image.
    build_attempted = False
    failure: Exception | None = None
    cleanup_ok = False
    with command_log_path.open("a", encoding="utf-8") as command_log:
        try:
            # Tool discovery happens after bootstrap evidence exists. A
            # missing tool therefore produces a durable failed result instead
            # of consuming the run with no explanation.
            tools = resolve_required_tools(platform.system())
            docker = tools["docker"]
            syft = tools["syft"]
            grype = tools["grype"]
            architectures = {
                "x86_64": "amd64",
                "amd64": "amd64",
                "arm64": "arm64",
                "aarch64": "arm64",
            }
            if machine not in architectures:
                raise GateError(f"unsupported host architecture: {machine}")
            target_arch = architectures[machine]
            binary = (
                input_dir
                / "context"
                / "dist"
                / "linux"
                / target_arch
                / "ainfra"
            )
            if not binary.is_file():
                raise GateError(
                    "build context lacks the host Linux binary: "
                    f"{target_arch}"
                )

            # Capture versions through the same logged executor used for every
            # other external command.
            for tool, name in (
                (docker, "docker"),
                (syft, "syft"),
                (grype, "grype"),
            ):
                result = execute(
                    command_log,
                    [str(tool), "version"],
                    env=env,
                    output=evidence / f"{name}-version.log",
                )
                metadata[f"{name}Version"] = result.stdout.decode(
                    "utf-8", "replace"
                ).strip()

            # Revalidate immediately before Docker sees the input.
            verify_manifest(input_dir)

            # Refuse to overwrite an image that predates this invocation.  A
            # collision would make ownership and cleanup of that tag ambiguous.
            preflight = execute(
                command_log,
                [
                    str(docker),
                    "image",
                    "ls",
                    "--quiet",
                    "--filter",
                    f"reference={image_tag}",
                ],
                env=env,
                output=evidence / "preflight-image.txt",
            )
            if preflight.stdout.strip():
                raise GateError("run-specific image tag already exists")
            build_attempted = True
            # The Dockerfile and context are the only input paths passed to the
            # build.  Pulls and build-time networking are disabled.
            execute(
                command_log,
                [
                    str(docker),
                    "build",
                    "--pull=false",
                    "--network=none",
                    "--build-arg",
                    f"TARGETARCH={target_arch}",
                    "--file",
                    str(input_dir / "Dockerfile"),
                    "--tag",
                    image_tag,
                    str(input_dir / "context"),
                ],
                env=env,
                output=evidence / "build.log",
            )
            # Retain the complete image description separately from the focused
            # non-root assertion below.
            execute(
                command_log,
                [str(docker), "image", "inspect", image_tag],
                env=env,
                output=evidence / "image-inspection.json",
            )
            user_result = execute(
                command_log,
                [
                    str(docker),
                    "image",
                    "inspect",
                    "--format",
                    "{{.Config.User}}",
                    image_tag,
                ],
                env=env,
                output=evidence / "image-user.txt",
            )
            image_user = user_result.stdout.decode().strip()
            if image_user in {"", "0", "root", "0:0", "root:root"}:
                raise GateError(
                    "container image does not declare a non-root user"
                )
            # Runtime hardening is fixed here: no network, writable root
            # filesystem, capabilities, mounts, devices, or relaxed policy.
            execute(
                command_log,
                [
                    str(docker),
                    "run",
                    "--rm",
                    "--network=none",
                    "--read-only",
                    "--cap-drop=ALL",
                    "--security-opt=no-new-privileges",
                    image_tag,
                    "--format",
                    "json",
                    "version",
                ],
                env=env,
                output=evidence / "runtime-smoke.json",
            )
            # Syft writes the SPDX document directly; its own diagnostic output
            # is retained separately so neither stream obscures the other.
            execute(
                command_log,
                [
                    str(syft),
                    image_tag,
                    "-o",
                    f"spdx-json={evidence / 'sbom.spdx.json'}",
                ],
                env=env,
                output=evidence / "syft.log",
            )
            # --fail-on high also rejects critical findings because Grype's
            # threshold includes all severities at or above the named level.
            execute(
                command_log,
                [str(grype), image_tag, "--fail-on", "high", "-o", "json"],
                env=env,
                output=evidence / "grype.json",
            )
            # A zero exit status is insufficient: every required structured
            # artifact must exist, contain data, and parse as JSON.
            for json_file in (
                evidence / "image-inspection.json",
                evidence / "runtime-smoke.json",
                evidence / "sbom.spdx.json",
                evidence / "grype.json",
            ):
                if not json_file.is_file() or json_file.stat().st_size == 0:
                    raise GateError(f"missing evidence: {json_file.name}")
                json.loads(json_file.read_text(encoding="utf-8"))
            # Finish by proving the immutable source still matches the manifest.
            verify_manifest(input_dir)
        except (
            Exception
        ) as error:  # retain evidence, then clean the exact image
            failure = error
        finally:
            # A failed build can still leave an image behind, so cleanup begins
            # after any build attempt—not only after a successful build.
            if build_attempted:
                try:
                    remaining = execute(
                        command_log,
                        [
                            str(docker),
                            "image",
                            "ls",
                            "--quiet",
                            "--filter",
                            f"reference={image_tag}",
                        ],
                        env=env,
                        output=evidence / "cleanup-inspection.txt",
                    )
                    if remaining.stdout.strip():
                        # Delete only the exact tag derived from this run.
                        # Never use pruning, globs, or broader Docker cleanup.
                        execute(
                            command_log,
                            [str(docker), "image", "rm", image_tag],
                            env=env,
                            output=evidence / "cleanup.log",
                        )
                        # Prove deletion rather than trusting `docker image rm`.
                        post_cleanup = execute(
                            command_log,
                            [
                                str(docker),
                                "image",
                                "ls",
                                "--quiet",
                                "--filter",
                                f"reference={image_tag}",
                            ],
                            env=env,
                            output=evidence / "post-cleanup-inspection.txt",
                        )
                        if post_cleanup.stdout.strip():
                            raise GateError(
                                "run-specific image remained after cleanup"
                            )
                    else:
                        (evidence / "cleanup.log").write_text(
                            "No run-specific image remained.\n",
                            encoding="utf-8",
                        )
                    cleanup_ok = True
                except Exception as error:
                    # Preserve the original failure as well as cleanup failure;
                    # either one independently makes the gate unsuccessful.
                    cleanup_error = GateError(f"image cleanup failed: {error}")
                    if failure is None:
                        failure = cleanup_error
                    else:
                        failure = GateError(f"{failure}; {cleanup_error}")
            else:
                (evidence / "cleanup.log").write_text(
                    "No image build was attempted; no cleanup was needed.\n",
                    encoding="utf-8",
                )
                cleanup_ok = True

    # result.json is the concise verdict.  It is written after cleanup so a
    # consumer never mistakes passed checks plus failed cleanup for success.
    metadata["completedAt"] = dt.datetime.now(dt.UTC).isoformat()
    metadata["cleanupOk"] = cleanup_ok
    metadata["ok"] = failure is None
    if failure is not None:
        metadata["error"] = str(failure)
    (evidence / "result.json").write_text(
        json.dumps(metadata, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    if failure is not None:
        raise GateError(str(failure))
    print(f"container gate complete: {run_id}")
    print(f"input: {input_dir}")
    print(f"evidence: {evidence}")
    return 0


if __name__ == "__main__":
    # Keep expected failures concise for the human operator while preserving a
    # nonzero process status for automation and review records.
    try:
        raise SystemExit(main())
    except (GateError, OSError, UnicodeError, json.JSONDecodeError) as error:
        print(f"container gate failed: {error}", file=sys.stderr)
        raise SystemExit(1) from None
