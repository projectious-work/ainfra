---
title: Hetzner private K3s template
weight: 48
---

Phase 9 ships the live-certified provider-backed production candidate at
`templates/hetzner-kubernetes-baseline` in `v1.0.0-alpha.9`.

It combines ownership-labelled Hetzner infrastructure, private management
addresses, initial pinned K3s bootstrap, an externally managed Cloudflare
Tunnel, and an optional temporary SSH bastion. It does not install workloads or
provide day-two Kubernetes operations.

## Security boundary

- Node firewalls expose no public inbound TCP service.
- Generated inventory contains private addresses only.
- The bastion defaults off and requires explicit narrow CIDRs.
- Host keys must be verified out of band and recorded in the deployment's
  bound `known_hosts` file; host-key checking stays enabled.
- Initial private-node configuration uses an operator-owned SSH `ProxyJump`
  route through the temporary bastion. Remove the bastion only after verifying
  the tunnel and an independent private management path.
- SSH private keys, K3s tokens, and tunnel tokens remain external.
- K3s and cloudflared binaries require architecture-specific SHA-256 values.

## Evaluation

The Phase 9 certification exercised a disposable three-node control plane,
verified K3s health and idempotent convergence, confirmed Cloudflare Service
Auth SSH after temporary-bastion removal, executed the exact reviewed destroy
plan, and independently verified that no owned Hetzner resources remained.
That evidence validates the released template candidate; it does not authorize
or certify a future deployment automatically.

Run `ainfra doctor template templates/hetzner-kubernetes-baseline` and the
template's `tests/validate.sh` for offline conformance. The test performs
OpenTofu initialization/validation, Ansible syntax, security policy, and
native-variable documentation drift checks.

Live use additionally requires explicit cost approval, protected provider
credentials, a reviewed apply/configure/check cycle, reviewed bastion removal,
reviewed destroy, and an ownership-scoped Hetzner API query proving that every
owned resource is gone. Offline checks are not certification evidence.
