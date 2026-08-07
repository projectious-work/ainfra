# ainfra v1 product and engineering specification

| Field | Value |
|---|---|
| Status | Ready for implementation handoff |
| Specification version | 1.0.0-draft.1 |
| Intended product line | ainfra v1 |
| Primary implementation | New Go CLI |
| Last updated | 2026-08-07 |

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
Operators edit whatever native OpenTofu and Ansible input files their selected
template documents, then list those files in `ainfra.yaml`. ainfra owns
execution state and audit evidence; OpenTofu owns infrastructure state and
backend behavior.

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
11. [Acceptance, migration, and implementation
    selections](11-acceptance-migration.md)
12. [CLI configuration and logging](12-cli-configuration-and-logging.md)
13. [Testing strategy](13-testing-strategy.md)
14. [Go source conventions](14-go-source-conventions.md)
15. [MCP server mode](15-mcp-server-mode.md)
16. [Market positioning](16-market-positioning.md)
17. [Spec-driven development cycle](17-spec-driven-development-cycle.md)

Roadmap data:

- [Implementation roadmap](roadmap.yaml)

Roadmap status has commitment semantics: `planned` means intended work with an
accepted place in the delivery sequence; `idea` is deliberately non-committal
and may be promoted, reshaped, or removed after validation. `in_progress` and
`shipped` describe implementation evidence rather than aspiration.

Reference diagrams:

- [Product execution flow](architecture-overview.svg)
- [Go package dependencies](package-dependencies.svg)

Supporting material:

- [`../../examples/v1/`](../../examples/v1/) contains conforming deployment
  and template examples.
- [`../../schemas/v1/`](../../schemas/v1/) contains JSON Schema contracts for
  ainfra-owned documents.
- [`../../tests/v1/`](../../tests/v1/) contains negative schema fixtures used
  by `scripts/validate-v1-spec`; they are expected to fail validation.

## Governing decisions

The specification incorporates these accepted decisions from the company
coordination repository:

- `DEC-20260805_1450-DeftWren`: keep ainfra a native OpenTofu and Ansible
  orchestrator.
- `DEC-20260805_1518-BalancedHorizon`: rewrite from scratch as a simple Go CLI.
- `DEC-20260805_1625-TallTiger`: publish Linux/macOS binaries and an optional
  Dockerfile, but no Windows binary or published container image.
- `DEC-20260806_1115-VividDell`: keep standardized OpenTofu output a closed,
  versioned inventory handoff rather than an arbitrary Ansible-variable
  channel.
- `DEC-20260806_1148-BriskRabbit`: use layered CLI configuration with committed
  deployment policy and typed effective-configuration diagnostics.
- `DEC-20260806_1313-HumbleTulip`: accept equivalent deployment path forms and
  make Ansible conditional for infrastructure-only templates.
- `DEC-20260806_1439-ProudFlame`: enforce standard JSON Schema formats in
  specification and runtime validation.
- `DEC-20260806_1629-SureBison`: adopt the company-wide Git branching and
  release-promotion standard.
- `DEC-20260806_1803-FreshAnt`: deliver the read-only MCP stdio server mode in
  v1 after the core lifecycle and hardening phases.

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

## Publishing format

The numbered chapter files are content fragments intended for both repository
review and inclusion in a future Hugo site:

- fragments contain no front matter and no page-level H1 heading;
- the including Hugo page owns its title, metadata, navigation, and H1;
- fragment headings begin at H2 and use portable Markdown/Goldmark syntax;
- fragments contain no Hugo shortcodes or theme-specific HTML; and
- the publishing project is responsible for mapping relative document and SVG
  links into its output structure.

A Hugo page can therefore read a chapter as a page resource or repository file
and pass it through `markdownify` without stripping source-specific wrappers.
The roadmap is YAML data rather than rendered Markdown so a future site can
group and present it without parsing prose.
