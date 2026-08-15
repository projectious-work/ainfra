---
title: Contributing
weight: 90
---

Contributions to ainfra should begin with the repository's contribution and
security guidance.

The complete development workflow is maintained in the repository
[`CONTRIBUTING.md`](https://github.com/projectious-work/ainfra/blob/v1.x-dev/CONTRIBUTING.md).

The release command boundary is strict:

1. In the devcontainer, run `release-package`. This is the only step that uses
   Go. It cross-builds four archives and uses Syft to create their SBOMs.
2. On the host, run `release-host-prepare`. It needs Git and Python, verifies
   the packaged artifacts, and does not compile anything.
3. On the host, run `release-host`. It uses uv, Docker, Syft, and Grype to test
   and inventory the independently built container image.
4. On the host, run `release-sign`, followed by `release-publish`.

Syft therefore belongs in both environments, but it inventories different
subjects. The workspace pins an aibox release whose catalog supplies the
complete `supply-chain` toolset: Gitleaks, OSV-Scanner, Syft, Grype, and Cosign.
Run `aibox apply` and rebuild the devcontainer after changing that pin or its
tool selections. Agents must never receive the host Docker socket or
equivalent container-runtime authority. See the repository guide for the
complete commands, prerequisites, and checks.
