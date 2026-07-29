---
title: Branching strategy
weight: 10
description: Promote v0 changes through isolated development and release lanes.
---

ainfra uses separate long-lived lanes for stable maintenance and v0
development.

## Long-lived branches

- `main` represents the latest published stable release. Stable release tags
  are merged here after publication. It is the repository's default branch.
- `v0.x-maintenance` carries supported fixes for published v0 releases.
  Features are not backported unless explicitly approved.
- `v0.x-dev` is the normal integration branch for v0 implementation. Feature
  branches and pull requests target this branch.
- `v0.x-pre-release` receives selected promotions from `v0.x-dev` for alpha,
  beta, and release-candidate validation. Prerelease tags are created only
  here.
- `v0.x-release` is promoted from the accepted prerelease state for general
  availability. Stable v0 tags are created here and then merged into `main`.

## Promotion flow

```text
feature branch
    → v0.x-dev
    → v0.x-pre-release
    → v0.x-release
    → main
```

Maintenance fixes begin on `v0.x-maintenance`. Promote a maintenance release
through the same validation expectations, and merge any still-relevant fix
forward into `v0.x-dev`.

## Rules

- Never tag releases from `v0.x-dev`.
- Promote tested commits forward through pull requests; do not develop
  directly on release-integration branches.
- Keep supported maintenance work isolated from new development.
- Use squash-merged pull requests with green local validation.
- Build and verify release artifacts locally before tagging.
- Treat published tags as immutable.
- Keep deployment-only branches, such as `gh-pages`, outside the source
  promotion flow.

## Release verification

Run the complete local gate before entering the release lane:

```sh
./scripts/maintain.sh test
```

Prepare the release version and `release-notes/vX.Y.Z.md` through a pull
request. The release command never edits or commits version metadata. Before
building or tagging, run the non-publishing phase:

```sh
AINFRA_RELEASE_CONFIRM=v0.1.0 \
  ./scripts/maintain.sh release 0.1.0 --steps phase0
```

It writes commit-bound state, doctor, documentation, and checksum reports
under `dist/release-evidence/v0.1.0/COMMIT/`. Resolve every doctor failure
before continuing. Evidence from another commit is not reusable.

Run the independent validation gates with bounded local concurrency:

```sh
AINFRA_RELEASE_CONFIRM=v0.1.0 \
  ./scripts/maintain.sh release 0.1.0 --steps checks
```

Set `AINFRA_RELEASE_JOBS` from `1` through `4` to control concurrency. Each
gate retains a separate log and status. A repeated run reuses evidence only
when the version, commit, Cargo lockfile, Rust toolchain, environment scope,
and every recorded log checksum still match. Delete the candidate evidence
directory to force a complete rerun.

The container-side release builds and verifies both Linux targets before
publishing them:

```sh
AINFRA_RELEASE_CONFIRM=v0.1.0 \
  ./scripts/maintain.sh release 0.1.0
```

On macOS, the host phase uses the standard system `tar`, builds both Darwin
targets, verifies all four local archives and checksum sidecars, uploads the
Darwin assets, and then confirms that the GitHub release contains every
expected archive, checksum, and the installer:

```sh
AINFRA_RELEASE_CONFIRM=v0.1.0 \
  ./scripts/maintain.sh release-host 0.1.0
```

The host phase refuses changes to the tagged binary inputs. Release-tooling
repairs may be newer than an immutable tag only when the Cargo metadata,
lockfile, license, installer, schemas, Rust source, and templates are
byte-for-byte unchanged from that tag.

To recheck a collected four-target artifact set without publishing anything:

```sh
./scripts/maintain.sh audit-release 0.1.0
```

## Live Hetzner release evidence

The live smoke is deliberately separate, local, and opt-in. It fixes the
topology at one `cx23` server with no workers, requires a unique release name,
and retains the recovery project unless both OpenTofu state and independently
queried provider resources are empty.

```sh
export HCLOUD_TOKEN
export AINFRA_HETZNER_E2E_CONFIRM=cost-and-destroy-approved
export AINFRA_HETZNER_E2E_NAME=release-v010-a1b2c3
export AINFRA_SSH_PUBLIC_KEY="ssh-ed25519 ... operator@example.com"
export AINFRA_MANAGEMENT_CIDR="203.0.113.24/32"
export AINFRA_KNOWN_HOSTS="$PWD/known_hosts"
scripts/live-hetzner-release-smoke
```

The operator must verify pricing, private-network reachability, the SSH host
key, and the teardown path before confirming. Sanitized logs and the final
cleanup result are retained under `dist/release-evidence/live-hetzner/`.

## Versioned documentation

The documentation root always represents `main`, the latest published stable
state. The **Releases** menu links to immutable documentation snapshots for
published versions.

Publish the current stable documentation:

```sh
docs/scripts/deploy-docs.sh
```

When publishing a release, first add its version and URL to `params.versions`
in `docs/hugo.yaml`. Build the accepted release commit or tag and publish its
snapshot under the matching path:

```sh
DOCS_VERSION=v0.1 docs/scripts/deploy-docs.sh
```

That command preserves the root site and replaces only `/v0.1/`. Versioned
pages identify themselves as archived snapshots and link readers back to the
latest documentation. Once published, a versioned snapshot should be treated
as immutable except for an explicitly approved documentation correction.
