# ainfra v1 product and engineering specification

| Field | Value |
|---|---|
| Status | Draft for owner review |
| Specification version | 1.0.0-draft.1 |
| Intended product line | ainfra v1 |
| Primary implementation | New Go CLI |
| Last updated | 2026-08-05 |

This directory is the canonical specification for the from-scratch ainfra v1
implementation. The existing Rust implementation is evidence and history, not
an implementation baseline or compatibility constraint.

The words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are to
be interpreted as normative requirement levels. Requirement identifiers are
stable references for implementation work, tests, reviews, and documentation.

## Product statement

ainfra is a small, transparent CLI that applies reusable infrastructure
templates composed of native OpenTofu and Ansible source. It acquires and
verifies templates, invokes OpenTofu, converts declared non-secret outputs into
Ansible inventory, invokes Ansible Runner, preserves review boundaries, and
provides explicit, safe teardown.

ainfra does not define an infrastructure language above OpenTofu or Ansible.
Operators edit native `terraform.tfvars` and `ansible-vars.yaml` files.

## Specification map

1. [Product boundary](01-product-boundary.md)
2. [Concepts and layouts](02-concepts-and-layouts.md)
3. [Contracts and template sources](03-contracts-and-sources.md)
4. [CLI and lifecycle](04-cli-and-lifecycle.md)
5. [Security and trust](05-security-and-trust.md)
6. [Software architecture](06-software-architecture.md)
7. [Build and quality environment](07-build-and-quality.md)
8. [Template authoring](08-template-authoring.md)
9. [Documentation and AI agents](09-documentation-and-ai-agents.md)
10. [Release engineering](10-release-engineering.md)
11. [Acceptance, migration, and open decisions](11-acceptance-migration.md)

Reference diagrams:

- [Product execution flow](architecture-overview.svg)
- [Go package dependencies](package-dependencies.svg)

Supporting material:

- [`../../examples/v1/`](../../examples/v1/) contains conforming deployment
  and template examples.
- [`../../schemas/v1/`](../../schemas/v1/) contains JSON Schema contracts for
  ainfra-owned documents.

## Governing decisions

The specification incorporates these accepted decisions from the company
coordination repository:

- `DEC-20260805_1450-DeftWren`: keep ainfra a native OpenTofu and Ansible
  orchestrator.
- `DEC-20260805_1518-BalancedHorizon`: rewrite from scratch as a simple Go CLI.
- `DEC-20260805_1625-TallTiger`: publish Linux/macOS binaries and an optional
  Dockerfile, but no Windows binary or published container image.

If this specification conflicts with an accepted decision, the decision wins
until the specification is amended.

## Conformance model

There are three independent conformance claims:

- **CLI conformance:** an ainfra binary satisfies all applicable CLI,
  lifecycle, security, and machine-interface requirements.
- **Template conformance:** a template satisfies the directory, manifest,
  engine, output, documentation, and acceptance requirements.
- **Deployment conformance:** a deployment contains valid ainfra metadata and
  native engine inputs, pins its template, and keeps local operational material
  out of version control.

The schemas validate document structure. They do not replace behavioral,
security, or lifecycle requirements in this specification.
