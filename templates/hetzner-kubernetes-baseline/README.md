# Hetzner private K3s baseline

This production-oriented template provisions a private-management Hetzner
cluster, bootstraps pinned K3s with native Ansible, and runs a connector for an
externally managed Cloudflare Tunnel. A public SSH bastion is an explicit,
temporary option and is disabled by default.

It stops after initial cluster membership verification. It does not install
workloads, operate upgrades or backups, manage Cloudflare Access policy, create
tunnel credentials, or provide day-two Kubernetes operations.

## Prerequisites and credentials

- OpenTofu `>=1.10.0,<2.0.0`, Ansible Core `>=2.18.0,<3.0.0`, and ainfra;
- a least-privilege Hetzner API token in `HCLOUD_TOKEN`;
- an existing Cloudflare Tunnel, DNS hostname, and Access policy;
- its connector token delivered only through a protected Ansible vars file;
- operator-owned Ed25519 public keys and independently verified host keys; and
- an operator-configured protected backend; the locked baseline uses local
  state for disposable certification only.

The template never generates private keys, passwords, K3s tokens, or tunnel
tokens. Do not commit credentials or state. Production operators should lock a
derivative that replaces the local backend declaration with a capability-tested
encrypted remote backend.

## Architecture

```text
Cloudflare Access -> external Tunnel -> cloudflared on K3s servers
                                           |
operator network -> temporary bastion -----+-> private node addresses
                                           |
                         Hetzner private network -> K3s server/agent nodes
```

Nodes have public IPv4 only for outbound package and tunnel connectivity. The
node firewall has no public inbound TCP rule; generated inventory uses private
addresses. The bastion has only explicitly allowlisted TCP/22 ingress.

## Variables and costs

The complete versioned native API is in
[`docs/variables.md`](docs/variables.md). Recurring costs include Hetzner
servers and provider-billed IPv4/network services. The temporary bastion adds
a server and IPv4 charge only while enabled. Cloudflare plan charges and the
OpenTofu backend are externally owned.

## Lifecycle

From a deployment copied from `examples/minimal/`:

Replace the placeholder `known_hosts` beside `ainfra.yaml` before locking.
Populate it only from fingerprints verified through the Hetzner console or
another independent channel; do not use an unauthenticated `ssh-keyscan`. The
deployment binds this file into the reviewed run, and ainfra refuses SSH
inventory without it.

```sh
ainfra doctor
ainfra template lock
ainfra plan
ainfra apply --plan <reviewed-run-id>
ainfra configure --run <applied-run-id>
```

Run configure again in check mode and require zero changes. Verify the tunnel,
node membership, and host fingerprints independently. If a bastion was used,
set `enable_temporary_bastion=false`, create and review a new apply plan, and
confirm through the Hetzner API that its `temporary=true` resource is absent.

The temporary bastion workflow uses the private node address and OpenSSH
ProxyJump. First verify both fingerprints through the Hetzner console and add
them to `known_hosts`; do not derive trust with `ssh-keyscan`. Then connect:

```sh
ssh -J ainfra@<bastion-public-ip> ainfra@<node-private-ip>
```

For `ainfra configure`, put the equivalent route in the operator's native SSH
configuration. Ansible uses the route while ainfra independently forces strict
host-key checking against the bound `known_hosts` file:

```sshconfig
Host 10.42.*
  User ainfra
  ProxyJump ainfra@<bastion-public-ip>
  IdentityFile /absolute/operator-owned/path/id_ed25519
```

Before bastion removal, prove the Cloudflare Access hostname reaches the same
private SSH service with host-key checking enabled. After applying the removal
plan, require an empty ownership-scoped provider query for `temporary=true` and
repeat the tunnel connection. Cluster nodes never permit direct public SSH.

Teardown requires a separately reviewed destroy plan and exact apply. After
destroy, query every paginated Hetzner resource collection using the ownership
labels printed by the template and wait a bounded interval for an empty result.
The external Cloudflare Tunnel is not owned or destroyed by this template;
revoke or retain it according to its separately documented lifecycle.

## Direct native diagnosis

```sh
tofu -chdir=tofu fmt -check
tofu -chdir=tofu init -backend-config=../backend.hcl
tofu -chdir=tofu validate
tofu -chdir=tofu plan -var-file=../terraform.tfvars -out=plan.tfplan
tofu -chdir=tofu apply plan.tfplan
tofu -chdir=tofu output -json ainfra_inventory
ansible-playbook -i inventory.yaml -e @ansible-vars.yaml ansible/site.yml
tofu -chdir=tofu plan -destroy -var-file=../terraform.tfvars -out=destroy.tfplan
tofu -chdir=tofu apply destroy.tfplan
```

ainfra uses Ansible Runner for structured evidence; `ansible-playbook` is the
direct recovery equivalent. Never disable host-key checking or use
`ssh-keyscan` as the trust source.

## Failure, recovery, and compatibility

Inspect the owning engine and `ainfra status` after any interruption; never
blindly repeat apply or destroy. Loss of `.ainfra/` removes local evidence, not
provider resources. Recover state through the configured backend and verify
provider inventory before continuing.

Breaking native-variable, output, access, or teardown changes require a new
template major version and migration guide. Additive optional inputs may use a
minor release. Supported image and binary upgrades require a new disposable
lifecycle.

## Validation status

Offline schema, policy, OpenTofu initialization/validation, Ansible syntax,
secret scanning, and documentation checks were completed on 2026-08-20. Live
certification remains pending a separately cost-approved disposable lifecycle,
zero-change convergence proof, bastion removal, and independent Hetzner
teardown evidence. Until then this is a production candidate, not a certified
deployment recommendation.
