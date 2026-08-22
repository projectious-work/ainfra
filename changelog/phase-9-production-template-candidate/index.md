# Phase 9 shipped: Private Hetzner K3s

> The first provider-backed template candidate adds private management, tunneled ingress, temporary administration, and pinned K3s bootstrap.


Phase 9's implementation is live-certified. The template keeps
provider and host semantics in native OpenTofu and Ansible, emits the standard
non-secret inventory contract, and preserves ainfra's reviewed lifecycle.

The cost-approved disposable deployment converged with zero change, completed
Ansible check mode without drift, removed its temporary bastion without losing
Cloudflare Service Auth SSH access, and retained a healthy three-server K3s
control plane. Its exact destroy plan completed successfully, and independent
Hetzner API queries confirmed that no Phase 9 resources remained.


---
Source: https://projectious-work.github.io/ainfra/changelog/phase-9-production-template-candidate/index.md
