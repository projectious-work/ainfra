# Phase 0: Product specification

## Status and scope

Status: shipped.

This phase established the normative product and engineering contract for the
ainfra v1 rewrite. It is intended for maintainers, implementers, template
authors, and reviewers of later phases.

## Delivered behavior and boundaries

PR #46 defined the v1 product boundary, native-file contracts, security model,
Go architecture, schemas, examples, acceptance journeys, documentation model,
and release-engineering process. The accepted specification is checked in
under `spec/doc/v1/` and its machine-readable contracts under
`spec/schemas/v1/`.

No runtime behavior shipped in this phase. Implementation begins with phase 1
and must remain consistent with the accepted specification.

## Decisions and deviations

The specification records the accepted decisions and alternatives for the v1
rewrite. The phase completed without a known deviation from that specification.

## Validation

The specification schemas, examples, roadmap, and negative format fixtures were
validated with `uv run scripts/validate-v1-spec` after integration.

## Consequences and follow-up

The v1 implementation now has an authoritative contract and a phased delivery
sequence. Phase 1 establishes the Go project foundation. Later phases must
update their phase notes, affected specification, tests, and user documentation
before being marked shipped.
