# Native variable reference

## OpenTofu variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `environment` | `string` | yes | — | Lowercase DNS-style name, 2-31 characters | no | Names and labels every owned resource. Changing it replaces resource identity. |
| `location` | `string` | no | `nbg1` | `fsn1`, `nbg1`, or `hel1` | no | Hetzner location; changing it replaces servers. |
| `control_plane_count` | `number` | no | `3` | `1`, `3`, or `5` | no | Number of K3s server nodes and the primary cost driver. |
| `worker_count` | `number` | no | `0` | Integer from `0` through `20` | no | Optional K3s agent nodes; each adds a server charge. |
| `image` | `string` | no | `debian-13` | Exactly `debian-13` | no | Supported image alias; changes replace servers. |
| `server_type` | `string` | no | `cx23` | Valid Hetzner server type | no | Cluster-node size and recurring compute cost. |
| `bastion_server_type` | `string` | no | `cx23` | Valid Hetzner server type | no | Temporary bastion size and transient cost. |
| `private_cidr` | `string` | no | `10.42.0.0/16` | RFC1918 IPv4, `/16` through `/24` | no | Private management network; changes replace attachments. |
| `admin_ssh_public_keys` | `list(string)` | yes | — | One or more Ed25519 public keys | no | Registers public keys only; private keys stay external. |
| `enable_temporary_bastion` | `bool` | no | `false` | Boolean | no | Billable temporary SSH access; disable after tunnel verification. |
| `bastion_admin_cidrs` | `list(string)` | no | `[]` | IPv4 `/24` or IPv6 `/64` and narrower | no | Bastion source allowlist; required when enabled. |
| `tunnel_hostname` | `string` | yes | — | Valid DNS hostname | no | External Cloudflare Access/Tunnel endpoint retained after bastion removal. |

## Ansible variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `ainfra_admin_user` | `string` | no | `ainfra` | Existing non-root account | no | SSH administration account. |
| `ainfra_ssh_port` | `integer` | no | `22` | `1`-`65535` | no | Private SSH port allowed by nftables. |
| `ainfra_private_cidr` | `string` | no | `10.42.0.0/16` | Must equal OpenTofu `private_cidr` | no | Private SSH source network. |
| `ainfra_unattended_reboot` | `bool` | no | `false` | Boolean | no | Allows unattended-upgrades to reboot. |
| `ainfra_unattended_reboot_time` | `string` | no | `03:30` | 24-hour `HH:MM` | no | Automatic reboot time when enabled. |
| `k3s_version` | `string` | no | `v1.36.1+k3s1` | Exact documented release | no | Pinned initial K3s version. |
| `k3s_binary_sha256` | `string` | yes | `""` | 64 lowercase hex characters | no | Architecture-specific upstream checksum. |
| `k3s_cluster_cidr` | `string` | no | `10.52.0.0/16` | Must not overlap the Hetzner private CIDR | no | K3s pod network. |
| `k3s_service_cidr` | `string` | no | `10.53.0.0/16` | Must not overlap the Hetzner private CIDR | no | K3s service network. |
| `k3s_cluster_token` | `string` | yes | `""` | External value, at least 32 characters | yes | Join token supplied in a protected native vars file. |
| `cloudflared_version` | `string` | no | `2026.7.2` | Exact documented release | no | Pinned connector version. |
| `cloudflared_binary_sha256` | `string` | yes | `""` | 64 lowercase hex characters | no | Architecture-specific upstream checksum. |
| `cloudflare_tunnel_token` | `string` | yes | `""` | Externally issued tunnel token | yes | Protected native input stored root-only. |

## Cross-variable rules

- `enable_temporary_bastion=true` requires at least one
  `bastion_admin_cidrs` entry. The bastion is not permanent management.
- `ainfra_private_cidr` must equal OpenTofu `private_cidr`; ainfra does not
  translate or synchronize native values.
- Verify the tunnel and host fingerprints before reviewing a plan that sets
  `enable_temporary_bastion=false`.
- All managed cluster nodes must use one architecture. Supply the checksum for
  the architecture selected by `server_type`; the tasks reject any host that
  is not `x86_64` or `aarch64`.
- Secrets belong in an untracked mode-`0600` Ansible variable file or an
  equivalent protected external delivery mechanism.

## Examples

Copy `examples/minimal/` outside the template, replace every `REPLACE_` value,
and keep secret-bearing files out of version control. The committed example is
non-secret and cannot configure live hosts unchanged.
