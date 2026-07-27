---
title: Create and tear down a Hetzner environment
linkTitle: Hetzner environment
weight: 10
description: Configure, plan, apply, inspect, and destroy the baseline template.
---

This tutorial creates a disposable Hetzner Cloud environment with the
`hetzner-kubernetes-baseline` template and the `ainfra` CLI. It provisions one
Debian 13 control-plane-capable host, a private network, firewall rules, and
the operator SSH key registration. It then destroys every managed resource.

{{% alert title="Cost and responsibility" color="warning" %}}
The apply step creates billable Hetzner resources. Check current pricing and
quotas in your Hetzner project before continuing. Keep the input, local state,
plan records, token, and destroy procedure available until independent checks
confirm that no managed resources remain.
{{% /alert %}}

## What this procedure does

```text
Hetzner project + token + SSH public key
    → non-secret TemplateInput
    → validate and check readiness
    → create and review an OpenTofu plan
    → approve that exact plan with ainfra
    → inspect the resulting environment
    → create and review a destroy plan
    → approve that exact destroy plan
    → verify zero remaining resources
```

`ainfra apply` runs the reviewed OpenTofu plan. It does not automatically run
the template's Ansible playbook or install Kubernetes. Host configuration
beyond cloud-init is a separate, explicit operation.

## 1. Prepare the workstation

Install the prerequisites:

- Git;
- Python 3.12 and `uv`;
- the pinned OpenTofu version from `tools.lock`;
- Ansible and the local security tools used by repository validation;
- an Ed25519 SSH key pair.

Clone and prepare the repository:

```sh
git clone --recurse-submodules \
  https://github.com/projectious-work/ainfra-templates.git
cd ainfra-templates
uv sync --all-groups
scripts/bootstrap-security-tools
scripts/validate-all
scripts/test-all
```

Check the CLI:

```sh
uv run ainfra --version
uv run ainfra doctor
```

Do not proceed while `doctor` reports a failed required dependency.

## 2. Create a Hetzner project token

