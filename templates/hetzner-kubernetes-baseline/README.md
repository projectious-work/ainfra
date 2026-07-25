# Hetzner Kubernetes-ready baseline

This template will provision one control-plane-capable Debian 13 host and
optional workers on a narrow private network. It does not install Kubernetes.

Milestone 0 contains the contract and non-secret example only. OpenTofu,
cloud-init, Ansible, policies, and their tests are delivered by later
WorkItems.

## Image

The complete supported image set currently contains one value:

- `debian-13` — official Hetzner Debian 13 image, updated within its major
  release and validated by this template.

## Networking

Public IPv4 and IPv6 are independent and disabled in the standard input.
Disabling IPv4 means that the provider configuration must not create or attach
a billable public IPv4 resource.
