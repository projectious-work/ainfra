---
title: Hetzner private K3s template
weight: 48
---

Phase 9 introduces a provider-backed production candidate at
`templates/hetzner-kubernetes-baseline`.

It combines ownership-labelled Hetzner infrastructure, private management
addresses, initial pinned K3s bootstrap, an externally managed Cloudflare
Tunnel, and an optional temporary SSH bastion. It does not install workloads or
provide day-two Kubernetes operations.

## Security boundary

- Node firewalls expose no public inbound TCP service.
- Generated inventory contains private addresses only.
- The bastion defaults off and requires explicit narrow CIDRs.
- Host keys must be verified out of band; host-key checking stays enabled.
- SSH private keys, K3s tokens, and tunnel tokens remain external.
- K3s and cloudflared binaries require architecture-specific SHA-256 values.

## Evaluation

Run `ainfra doctor template templates/hetzner-kubernetes-baseline` and the
template's `tests/validate.sh` for offline conformance. The test performs
OpenTofu initialization/validation, Ansible syntax, security policy, and
native-variable documentation drift checks.

Live use additionally requires explicit cost approval, protected provider
credentials, a reviewed apply/configure/check cycle, reviewed bastion removal,
reviewed destroy, and an ownership-scoped Hetzner API query proving that every
owned resource is gone. Offline checks are not certification evidence.
