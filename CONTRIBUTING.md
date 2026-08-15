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

## Release environments

The release deliberately crosses one trust boundary. Run commands in this
order and in the environment shown:

| Order | Environment | Command | Required tools | Result |
|---|---|---|---|---|
| 1 | Devcontainer | `release-package` | Git, Go, tar, gzip, Syft, sha256sum | Four cross-built archives, four archive SBOMs, and checksums |
| 2 | Host | `release-host-prepare` | Git, Python | Immutable, checksum-verified gate input |
| 3 | Host | `release-host` | uv, Docker, Syft, Grype | Container smoke-test, image SBOM, vulnerability report, and evidence |
| 4 | Host | `release-sign` | Cosign | Signed and verified checksum manifest |
| 5 | Host | `release-publish` | Git, GitHub CLI, Cosign | Tag, prerelease assets, and independent verification |

`release-package` is the only command that compiles Go. Do not run it on the
host. Syft is required in both environments for different subjects: the
devcontainer inventories the publishable archives, while the host inventories
the independently built container image. The host never compiles or modifies
the release archives.

### 1. Package in the devcontainer

The restricted devcontainer packages all four supported targets, including
their SBOMs and checksum manifest:

```bash
scripts/maintain.sh release-package --version=1.0.0-alpha.2
```

If the current aibox catalog reports the `supply-chain` addon as unknown, the
repository's `.devcontainer/Dockerfile.local` compatibility layer installs
Syft. Rebuild the devcontainer after `aibox apply`, then verify `go version`
and `syft version` before packaging.

### 2. Prepare and run the gate on the host

After packaging, leave the devcontainer. On the host, preparation verifies the
archives and copies their binaries into an immutable validation input:

```bash
scripts/maintain.sh release-host-prepare --version=1.0.0-alpha.2
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
scripts/maintain.sh release-host --version=1.0.0-alpha.2 --dry-run
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

Go is required only inside the devcontainer. Host preparation requires Git and
Python, consumes the checksum-verified release archives, and never compiles
release artifacts.

## Release checksum signing

Release packaging runs inside the devcontainer and writes the four
publishable archives, their SPDX JSON SBOMs, and `checksums.sha256` below
`dist/release/<version>/`:

```bash
scripts/maintain.sh release-package --version=1.0.0-alpha.2
```

The command requires a clean worktree, refuses to overwrite an earlier
release directory, injects the release version, uses the source commit time for
reproducible archives, and verifies the completed checksum manifest. On the
host, install Cosign once:

```bash
brew install cosign
```

Before signing, validate the prepared directory without opening a browser or
creating a public transparency-log entry:

```bash
scripts/maintain.sh release-sign --version=1.0.0-alpha.1 --dry-run
```

For the release, run the same command without `--dry-run`:

```bash
scripts/maintain.sh release-sign --version=1.0.0-alpha.1
```

Cosign opens Sigstore's login flow. Select GitHub and authenticate as the
`projectious` user. The script signs `checksums.sha256`, writes
`checksums.sha256.sigstore.json`, and immediately verifies the bundle against
the fixed identity `info@projectious.work` and OIDC issuer
`https://github.com/login/oauth`. The browser authorization flow starts at
`https://oauth2.sigstore.dev/auth`, but that endpoint is not the issuer claim
in the Fulcio certificate. `GH_TOKEN` is not used for signing; it
remains available to the later GitHub release publication stage.

GitHub email configuration and Cosign installation are one-time setup. The
interactive approval, short-lived Fulcio certificate, Rekor entry, signature,
and verification are intentionally repeated for every release.

After the exact release commit has been promoted to `v1.x-release`, validate
publication without changing GitHub:

```bash
scripts/maintain.sh release-publish --version=1.0.0-alpha.1 --dry-run
```

The real command creates the immutable annotated tag, uploads the ten prepared
assets as a GitHub prerelease, downloads them into a private temporary
directory, and independently re-verifies checksums and the Sigstore identity:

```bash
scripts/maintain.sh release-publish --version=1.0.0-alpha.1
```

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
