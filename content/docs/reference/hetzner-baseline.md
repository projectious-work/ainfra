---
title: Hetzner Kubernetes-ready baseline
linkTitle: Hetzner baseline
weight: 30
description: The first pinned infrastructure template.
---

The baseline provisions one control-plane-capable Debian 13 host and optional
workers on a narrow private network. It does not install Kubernetes.

OpenTofu creates:

- a project-owned private network and subnet;
- a narrowly scoped firewall;
- registrations for operator-supplied SSH public keys;
- Debian 13 servers;
- public IPv4 or IPv6 attachments only when explicitly enabled.

Ansible configures:

- secure SSH daemon policy;
- nftables host firewall policy;
- unattended security updates;
- persistent journal and audit configuration;
- Kubernetes-ready operating-system prerequisites.

## Networking

Public IPv4 and IPv6 are independent and disabled in the standard input.
Disabling IPv4 means no billable public IPv4 resource or attachment may exist.
The generated inventory prefers private management addresses.

## Image

The complete supported image set contains one value:

- `debian-13` — the official Hetzner Debian 13 image, updated within its major
  release and validated by this template.

## Host verification

Before the first Ansible connection, verify each SSH host-key fingerprint
through a trusted console or another out-of-band channel. Add that verified key
to `known_hosts`.

The approved disposable live sequence performs:

1. cloud-init schema validation;
2. OpenTofu plan and apply;
3. deterministic SSH host identity verification;
4. Ansible check mode;
5. Ansible apply;
6. a second apply with zero changes;
7. destroy;
8. provider and state checks confirming zero resources.

## Source layout

```text
templates/hetzner-kubernetes-baseline/
├── ainfra-template.yaml
├── inputs/
├── tofu/
├── cloud-init/
└── ansible/
```
