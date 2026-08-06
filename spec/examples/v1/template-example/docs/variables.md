# Variables

## OpenTofu variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `environment` | `string` | yes | — | 2–32 lowercase letters, digits, or hyphens; starts with a letter | no | Label passed through the standardized inventory to Ansible. |
| `host_name` | `string` | no | `localhost` | Valid Ansible inventory host name; maximum 63 characters | no | Inventory identity used for the local conformance host. |

## Ansible variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `ainfra_message` | `string` | yes | — | Non-empty string | no | Non-secret message reported by the example playbook. |
| `ainfra_emit_summary` | `boolean` | no | `true` | `true` or `false` | no | Controls whether the example emits its summary message. |

## Cross-variable rules

The OpenTofu `environment` value becomes the `ainfra_environment` inventory
variable. The playbook requires that generated value and `ainfra_message`.
ainfra only transports and validates the standardized inventory structure; it
does not translate either deployment-facing variable.

## Examples

See the complete [`minimal`](../examples/minimal/) deployment input set.
