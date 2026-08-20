# Native variables

## OpenTofu variables

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `environment_name` | `string` | yes | — | Non-empty after trimming whitespace | no | Labels the provider-free local object; changing it updates that object. |

## Ansible variables

None. This template declares no configurable hosts.

## Cross-variable rules

None. The template has one independent input.

## Examples

```hcl
environment_name = "clean-room"
```
