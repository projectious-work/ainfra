## Authoring promise

A competent infrastructure engineer or AI coding agent MUST be able to create a
new template by following this specification and the template-authoring guide,
without inspecting or modifying ainfra implementation source.

Every authored template MUST validate all ainfra-owned documents against the
schema declared by their `apiVersion`. Schema validity is necessary but not
sufficient: the template MUST also satisfy applicable behavioral, security,
documentation, engine, and lifecycle requirements before it can claim
conformance or certification.

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
9. Validate every ainfra-owned document against its declared schema.
10. Add offline policy tests and local engine validation.
11. Execute a cost-approved disposable lifecycle and retain redacted evidence.

## Template requirements

- **AINFRA-TPL-000:** a template MUST pass the published ainfra schemas for its
  declared contract version; unknown fields or versions MUST fail conformance.
- **AINFRA-TPL-001:** variables MUST be native OpenTofu variables documented in
  `variables.tf` and `docs/variables.md`.
- **AINFRA-TPL-002:** host configuration variables MUST be native Ansible
  variables documented in `docs/variables.md` with defaults and supported
  values.
- **AINFRA-TPL-003:** templates MUST NOT depend on ainfra generating provider-
  specific tfvars or Ansible variables.
- **AINFRA-TPL-004:** provider versions MUST be constrained and every resulting
  `.terraform.lock.hcl` provider selection MUST be committed. Module sources
  MUST use immutable native versions or revisions. A template using only
  built-in OpenTofu functionality MAY omit a lockfile when OpenTofu generates
  none.
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
- **AINFRA-TPL-011:** templates MAY define any provider- or deployment-specific
  native variables, but MUST treat that variable surface as a documented,
  versioned template API with migration guidance for breaking changes.

## Required README content

- purpose and explicitly unsupported outcomes;
- provider/account prerequisites;
- credentials and least-privilege guidance;
- architecture diagram;
- link to the standard native-variable reference;
- network and access model;
- estimated cost categories and billable opt-ins;
- native backend-configuration requirements and engine-owned state guidance;
- plan/apply/configure/destroy walkthrough;
- direct OpenTofu and Ansible equivalents;
- failure and recovery guidance;
- teardown and independent verification;
- template version and compatibility policy;
- live validation status, date, and limitations.

## Standard variable reference

Every template MUST provide `docs/variables.md`. It documents the complete
deployment-facing input API while leaving native OpenTofu and Ansible source as
the authoritative executable contract. Internal role variables that a
deployment is not expected to set need not be exposed.

The file contains these H2 sections in order:

1. `OpenTofu variables`;
2. `Ansible variables`;
3. `Cross-variable rules`; and
4. `Examples`.

Each engine section uses this table shape:

| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |
|---|---|---|---|---|---|---|
| `example_name` | `string` | yes | — | Non-empty; template-specific constraint | no | What the value controls and its operational effect. |

Documentation rules are:

- list every value the deployment may or must set, using its exact native name;
- distinguish required values from values with native defaults;
- reproduce defaults faithfully, using `—` only when no default exists;
- state enumerations, ranges, formats, units, conditional requirements,
  conflicts, and relationships in “Valid values and constraints”;
- describe nested objects, lists, and maps field-by-field below the table when
  one row cannot express the shape clearly;
- mark sensitive inputs and name the supported external delivery mechanism,
  never an example secret value;
- explain infrastructure, access, cost, replacement, and teardown consequences
  where changing a value can have them;
- mark deprecations with the first deprecated version, replacement, and planned
  removal version; and
- keep examples minimal, non-secret, and valid against the documented version.

`Cross-variable rules` documents conditions that span variables or engines,
such as exactly-one-of constraints or an Ansible choice that depends on an
OpenTofu-created capability. It explains the relationship but MUST NOT cause
ainfra to translate or synchronize the values.

- **AINFRA-TPL-012:** `docs/variables.md` MUST contain every deployment-facing
  native variable in the standard form above.
- **AINFRA-TPL-013:** native declaration type, required/default state,
  validation, and sensitivity are authoritative; contradictory documentation
  is a conformance failure.
- **AINFRA-TPL-014:** template tests MUST detect undocumented public inputs and
  stale documented names where the native engine exposes that information.
- **AINFRA-TPL-015:** ainfra doctor MUST check the reference structure and MAY
  delegate drift checking to pinned native documentation tools; it MUST NOT add
  its own provider-variable schema.
- **AINFRA-TPL-016:** after the credential-slot contract ships, templates MUST
  declare each deployment credential's lifecycle phases, consuming engine, and
  engine-native destination without choosing an acquisition provider.
- **AINFRA-TPL-017:** template credential documentation MUST include the
  minimum required provider scope, direct-tool equivalent, expiry or renewal
  expectations, and safe teardown requirements.

## Provider example: temporary bastion and tunnel

A Hetzner Kubernetes-ready template with Cloudflare Tunnel and temporary SSH
bastion access would own all related semantics. Its tfvars might select node
counts, Hetzner placement, Cloudflare identifiers/references, private networks,
and temporary bastion ingress. Its Ansible variables would select host
hardening and tunnel configuration. The template—not ainfra—defines how the
bastion is created and removed.

[kubeclaw](https://github.com/projectious-work/kubeclaw) is an early prototype
for a comparable secure Kubernetes deployment and MAY inform the first
template's architecture, threat model, and operational lessons. It is reference
material, not ainfra source, a reusable template contract, an implementation
dependency, or evidence of v1 conformance. The resulting template MUST be
authored and accepted independently against the ainfra schemas and lifecycle.

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
