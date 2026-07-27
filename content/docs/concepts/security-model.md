---
title: Security model
weight: 20
description: Threat assumptions, enforced invariants, and verification layers.
---

## Assumptions

Cloud credentials, state, plans, private keys, generated inventories, and
provider logs may contain sensitive information. The local operator environment
is trusted to hold short-lived credentials; the repository and ordinary output
documents are not secret stores.

The model reduces accidental exposure and unsafe defaults. It does not turn an
untrusted workstation or compromised provider account into a trusted one.

## Invariants

- Operators supply public SSH keys; `ainfra` never creates private keys.
- Root login and password SSH are prohibited.
- SSH host-key checking is mandatory.
- Management ingress is private by default.
- Public IPv4 allocation is opt-in.
- Private networks use narrow RFC1918 ranges and reject broad management
  ranges.
- Non-disposable environments require encrypted, locked, recoverable remote
  state with TLS and access control.
- Providers, collections, roles, images, and scanners are pinned.
- Standard outputs contain references to credentials, never their contents.
- Destructive operations name their exact ownership scope and require
  explicit approval.

## Layered verification

| Layer | What it proves |
|---|---|
| JSON Schema | Document shape, version, enums, and unknown-field rejection |
| Policy checks | Cross-field security invariants |
| OpenTofu validation | Provider configuration and expression correctness |
| Checkov | Known infrastructure-policy findings |
| Gitleaks | Repository secret patterns |
| Plan assertions | The proposed resource graph matches safety expectations |
| Ansible checks | Syntax, lint, check mode, and idempotence |
| Disposable live test | Provider behavior and end-to-end teardown |

No single layer is treated as complete proof. Live Hetzner verification is the
final cost-bearing gate and requires explicit approval.

## SSH trust

`ssh-keyscan` can collect a key but cannot authenticate it. Before the first
connection, compare the server fingerprint through a trusted Hetzner console or
another out-of-band channel and place the verified key in `known_hosts`.

## Reporting vulnerabilities

Do not open public issues containing credentials, state, plans, inventories, or
exploitation details. Use the repository security contact or GitHub private
vulnerability reporting when available.
