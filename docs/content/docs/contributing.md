---
title: Contributing
weight: 90
---

Contributions to ainfra should begin with the repository's contribution and
security guidance.

The complete development workflow is maintained in the repository
[`CONTRIBUTING.md`](https://github.com/projectious-work/ainfra/blob/v1.x-dev/CONTRIBUTING.md).

Container validation consumes archives built and checksummed in the restricted
development environment. Host preparation uses Git and Python to verify and
unpack those artifacts into an immutable input; it does not require Go or
compile binaries. A jointly reviewed launcher and Python entrypoint then run
through uv on a human-operated macOS or Linux Docker-capable host. The
entrypoint checks Docker, Syft, and Grype before creating authoritative
evidence. Agents must never receive the host Docker socket or equivalent
container-runtime authority. See the repository guide for the complete flow.
