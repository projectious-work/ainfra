# Security model

## Assumptions

Cloud credentials, state, plans, private keys, and generated inventories may
contain sensitive information. The local operator environment is trusted to
hold short-lived credentials; the repository and ordinary output documents
are not secret stores.

## Invariants

- Operators supply public SSH keys; ainfra never creates private keys.
- Root login and password SSH are prohibited.
- Management ingress is private by default.
- Public IPv4 allocation is opt-in. When disabled, no billable IPv4 resource
  or attachment may exist.
- Private networks use narrow RFC1918 ranges and reject broad management
  ranges.
- Non-disposable environments require encrypted, locked, recoverable remote
  state with TLS and access control.
- Providers, collections, roles, images, and scanners are pinned.
- Standard outputs contain references to credentials, never their contents.
- Destructive operations name their exact ownership scope and require
  explicit approval.

Schema validation establishes document shape. Policy validation, plan
assertions, scanners, and disposable-environment tests provide independent
layers of enforcement in later milestones.

## Automation

All checks run locally through scripts or the CLI. GitHub Actions and workflow
files are prohibited. Live cloud verification is the final gate and requires
explicit user approval after a cost warning.
