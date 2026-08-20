# Phase 8: Template authoring and conformance

Status: implementation complete; release pending

Phase 8 publishes the v1 template-authoring contract for humans and AI agents,
adds deterministic local conformance findings, and proves that an independent
provider-free template can be authored and validated from the public contract.
It does not certify a provider template or authorize a live lifecycle.

## Implemented boundary

`ainfra doctor template` first loads the strict v1 manifest and contained
template layout. It then checks the standard variable-reference structure,
required README subjects, native variable declarations, clean-room test entry
point, and an applicable standardized-output fixture. An inventory-free
template is not required to invent an output fixture. Native formatting,
validation, dependency, secret, output and idempotency checks remain owned by
OpenTofu, Ansible and the template test script; doctor reports that delegation
as an explicit skip instead of claiming a pass.

Failed conformance findings render a typed partial-failure result and exit with
the dependency code. File inspection uses a scoped filesystem root and never
interprets provider topology, credentials, connectivity or application health.

## Clean-room proof

`spec/examples/v1/template-clean-room` was authored as a second,
infrastructure-only example from the published manifest schema and authoring
guide. It uses only OpenTofu's built-in `terraform_data` resource, declares no
inventory and contacts no provider. Its manifest and deployment validate
against the published schemas; doctor reports seven passes and the one explicit
native-tool delegation; its executable test performs format, init, validate,
plan, apply, destroy plan and destroy apply in a temporary directory.

This proof establishes the Phase 8 authoring promise without claiming the
provider lifecycle, security posture or certification required by Phase 9.

## Requirement disposition

| Requirement | Disposition and evidence |
|---|---|
| AINFRA-TPL-000 | Strict manifest and deployment loading rejects unknown fields, unsupported versions, unsafe paths and invalid applicability. Schema gates validate both published examples. |
| AINFRA-TPL-001–003 | Native OpenTofu/Ansible files remain authoritative; the guide and variable-reference check prohibit an ainfra variable meta-language. |
| AINFRA-TPL-004–011 | The guide specifies dependency pins, external secrets, opt-in ingress, ownership labels, minimal outputs, idempotency and versioned native APIs. Provider-specific enforcement remains in template policy/tests and the native tools. |
| AINFRA-TPL-012–015 | Doctor checks the ordered standard reference and columns; the guide requires native drift tests and explicitly delegates those tests without adding a provider-variable schema. |
| AINFRA-DOC-001–005 | The AI entry point now includes a manifest, native rules, output shape, security and documentation requirements, exact local commands, live acceptance protocol and evidence checklist. A repository test verifies required content and referenced resources. |

## Validation and security

The permanent gates cover Go formatting, vet, static analysis, lint, unit and
black-box behavior, race detection, schemas, CLI contracts, container tests,
vulnerability scanning, security scanning, Hugo generation and both
provider-free template doctors. The clean-room native lifecycle is run as an
additional Phase 8 acceptance check.

The local proof creates no billable resource and retains no credential, state
or raw engine evidence. Live support requires a new, immediate approval naming
the provider account, region, cost ceiling and teardown window, followed by
sanitized lifecycle evidence and independent provider-side teardown
confirmation.

## Compatibility, migration and remaining boundary

The v1 document API is unchanged. Phase 8 tightens only authoring diagnostics;
it does not rewrite templates or native inputs. Existing inventory-free
templates remain valid without artificial output files. Templates failing new
checks receive actionable diagnostics and can update their documentation and
clean-room tests without a contract migration.

Phase 9 is ready to begin authoring the first production provider template.
Actual hosting remains unexecuted until its explicit lifecycle approval gate.
The roadmap stays `in_progress` until a Phase 8 release is published and
independently verified.
