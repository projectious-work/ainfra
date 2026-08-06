# ainfra Documentation Site

This directory contains the public ainfra end-user documentation. It uses
[Hugo](https://gohugo.io/) with the pinned
[Hextra](https://imfing.github.io/hextra/) module and is published to GitHub
Pages.

## Local development

```bash
docs/scripts/serve.sh
```

## Build

```bash
docs/scripts/build.sh
```

The static site is generated in `docs/public/`.

## Deployment

```bash
docs/scripts/deploy.sh
```

Deployment builds locally and pushes the generated site directly to the
`gh-pages` branch. This repository does not use GitHub Actions.