Open the intended project in
[Hetzner Console](https://console.hetzner.cloud/). Under **Security → API
tokens**, generate a project-specific **Read & Write** token. Hetzner displays
the full token only once, so transfer it directly to an approved secret
manager.

The token is bound to the selected Hetzner project. A read-only token cannot
create or destroy the resources in this procedure. See Hetzner's official
[API-token instructions](https://docs.hetzner.com/cloud/api/getting-started/generating-api-token/).

Load the token into the current shell without writing it into the repository
or input document:

```sh
read -rsp "Hetzner API token: " HCLOUD_TOKEN
printf '\n'
export HCLOUD_TOKEN
```

The example input refers to the environment variable by name:

```yaml
projectTokenRef:
  type: environment
  name: HCLOUD_TOKEN
```

## 3. Prepare an SSH public key

Use an existing Ed25519 key or create one:

```sh
ssh-keygen -t ed25519 -f ~/.ssh/ainfra-hetzner \
  -C "ainfra disposable Hetzner environment"
```

Display the public half:

```sh
cat ~/.ssh/ainfra-hetzner.pub
```

Only the single-line `.pub` value belongs in the input. Never copy the private
key into the repository, YAML document, state configuration, or plan record.

## 4. Create the environment input

Create a private working directory and copy the non-secret example:

```sh
mkdir -p .ainfra
cp templates/hetzner-kubernetes-baseline/inputs/example.input.yaml \
  .ainfra/hetzner-tutorial.input.yaml
chmod 600 .ainfra/hetzner-tutorial.input.yaml
```

Edit `.ainfra/hetzner-tutorial.input.yaml`:

```yaml
apiVersion: ainfra.projectious.work/v1alpha1
kind: TemplateInput
metadata:
  template: hetzner-kubernetes-baseline
  environment: tutorial
  disposable: true
spec:
  provider:
    location: fsn1
    projectTokenRef:
      type: environment
      name: HCLOUD_TOKEN
  topology:
    controlPlaneCount: 1
    workerCount: 0
    image: debian-13
    serverType: cx23
  network:
    privateCidr: 10.42.0.0/16
    publicIPv4: true
    publicIPv6: false
    managementIngressCidrs:
      - 203.0.113.24/32
    workloadIngressCidrs: []
  access:
    adminSshPublicKeys:
      - ssh-ed25519 REPLACE_WITH_YOUR_PUBLIC_KEY operator@example.com
  state:
    mode: local-disposable
```

Replace both placeholders. `203.0.113.24/32` must become the trusted public
IPv4 egress address of the operator workstation or VPN. Do not use
`0.0.0.0/0`; policy requires IPv4 management ranges of `/24` or narrower, and
this tutorial deliberately uses a single `/32`.

### Input-field choices

| Field | Configure it as follows |
|---|---|
| `metadata.environment` | A unique lowercase environment name. It becomes part of resource names and ownership labels. |
| `metadata.disposable` | Keep `true` when using `local-disposable` state. |
| `provider.location` | Choose `fsn1`, `nbg1`, or `hel1`, the European locations supported by this contract. |
| `projectTokenRef` | Keep the `HCLOUD_TOKEN` environment reference; never paste the token value. |
| `controlPlaneCount` | Exactly `1` in the current contract. |
| `workerCount` | `0` for the smallest tutorial, or the required non-negative count. |
| `image` | `debian-13`, currently the only validated image. |
| `serverType` | `cx23` in the example; confirm availability and price in the selected location. |
| `privateCidr` | A private IPv4 range that does not overlap networks used to reach the environment. |
| `publicIPv4` | `true` for direct tutorial access; this can incur a separate charge. |
| `publicIPv6` | Keep `false` unless IPv6 access is explicitly required and tested. |
| `managementIngressCidrs` | Only trusted operator or VPN egress CIDRs. These permit TCP/22 when a public address is enabled. |
| `workloadIngressCidrs` | Keep empty; the baseline does not install a workload. |
| `adminSshPublicKeys` | One or more Ed25519 public keys, never private keys. |
| `state.mode` | `local-disposable` only for an environment that will be torn down in this session. |

For a private-only environment, set both public-address values to `false` and
leave both ingress lists empty. You must then already have a trusted route into
the Hetzner private network; ainfra does not create a VPN or bastion.

## 5. Validate input and readiness

Validate the manifest and input together:

```sh
INPUT=.ainfra/hetzner-tutorial.input.yaml
TEMPLATE=hetzner-kubernetes-baseline

uv run ainfra validate "$TEMPLATE" --input "$INPUT"
uv run ainfra doctor --input "$INPUT"
```

These commands do not create infrastructure. Resolve every error before
planning.

Keep the same input path and contents for plan, apply, destroy planning, and
destroy. ainfra binds the reviewed plan to the exact template, version,
environment, input path, input digest, template digest, and plan bytes.

## 6. Create the apply plan

Create a machine-readable plan record:

```sh
uv run ainfra plan "$TEMPLATE" \
  --input "$INPUT" \
  --format json | tee .ainfra/apply-plan.json
```

Extract its exact identifier and plan path without adding another dependency:

```sh
APPLY_PLAN_ID="$(
  uv run python -c \
    'import json; print(json.load(open(".ainfra/apply-plan.json"))["id"])'
)"
APPLY_PLAN_PATH="$(
  uv run python -c \
    'import json; print(json.load(open(".ainfra/apply-plan.json"))["plan_path"])'
)"
printf 'Apply plan ID: %s\n' "$APPLY_PLAN_ID"
```

## 7. Review the plan

Render the exact saved OpenTofu plan:

```sh
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu \
  show "$APPLY_PLAN_PATH"
```

Before approval, verify:

- exactly one server and zero workers for this tutorial;
- the expected `cx23` server type and `debian-13` image;
- the selected location;
- one private network and subnet using the chosen CIDR;
- public IPv4 enabled and public IPv6 disabled;
- TCP/22 allowed only from the trusted management `/32`;
- no workload ingress;
- ownership labels containing `managed-by=ainfra`,
  `template=hetzner-kubernetes-baseline`, and `environment=tutorial`;
- registrations for only the intended SSH public keys;
- the resource count and current estimated Hetzner cost are acceptable.

Do not edit the input or template after planning. If anything is wrong, discard
the plan and create a new one after correcting the input.

## 8. Apply the reviewed plan

Apply only the exact plan ID printed above:

```sh
uv run ainfra apply "$TEMPLATE" \
  --input "$INPUT" \
  --approve "$APPLY_PLAN_ID"
```

Do not interrupt the command unless continuing would be more dangerous. If it
fails, inspect both the Hetzner project and local state before retrying or
destroying.

## 9. Confirm the environment

Inspect the local OpenTofu state:

```sh
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu state list
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu \
  output -json inventory_nodes
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu \
  output -json ownership
```

In Hetzner Console, confirm the expected server, network, firewall, and SSH-key
registrations. Their names begin with `ainfra-tutorial-`, and their labels
identify the template and environment.

Before the first SSH connection, obtain the server host-key fingerprint
through the Hetzner console or another trusted out-of-band channel. Add the
verified key to `known_hosts`; do not use `ssh-keyscan` as the source of trust.
Cloud-init creates the `ainfra` user and installs its authorized public key.

{{% alert title="Ansible is separate" color="info" %}}
The infrastructure is now up, but the CLI has not run Ansible. The baseline's
Ansible playbook is intentionally explicit. Its generated inventory uses
private node addresses, so run it only from a trusted host with private-network
reachability and after host-key verification. See the
[Hetzner baseline reference]({{< relref
"/docs/reference/hetzner-baseline" >}}).
{{% /alert %}}

## 10. Create and review the destroy plan

Do not reuse the apply plan. Create a dedicated destroy plan from the unchanged
template and input:

```sh
uv run ainfra plan "$TEMPLATE" \
  --input "$INPUT" \
  --destroy \
  --format json | tee .ainfra/destroy-plan.json
```

Extract the destroy approval ID and plan path:

```sh
DESTROY_PLAN_ID="$(
  uv run python -c \
    'import json; print(json.load(open(".ainfra/destroy-plan.json"))["id"])'
)"
DESTROY_PLAN_PATH="$(
  uv run python -c \
    'import json; print(json.load(open(".ainfra/destroy-plan.json"))["plan_path"])'
)"
printf 'Destroy plan ID: %s\n' "$DESTROY_PLAN_ID"
```

Review it:

```sh
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu \
  show "$DESTROY_PLAN_PATH"
```

Confirm that it removes every resource owned by the tutorial environment and
does not affect unrelated Hetzner resources.

## 11. Destroy the environment

The exact destroy-plan ID is the scope-bound approval token expected by
`--approve-destroy`:

```sh
uv run ainfra destroy "$TEMPLATE" \
  --input "$INPUT" \
  --approve-destroy "$DESTROY_PLAN_ID"
```

## 12. Verify complete teardown

The state list must be empty:

```sh
tofu -chdir=templates/hetzner-kubernetes-baseline/tofu state list
```

Independently inspect the Hetzner project and confirm that no tutorial-owned
servers, networks, firewalls, primary IPs, or SSH-key registrations remain.
Check for the environment name and the ownership labels rather than relying
only on the command exit status.

Only after both checks are clean:

```sh
unset HCLOUD_TOKEN APPLY_PLAN_ID APPLY_PLAN_PATH
unset DESTROY_PLAN_ID DESTROY_PLAN_PATH INPUT TEMPLATE
```

Retain redacted evidence if required, then remove the disposable local state
and plan records according to your project's retention policy. Never delete
state or plan records before teardown has been independently confirmed.

## Troubleshooting guarded operations

If apply or destroy reports that the reviewed plan no longer matches, do not
bypass the guard. Common causes are:

- the input path or content changed;
- the template changed;
- the saved plan was removed or modified;
- the plan ID belongs to the opposite operation;
- a different environment or template was selected.

Revalidate, create a new plan, review it, and approve its new exact ID.

For general commands, see [Lifecycle operations]({{< relref
"/docs/guides/lifecycle" >}}). For contract fields, see [Contracts]({{< relref
"/docs/reference/contracts" >}}).
