---
title: Quickstart
weight: 10
description: Validate ainfra and prepare a disposable Hetzner deployment.
---

This guide takes you from a fresh clone to a reviewed disposable-infrastructure
plan. Applying the plan creates billable Hetzner resources, so the final apply
and destroy commands remain explicit.

## Prerequisites

- Python 3.12
- [uv](https://docs.astral.sh/uv/)
- Rust `1.96.1` with Cargo, Clippy, Rustfmt, and `cargo-audit`
- OpenTofu
- Ansible
- Node.js 18 or newer and Hugo Extended for the documentation
- A Hetzner Cloud project token for live operations
- An existing SSH public key

## Install and validate

```sh
git clone https://github.com/projectious-work/ainfra.git
cd ainfra
uv sync --all-groups
scripts/bootstrap-security-tools
scripts/validate-all
scripts/test-all
```

Check local readiness without changing infrastructure:

```sh
uv run ainfra doctor
```

## Initialize a disposable project

Build the Rust binary, create a separate project, and initialize it:

```sh
cargo build
mkdir ../my-infrastructure
cd ../my-infrastructure
../ainfra/target/debug/ainfra init \
  --name my-infrastructure \
  --environment development
../ainfra/target/debug/ainfra validate
```

Commit `ainfra.yaml`, `ainfra.lock`, and the environment input. Keep
`.ainfra/` ignored; it contains operational run state. Edit
`environments/development.yaml` with the location, SSH public key, topology,
and network policy. Export the token; never place it in a project file:

```sh
export HCLOUD_TOKEN='...'
```

## Plan and review

```sh
../ainfra/target/debug/ainfra plan hetzner-kubernetes-baseline \
  --input environments/development.yaml
```

Review the resource count, networking, public-address choices, and ownership
scope. The command returns a plan identifier. Apply requires that exact
identifier:

```sh
../ainfra/target/debug/ainfra apply hetzner-kubernetes-baseline \
  --input environments/development.yaml \
  --approve PLAN_ID
```

{{% alert title="Cost and teardown" color="warning" %}}
Apply creates billable resources. Keep the input, state, ownership scope, and
destroy command available throughout the test. Do not end a disposable test
until Hetzner and the local state both confirm that no managed resources
remain.
{{% /alert %}}

## Read outputs and configure hosts

```sh
../ainfra/target/debug/ainfra outputs hetzner-kubernetes-baseline \
  --run PLAN_ID \
  --format json
../ainfra/target/debug/ainfra configure hetzner-kubernetes-baseline \
  --run PLAN_ID \
  --known-hosts .ainfra/known_hosts
```

Verify SSH host-key fingerprints through the Hetzner console or another
trusted out-of-band channel before the first Ansible connection. Never treat
`ssh-keyscan` as a source of trust.

## Tear down

Create and review a destroy plan:

```sh
../ainfra/target/debug/ainfra plan hetzner-kubernetes-baseline \
  --input environments/development.yaml \
  --destroy
```

Then use the exact destroy-plan ID returned by the lifecycle:

```sh
../ainfra/target/debug/ainfra destroy hetzner-kubernetes-baseline \
  --input environments/development.yaml \
  --approve-destroy PLAN_ID
```

Confirm zero project-owned servers, networks, firewalls, and SSH keys in
Hetzner, then verify that local OpenTofu state contains no resources.

## Next steps

- Review [lifecycle operations]({{< relref "/docs/guides/lifecycle" >}}).
- Understand the [security model]({{< relref
  "/docs/concepts/security-model" >}}).
- Learn how the [Hetzner baseline]({{< relref
  "/docs/reference/hetzner-baseline" >}}) is assembled.
