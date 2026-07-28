# Hetzner Kubernetes-ready baseline

This template will provision one control-plane-capable Debian 13 host and
optional workers on a narrow private network. It does not install Kubernetes.

OpenTofu creates the private network, firewall, SSH key registrations, and
Debian hosts. Ansible then applies host hardening; Kubernetes installation is
deliberately outside this template.

## Image

The complete supported image set currently contains one value:

- `debian-13` — official Hetzner Debian 13 image, updated within its major
  release and validated by this template.

## Networking

Public IPv4 and IPv6 are independent and disabled in the standard input.
Disabling IPv4 means that the provider configuration must not create or attach
a billable public IPv4 resource.

## Ansible

Generate inventory only from a validated, sanitized output document:

```sh
ainfra inventory \
  --output .ainfra/output.json \
  --destination .ainfra/inventory.yml
```

The generated inventory selects private addresses and carries the actual
private network CIDR into the host firewall policy. Public management CIDRs
remain empty by default.

Host-key checking is mandatory. Before the first Ansible connection, verify
each SSH host-key fingerprint through a trusted console or other out-of-band
channel and add it to `known_hosts`. Do not use `ssh-keyscan` as the source of
trust.

Local lint and syntax validation runs through `scripts/test-ansible`. The
cost-bearing disposable live test runs `scripts/verify-ansible-host`, which
performs check mode, an apply, and a second apply that must report zero
changes. That test requires separate explicit approval before provisioning.
