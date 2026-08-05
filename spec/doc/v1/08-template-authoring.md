# 8. Template authoring

## Authoring promise

A competent infrastructure engineer or AI coding agent MUST be able to create a
new template by following this specification and the template-authoring guide,
without inspecting or modifying ainfra implementation source.

## Workflow

1. Choose a provider, deployment outcome, and teardown ownership boundary.
2. Create the required template layout.
3. Write the minimal manifest.
4. Author a native OpenTofu root module with pinned providers.
5. Define native variables and complete examples.
6. Emit the standardized non-secret inventory output or declare no inventory.
7. Author native Ansible content with pinned collections and idempotent roles.
8. Document prerequisites, credentials, costs, access, failure modes, and
   direct-tool commands.
9. Add offline policy tests and local engine validation.
10. Execute a cost-approved disposable lifecycle and retain redacted evidence.

## Template requirements

- **AINFRA-TPL-001:** variables MUST be native OpenTofu variables documented in
  `variables.tf` and the template README.
- **AINFRA-TPL-002:** host configuration variables MUST be native Ansible
  variables documented with defaults and supported values.
- **AINFRA-TPL-003:** templates MUST NOT depend on ainfra generating provider-
  specific tfvars or Ansible variables.
- **AINFRA-TPL-004:** provider versions and module sources MUST be pinned with a
  committed OpenTofu lockfile where applicable.
- **AINFRA-TPL-005:** Ansible collections and external roles MUST be pinned.
- **AINFRA-TPL-006:** private keys and passwords MUST NOT be generated unless a
  template-specific security review explicitly justifies an ephemeral value
  and proves safe delivery/destruction. Certified baseline templates SHOULD
  prohibit this entirely.
- **AINFRA-TPL-007:** public ingress MUST be opt-in, narrow, documented, and
  represented in plan review.
- **AINFRA-TPL-008:** every resource SHOULD receive deployment ownership labels
  where supported.
- **AINFRA-TPL-009:** outputs MUST be minimized to stable, non-secret facts.
- **AINFRA-TPL-010:** Ansible execution MUST be idempotent; a second check-mode
  pass after convergence reports zero change.

## Required README content

- purpose and explicitly unsupported outcomes;
- provider/account prerequisites;
- credentials and least-privilege guidance;
- architecture diagram;
- native tfvars reference with every supported value;
- native Ansible variables reference;
- network and access model;
- estimated cost categories and billable opt-ins;
- state/backend requirements;
- plan/apply/configure/destroy walkthrough;
- direct OpenTofu and Ansible equivalents;
- failure and recovery guidance;
- teardown and independent verification;
- template version and compatibility policy;
- live validation status, date, and limitations.

## Provider example: temporary bastion and tunnel

A Hetzner Kubernetes-ready template with Cloudflare Tunnel and temporary SSH
bastion access would own all related semantics. Its tfvars might select node
counts, Hetzner placement, Cloudflare identifiers/references, private networks,
and temporary bastion ingress. Its Ansible variables would select host
hardening and tunnel configuration. The template—not ainfra—defines how the
bastion is created and removed.

It MUST document and test:

- tunnel credentials remain external and secret;
- cluster nodes default to private management addresses;
- bastion ingress is explicit and narrow;
- host-key fingerprints are verified independently;
- bastion removal does not remove the remaining supported management path;
- destroy covers Cloudflare and Hetzner resources owned by the deployment.

## Conformance command

`ainfra doctor template <source>` SHOULD perform layout/schema checks,
OpenTofu formatting and local validation when prerequisites are already
available, Ansible syntax checks, dependency pin checks, secret scans,
required-output checks using fixtures, and documentation presence checks.
Child-tool checks MUST respect the doctor boundary: ainfra reports what
OpenTofu or Ansible determines without interpreting provider, topology,
connectivity, or application semantics.

The same report explains local tool incompatibilities, source/lock drift,
deprecated contract features, missing migration steps, and checks that could
not run. Template repositories SHOULD keep fixtures for every supported
contract version so migration and diagnostic behavior can be reproduced
without provider credentials.

Every template contract release MUST publish machine-readable deprecations and
a human migration guide. If a transformation is safe and deterministic, it MAY
also provide a migration implemented by the ainfra `migration` package. A
template-specific infrastructure or state migration belongs in template
documentation and playbooks, not in the generic CLI migrator.
