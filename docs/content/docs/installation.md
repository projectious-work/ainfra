---
title: Installation
weight: 30
---

ainfra currently builds from source for Linux and macOS on `amd64` and
`arm64`. Install Go 1.26.5, clone the repository, and run:

```sh
go build -o ainfra ./cmd/ainfra
./ainfra version
```

OpenTofu, Ansible, Git, and a compatible container runtime are expected to be
the lifecycle prerequisites. Their supported version ranges will be published
when the corresponding lifecycle phases ship.

The repository also contains a convenience Dockerfile; no prebuilt image is
published. Build the Linux binaries and local image with:

```sh
scripts/build-targets.sh dist
docker build --build-arg TARGETARCH=amd64 -t ainfra:local .
docker run --rm ainfra:local version
```
