# Variables

## OpenTofu variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `environment` | `string` | yes | — | 2–32 lowercase letters, digits, or hyphens; starts with a letter | no | Deployment label recorded by the provider-free OpenTofu resource. |
| `host_name` | `string` | no | `localhost` | Valid Ansible inventory host name; maximum 63 characters | no | Inventory identity used for the local conformance host. |

## Ansible variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `deployment_label` | `string` | yes | — | Non-empty string | no | Deployment label used by the example playbook. |
| `message` | `string` | yes | — | Non-empty string | no | Non-secret message reported by the example playbook. |
| `emit_summary` | `boolean` | no | `true` | `true` or `false` | no | Controls whether the example emits its summary message. |

## Cross-variable rules

The deployment supplies `deployment_label`, `message`, and `emit_summary`
through its native Ansible variables. OpenTofu output
contains only the standardized inventory identity, groups, and connection
facts; it is not used to transport template-specific Ansible variables.
ainfra only transports and validates the standardized inventory structure; it
does not translate either deployment-facing variable.

## Examples

See the complete [`minimal`](../examples/minimal/) deployment input set.
