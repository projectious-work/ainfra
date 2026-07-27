# ainfra documentation

This directory is the complete Hugo and Docsy documentation project. A fresh
checkout needs only the repository-level Git metadata plus the files beneath
this directory to install documentation dependencies and build the site.

## Layout

```text
docs/
├── hugo.yaml
├── package.json
├── package-lock.json
├── assets/
├── content/
├── layouts/
├── static/
├── themes/docsy/
└── scripts/
```

`themes/docsy` is a pinned Git submodule. `node_modules/`, `.cache/`,
`resources/_gen/`, and `public/` are generated locally and ignored.

## Commands

Run commands from the repository root:

```sh
docs/scripts/build-docs.sh
docs/scripts/test-docs.sh
docs/scripts/serve-docs.sh
docs/scripts/deploy-docs.sh
```

The scripts resolve `docs/` from their own location, so they work regardless
of the caller's current directory. The build and serve commands initialize the
Docsy submodule and run `npm ci` inside `docs/` when required.

To publish an immutable version snapshot:

```sh
DOCS_VERSION=v0.1 docs/scripts/deploy-docs.sh
```
