# Contributing to ainfra

## Prerequisites

- Go, using the version selected by the repository once implementation begins.
- Hugo Extended, compatible with the pinned Hextra module.
- OpenTofu and Ansible for engine integration work.

## Build

The Go implementation has not been bootstrapped yet. Build commands will be
added with the first executable package.

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
