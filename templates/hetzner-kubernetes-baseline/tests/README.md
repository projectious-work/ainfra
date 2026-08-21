# Validation

Run `./validate.sh` from any directory. It performs the offline certification
subset: ainfra layout checks, OpenTofu formatting/init/validate, Ansible syntax,
native-variable documentation drift, dependency pins, and security policy.

It deliberately does not contact Hetzner or Cloudflare and is not live
certification evidence. Live acceptance additionally requires cost approval,
reviewed plan/apply/configure/check/reapply/destroy, bastion removal, and a
paginated ownership-scoped Hetzner API query proving teardown.
