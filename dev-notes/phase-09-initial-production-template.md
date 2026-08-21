# Phase 9: Initial production template

Status: implementation complete; live certification and release pending

Phase 9 adds `templates/hetzner-kubernetes-baseline`, the first provider-backed
ainfra v1 production candidate. It provisions ownership-labelled Hetzner
resources, emits private-address inventory, bootstraps pinned K3s through
template-owned Ansible, runs a checksum-pinned connector for an externally
managed Cloudflare Tunnel, and makes public SSH available only through an
explicit temporary bastion.

## Requirement disposition

| Requirement | Evidence |
|---|---|
| V1 template contract and native inputs | `ainfra-template.yaml`, `tofu/`, `ansible/`, and `examples/minimal/` |
| Private management and narrow ingress | Private inventory output, node firewall without public TCP rules, opt-in bastion CIDRs |
| External secrets and no key generation | Native secret variables, root-only files, `no_log`, and policy tests prohibiting generated credentials |
| Pinned dependencies | hcloud 1.64.0 lockfile, K3s v1.36.1+k3s1, cloudflared 2026.7.2, exact runtime checksums |
| Kubernetes bootstrap boundary | Initial K3s services and membership verification; no workloads or day-two operations |
| Documentation and variable drift | Complete README, four-section native reference, doctor checks, and permanent Go/Python tests |
| Offline engine validation | OpenTofu fmt/init/validate and Ansible syntax in `tests/validate.sh` |
| Live certification | Pending explicit cost approval and disposable provider lifecycle |

## Implementation-time selections

The accepted dependency decision is
`DEC-20260821_0208-FreshMoss-pin-phase-9-provider-and-bootstrap`. The selected
Hetzner provider is 1.64.0, K3s is v1.36.1+k3s1, and cloudflared is 2026.7.2.
The provider is locked through OpenTofu. Runtime binaries require documented
architecture-specific SHA-256 inputs and never resolve a mutable channel.

## Review outcome

Independent requirements, implementation, and documentation audits identified
the current contract, historical-baseline mismatch, and evidence boundary. The
implementation removed the historical v1alpha1 manifest, direct node ingress,
and unsupported provider-variable layer. It added the current v1 manifest,
standard inventory output, tunnel/bastion semantics, initial K3s bootstrap,
complete variable documentation, and offline conformance tests.

No certification claim is made from offline evidence. Phase 9 remains
`in_progress` until a cost-approved lifecycle records reviewed apply,
configure, zero-change check, bastion removal, destroy, and independent
ownership-scoped Hetzner teardown evidence.
