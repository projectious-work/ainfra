# Contributing to ainfra

## Prerequisites

- Go, using the version selected by the repository once implementation begins.
- Hugo Extended, compatible with the pinned Hextra module.
- OpenTofu and Ansible for engine integration work.

## Build

Build the CLI and all four supported release targets with:

```bash
go build ./cmd/ainfra
scripts/build-targets.sh
```

Build the end-user documentation locally with:

```bash
docs/scripts/build.sh
```

## Test

Validate the normative v1 specification with:

```bash
uv run scripts/validate-v1-spec
```

Documentation changes must also pass a clean production build.

## Phase 1 host container gate

The restricted development container cross-builds all four supported targets
and prepares—but never executes—the host-validation input:

```bash
scripts/maintain.sh release-host-prepare --version=1.0.0-alpha.1
```

The command prints a unique ignored directory below `tmp/container-gate/`.
Review its immutable `input/`, the current source diff, and
[`scripts/container-gate-host`](scripts/container-gate-host). A human operator
may then run the jointly reviewed launcher and Python entrypoint on a host with
uv, Docker, Syft, and Grype.

The gate pins Python 3.13.14. The launcher asks uv to install or reuse that
exact managed runtime and creates a private virtual environment inside the
prepared run. No manual Python path or digest approval is required. The first
run may access the network to acquire Python; later runs reuse uv's cache.
On the host, select the newest unused run for exactly that version:

```bash
scripts/maintain.sh release-host --version=1.0.0-alpha.1 --dry-run
```

The host gate always creates evidence only; `--dry-run` makes that
non-publishing behavior explicit. The version-scoped handover removes any need
to copy or type the generated timestamp/random run identifier.

The launcher creates `runtime/bootstrap/`, records the selected uv and Python
paths, versions, and executable digests, and then directly executes the Python
entrypoint. On macOS, one
installation possibility is:

```bash
brew install uv syft grype
brew install --cask docker
```

Start Docker Desktop, install its CLI tools into `/usr/local/bin`, and enable
the default Docker socket. The gate checks all required executables before it
creates authoritative `evidence/` and prints OS-specific installation
guidance when a container tool is missing. uv's managed Python and download
cache use the fixed per-user cache documented by the release specification;
all other per-run state remains below `runtime/`.

On Linux, install Docker Engine from the distribution or Docker's official
repository, and install Syft and Grype into `/usr/local/bin`, `/usr/bin`, or
the fixed Linuxbrew prefix. Both conventional and rootless Docker daemons are
supported because the gate invokes only the Docker CLI and does not select or
mount a socket itself.

Go is required only inside the development container. The host consumes the
prepared Linux and macOS binaries and does not compile release artifacts.

Do not give an agent container-runtime authority. Do not alter the host script
after owner review without new, explicit owner permission. The host script
creates `evidence/`, fails rather than overwriting an existing run, removes
only its unique temporary image, and retains its result for review.

## Before committing

Run the relevant local build, test, formatting, linting, and security checks.
All repository validation and publication runs locally. Do not introduce
GitHub Actions or another hosted CI or deployment path.

## Commit message format

Use Conventional Commits such as `feat:`, `fix:`, `docs:`, and `chore:`.
Reference the relevant GitHub issue or pull request when one exists.

## Branches and releases

Topic branches merge into `v1.x-dev` through review. Promotion to
`v1.x-pre-release`, `v1.x-release`, and `main` is fast-forward-only. Read
[`spec/doc/v1/10-release-engineering.md`](spec/doc/v1/10-release-engineering.md)
before changing a long-lived branch or preparing a release.

## Repository layout

| Path | Owns |
|---|---|
| `docs/` | Hugo/Hextra end-user documentation |
| `spec/` | Normative v1 specification, examples, schemas, and fixtures |
| `scripts/` | Local validation and release tooling |
| `context/` | processkit-managed project context |
