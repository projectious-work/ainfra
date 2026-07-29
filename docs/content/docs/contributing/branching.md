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

To recheck a collected four-target artifact set without publishing anything:

```sh
./scripts/maintain.sh audit-release 0.1.0
```

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
