# ainfra documentation

This directory is the complete Hugo and Hextra documentation project. A fresh
checkout needs Hugo Extended, Git, the repository-level Git metadata, and the
files beneath this directory. Node.js and npm are not required.

## Layout

```text
docs/
├── hugo.yaml
├── assets/
├── content/
├── layouts/
├── static/
├── themes/hextra/
└── scripts/
```

`themes/hextra` is pinned to Hextra `v0.12.3` as a Git submodule. `.cache/`,
`resources/_gen/`, and `public/` are generated locally and ignored. The
FlexSearch `0.8.143` browser bundle is vendored at
`assets/js/vendor/flexsearch.bundle.min.js` so builds do not fetch it from a
CDN. Its SHA-256 digest is
`433e941a8a573ebb9931fc16fc75266ab6b93f569ac2fb4f3dc66882e0416f4c`.

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
Hextra submodule when required.

Preview the exact `gh-pages` change without committing or pushing:

```sh
DOCS_DEPLOY_DRY_RUN=true docs/scripts/deploy-docs.sh
```

To publish an immutable version snapshot:

```sh
DOCS_VERSION=v0.1 docs/scripts/deploy-docs.sh
```
