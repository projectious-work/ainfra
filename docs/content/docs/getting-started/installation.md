---
title: Installation
weight: 5
description: Install a verified ainfra release without root access.
---

The supported installer downloads only from the canonical GitHub release,
verifies the release checksum before extraction, and installs to
`$HOME/.local/bin` by default. It never invokes `sudo`.

```bash
curl --proto '=https' --tlsv1.2 --fail --location \
  --proto-redir '=https' \
  https://github.com/projectious-work/ainfra/releases/latest/download/install.sh \
  -o /tmp/ainfra-install.sh
sh /tmp/ainfra-install.sh
```

Review a downloaded installer before running it when that is required by your
local security policy.

To install a specific stable version:

```bash
AINFRA_VERSION=0.1.0 sh /tmp/ainfra-install.sh
```

To select another unprivileged destination:

```bash
AINFRA_INSTALL_DIR="$HOME/bin" sh /tmp/ainfra-install.sh
```

The installer supports these release targets:

| Operating system | Architecture | Release target |
|---|---|---|
| Linux | x86_64 | `x86_64-unknown-linux-gnu` |
| Linux | ARM64 | `aarch64-unknown-linux-gnu` |
| macOS | Intel | `x86_64-apple-darwin` |
| macOS | Apple Silicon | `aarch64-apple-darwin` |

After installation, the script runs both `ainfra --version` and
`ainfra --help`. If the destination is not already on `PATH`, it prints the
directory that must be added.

For development from a source checkout, use the pinned Rust toolchain:

```bash
cargo build
./target/debug/ainfra --version
```
