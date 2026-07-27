---
title: Local validation
weight: 20
description: Install and run the repository's local quality and security gates.
---

The repository intentionally runs its gates locally. It does not contain
GitHub Actions workflow files.

## Install security tools

The supported bootstrap installs pinned tools into the ignored
`.ainfra/tools/` directory:

```sh
scripts/bootstrap-security-tools
```

It installs:

- Checkov `3.2.529` in an isolated uv tool environment;
- Gitleaks `8.30.1`, verified against its release checksum.

Checkov is isolated because its dependency constraints conflict with the
project environment. Runtime version checks fail closed.

## Run all gates

```sh
scripts/validate-all
scripts/test-all
```

The validation suite covers formatting, typing, contracts, OpenTofu,
Ansible, repository policy, Checkov, and Gitleaks. Missing required tooling is
a failure, not a skipped check.

## Build the documentation

```sh
docs/scripts/build-docs.sh
docs/scripts/serve-docs.sh
```

The local server uses
`http://localhost:1313/ainfra-templates/` so relative links behave like the
GitHub Pages deployment.
