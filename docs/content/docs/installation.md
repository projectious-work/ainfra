---
title: Installation
weight: 30
---

ainfra publishes signed archives for Linux and macOS on `amd64` and `arm64`.
Download the archive, `checksums.sha256`, and
`checksums.sha256.sigstore.json` from the latest GitHub release, then verify
the checksum manifest as described in its release notes before installing the
binary on your `PATH`.

To build from source instead, install Go 1.26.5, clone the repository, and run:

```sh
go build -o ainfra ./cmd/ainfra
./ainfra version
```

Git is required only when locking or updating a Git template source. OpenTofu,
Ansible, and a compatible container runtime are expected to become lifecycle
prerequisites as their corresponding phases ship; they are not used by the
current immutable-source workflow.

The repository also contains a convenience Dockerfile; no prebuilt image is
published. Build the Linux binaries and local image with:

```sh
scripts/build-targets.sh dist
docker build --build-arg TARGETARCH=amd64 -t ainfra:local .
docker run --rm ainfra:local version
```
