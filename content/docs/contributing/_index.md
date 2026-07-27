---
title: Contributing
weight: 50
description: Develop, test, and document changes locally.
---

Issues and pull requests are welcome. Keep changes narrow, preserve the
security boundaries, and include evidence proportional to the risk.

## Development setup

```sh
uv sync --all-groups
scripts/bootstrap-security-tools
scripts/validate-all
scripts/test-all
```

Python and Markdown lines are hard-wrapped at 80 columns. Use Conventional
Commits and never bypass hooks.

## Branches

`main` contains releases. Minor-release work integrates through a
version-specific branch such as `v0.1-dev`; feature branches start from and
merge back to that branch. Release commits receive annotated semantic-version
tags.

## Documentation

Build the site before opening a documentation change:

```sh
scripts/build-docs.sh
```

Write task-oriented procedures as guides, stable facts as reference, and
design rationale in concepts. Update the README only as the concise front door;
the documentation site remains the detailed source.

## Infrastructure changes

Infrastructure changes should include:

- strict contract or policy coverage;
- positive and negative tests;
- exact dependency pins;
- direct-tool equivalents;
- redacted evidence;
- a teardown plan before any live apply.

Live cost-bearing tests require separate explicit approval.
