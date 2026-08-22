# Contributing


Contributions to ainfra should begin with the repository's contribution and
security guidance.

The complete development workflow is maintained in the repository
[`CONTRIBUTING.md`](https://github.com/projectious-work/ainfra/blob/v1.x-dev/CONTRIBUTING.md).

The release command boundary is strict. All release lanes must first point to
one exact commit; `release-freeze` records that commit and tree before any
artifact production:

> Do not paste the complete sequence into one terminal. **HOST** commands run
> in the normal host checkout; **DEVCONTAINER** commands run in that checkout's
> ainfra devcontainer. Both environments see the same repository files. No
> second clone, checkout, worktree, or artifact copy is needed, and Go is never
> installed or invoked on the host.

1. **HOST:** run `release-freeze` after all merges and lane promotions.
2. **DEVCONTAINER:** run `release-package`. This is the only step that uses Go.
   It cross-builds four archives and uses Syft to create their SBOMs. A missing
   `go` error means the command was run on the host; return to the devcontainer
   instead of installing Go on the host.
3. **HOST:** run `release-host-prepare`. It needs Git and Python, verifies the
   packaged artifacts, and does not compile anything.
4. **HOST:** run `release-host`. It uses uv, Docker, Syft, and Grype to test and
   inventory the independently built container image.
5. **HOST:** run `release-publish --dry-run` to preflight credentials and
   publication.
6. **HOST:** run `release-sign`, followed by resumable `release-publish`.

Syft therefore belongs in both environments, but it inventories different
subjects. The workspace pins an aibox release whose catalog supplies the
complete `supply-chain` toolset: Gitleaks, OSV-Scanner, Syft, Grype, and Cosign.
Run `aibox apply` and rebuild the devcontainer after changing that pin or its
tool selections. Agents must never receive the host Docker socket or
equivalent container-runtime authority. See the repository guide for the
complete commands, prerequisites, and checks.


---
Source: https://projectious-work.github.io/ainfra/docs/contributing/index.md
