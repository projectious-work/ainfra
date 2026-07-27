---
title: Contributing
weight: 60
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

`main` contains the latest published stable release. Normal implementation
work integrates through `v0.x-dev`; feature branches start from and merge back
to that branch.

See the [branching strategy](branching/) for the maintenance, development,
prerelease, and stable-release promotion lanes.

## Documentation

Build the site before opening a documentation change:

```sh
docs/scripts/build-docs.sh
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
