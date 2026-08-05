# Variables

## OpenTofu variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `environment` | `string` | yes | — | Any string in this structural fixture | no | Deployment environment label. |
| `location` | `string` | yes | — | Provider location identifier | no | Placement selected for infrastructure. |
| `control_plane_count` | `number` | yes | — | Whole number; production templates must constrain the minimum | no | Number of control-plane hosts. |
| `worker_count` | `number` | yes | — | Non-negative whole number | no | Number of worker hosts. |
| `server_type` | `string` | yes | — | Provider server-type identifier | no | Compute shape selected for hosts. |
| `private_cidr` | `string` | yes | — | IPv4 CIDR notation | no | Private network address range. |
| `admin_ssh_public_keys` | `list(string)` | yes | — | Public SSH keys only; no private key material | no | Keys authorized for administrative access. |

## Ansible variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `ainfra_admin_user` | `string` | yes | — | Valid target operating-system user name | no | Administrative account used by the example deployment. |
| `ainfra_unattended_reboot` | `boolean` | no | `false` | `true` or `false` | no | Whether unattended maintenance may reboot a host. |
| `ainfra_unattended_reboot_time` | `string` | no | `03:30` | Twenty-four-hour `HH:MM` | no | Preferred unattended reboot time. |

## Cross-variable rules

`ainfra_unattended_reboot_time` has an operational effect only when
`ainfra_unattended_reboot` is `true`. This structural fixture does not translate
values between OpenTofu and Ansible.

## Examples

See the maintained
[`terraform.tfvars`](../../deployment/terraform.tfvars) and
[`ansible-vars.yaml`](../../deployment/ansible-vars.yaml) deployment
examples.
