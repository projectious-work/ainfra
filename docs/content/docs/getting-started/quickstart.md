---
title: Quickstart
weight: 10
description: Validate ainfra and prepare a disposable Hetzner deployment.
---

This guide takes you from an installed binary to a reviewed disposable-
infrastructure plan. Applying the plan creates billable Hetzner resources, so
the final apply and destroy commands remain explicit.

## Prerequisites

- A verified `ainfra` release from the [installation guide]({{< relref
  "/docs/getting-started/installation" >}})
- OpenTofu
- Ansible
- A Hetzner Cloud project token for live operations
- An existing SSH public key

## Install and validate

Check local readiness without changing infrastructure:

```sh
ainfra --version
```

## Initialize a disposable project

Create a separate project and initialize it:

```sh
mkdir ../my-infrastructure
cd ../my-infrastructure
ainfra init \
  --name my-infrastructure \
  --environment development
ainfra validate
ainfra doctor --environment development
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
ainfra plan \
  --environment development
```

Review the resource count, networking, public-address choices, and ownership
scope. The command returns a plan identifier. Apply requires that exact
identifier:

```sh
ainfra apply \
  --environment development \
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
ainfra outputs \
  --environment development \
  --run PLAN_ID \
  --format json
ainfra configure \
  --environment development \
  --run PLAN_ID \
  --known-hosts .ainfra/known_hosts
ainfra status --environment development
```

Verify SSH host-key fingerprints through the Hetzner console or another
trusted out-of-band channel before the first Ansible connection. Never treat
`ssh-keyscan` as a source of trust.

## Tear down

Create and review a destroy plan:

```sh
ainfra plan \
  --environment development \
  --destroy
```

Then use the exact destroy-plan ID returned by the lifecycle:

```sh
ainfra down \
  --environment development \
  --approve-destroy PLAN_ID
ainfra status --environment development
```

Confirm zero project-owned servers, networks, firewalls, and SSH keys in
Hetzner, then verify that local OpenTofu state contains no resources.

## Next steps

- Review [lifecycle operations]({{< relref "/docs/guides/lifecycle" >}}).
- Understand the [security model]({{< relref
  "/docs/concepts/security-model" >}}).
- Learn how the [Hetzner baseline]({{< relref
  "/docs/reference/hetzner-baseline" >}}) is assembled.
