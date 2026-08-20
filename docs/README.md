# ainfra documentation site

This directory contains the public ainfra end-user documentation. It uses
[Hugo](https://gohugo.io/) with the pinned
[projectious.work brand theme](https://github.com/projectious-work/brand-theme-hugo-vanilla)
and is published to GitHub Pages.

The site uses the theme's Tailwind pipeline. Install its pinned build
dependencies once:

```bash
cd docs
npm install
```

## Local development

```bash
docs/scripts/serve.sh
```

The watcher listens on port 1314 and serves the project below `/ainfra/`.

## Build

```bash
docs/scripts/build.sh
```

The static site is generated in `docs/public/`. The build runs twice so the
second Tailwind pass consumes Hugo's complete class inventory.

## Deployment

```bash
docs/scripts/deploy.sh
```

Deployment builds locally and pushes the generated site directly to the
`gh-pages` branch. This repository does not use GitHub Actions.

Project-specific theme integration gaps are recorded in
[`theme-gap-ledger.md`](theme-gap-ledger.md) and reported upstream only after
they are reproduced against the pinned final theme release.
