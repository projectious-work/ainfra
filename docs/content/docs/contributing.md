---
title: Contributing
weight: 90
---

Contributions to ainfra should begin with the repository's contribution and
security guidance.

The complete development workflow is maintained in the repository
[`CONTRIBUTING.md`](https://github.com/projectious-work/ainfra/blob/v1.x-dev/CONTRIBUTING.md).

Phase 1 container validation uses an immutable input directory prepared in the
restricted development environment and a jointly reviewed launcher and Python
entrypoint executed through uv by a human on a macOS or Linux Docker-capable
host. The launcher verifies an exact owner-approved Python runtime; the
entrypoint checks Docker, Syft, and Grype before creating authoritative
evidence. Agents must never receive the host Docker socket or equivalent
container-runtime authority. See the repository guide for setup, preparation,
review, execution, and retained-evidence steps.
