# Installation


ainfra publishes archives for Linux and macOS on `amd64` and `arm64`. The
official installer detects the current platform, downloads the matching
release, verifies its published checksum, and installs the binary:

```sh
curl -fsSL https://raw.githubusercontent.com/projectious-work/ainfra/v1.x-release/scripts/install.sh | bash
```

The default destination is `~/.local/bin`. Ensure that directory is on
`PATH`, or select a version and destination explicitly:

```sh
curl -fsSL https://raw.githubusercontent.com/projectious-work/ainfra/v1.x-release/scripts/install.sh \
  | VERSION=1.0.0-alpha.7 INSTALL_DIR=/usr/local/bin bash
```

Every release includes `checksums.sha256` and its
`checksums.sha256.sigstore.json` signature bundle. Checksum verification is
mandatory. When Cosign is available, the installer also verifies that the
manifest was signed by `info@projectious.work` through GitHub's OIDC issuer.
Require that identity check in controlled environments with:

```sh
curl -fsSL https://raw.githubusercontent.com/projectious-work/ainfra/v1.x-release/scripts/install.sh \
  | VERIFY_SIGNATURE=1 bash
```

Alternatively, download the archive, checksum manifest, and signature bundle
from the [release page](https://github.com/projectious-work/ainfra/releases)
and verify them before placing `ainfra` on `PATH`.

To build from source instead, install Go 1.26.5, clone the repository, and run:

```sh
go build -o ainfra ./cmd/ainfra
./ainfra version
```

Git is required when locking or updating a Git template source. OpenTofu is
required for planning and infrastructure changes; Ansible is required for
configuration and convergence operations. Run `ainfra doctor environment`
after installation to see which tools the intended workflow needs.

The repository also contains a convenience Dockerfile; no prebuilt image is
published. Build the Linux binaries and local image with:

```sh
scripts/build-targets.sh dist
docker build --build-arg TARGETARCH=amd64 -t ainfra:local .
docker run --rm ainfra:local version
```


---
Source: https://projectious-work.github.io/ainfra/docs/installation/index.md
