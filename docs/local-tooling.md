# Local security tooling

The repository never uses GitHub Actions or workflow files. Install these
exact tool versions locally before running `scripts/validate-all`:

- Checkov `3.2.529`, installed in an isolated uv tool environment
- Gitleaks `8.30.1`, verified against its published release checksum

The approved exact-tool bootstrap performs both operations on supported Linux
and macOS architectures:

```sh
scripts/bootstrap-security-tools
```

This is a required setup prerequisite because Checkov `3.2.529` pins
`packaging<24` and therefore cannot share the project environment, which pins
`packaging==26.2`. The bootstrap installs into the gitignored
`.ainfra/tools/` directory. Both runtime version checks fail closed.

The Gitleaks configuration excludes only generated/private harness state,
language caches, dependencies, and the read-only upstream ProcessKit mirror.
Repository source, tests, fixtures, scripts, schemas, and template
implementation remain in scope.
