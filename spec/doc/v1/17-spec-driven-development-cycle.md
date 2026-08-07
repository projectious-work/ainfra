## Scope

This chapter adapts the projectious.work spec-driven development-cycle
standard to the ainfra v1 implementation. It governs every roadmap phase and
material cross-phase change.

The current `spec/doc/v1/` tree, schemas, examples, and acceptance journeys are
the implementation baseline for this amendment. This chapter does not decide
whether future specifications remain in `spec/`, move into processkit
Artifacts, or use OpenSpec change artifacts. That company-level question is
deferred. Any baseline change remains explicit and reviewable.

## Phase baseline

Before planning a roadmap phase, record:

- the specification commit and roadmap phase ID;
- every applicable chapter and `AINFRA-*` requirement;
- schemas, positive and negative examples, and acceptance journeys;
- governing decisions and company standards;
- implementation-time selections required by chapter 11;
- unresolved ambiguities; and
- evidence required to complete the phase.

The phase does not enter implementation while an essential product, security,
compatibility, or ownership ambiguity remains unresolved.

- **AINFRA-DEV-001:** every implementation phase MUST identify an immutable or
  exact specification baseline and its applicable requirement inventory before
  implementation planning.
- **AINFRA-DEV-002:** unresolved ambiguity that can materially change behavior,
  security, compatibility, or architecture MUST block affected implementation.

## Requirement, dependency, and risk inventory

Create a traceability matrix mapping each applicable normative requirement to
planned work and expected evidence. Classify requirements by package or
component, public interface, security, compatibility, persistence, recovery,
concurrency, external process, documentation, test layer, and ordering.

Schemas and examples are executable contracts. Explanatory prose is reviewed
for contradiction but is not mechanically treated as a separate requirement
when a stable `AINFRA-*` requirement already captures the obligation.

- **AINFRA-DEV-003:** no applicable normative requirement, schema, maintained
  example, or acceptance journey MAY disappear between baseline inventory,
  planning, implementation, and verification.

## Detailed implementation plan

The phase plan defines:

- task dependency graph and implementation waves;
- expected packages, files, commands, schemas, and documentation;
- public and internal interfaces created or changed;
- tests, fixtures, and validation commands;
- security, compatibility, migration, interruption, and recovery effects;
- phase-note, roadmap, changelog, and user-documentation work;
- integration checkpoints and rollback strategy;
- owner and independent reviewer for every work package; and
- requirement IDs addressed by every task.

The plan covers conformance and evidence work, not only production code.

## Agent and model assignment

Use the maximum useful parallelism justified by the task graph and integration
cost. Shared interfaces and ownership boundaries are established before
parallel work starts. Each work package has exclusive or explicitly coordinated
file ownership and a defined output contract.

Use cost-efficient models for bounded inventories, mechanical implementation,
fixtures, and routine checks. Use deeper reasoning for architecture, security,
migrations, ambiguous failures, plan synthesis, and conformance review. Work
that depends on one evolving shared mental model remains sequential. One owner
is responsible for integration across all packages.

- **AINFRA-DEV-004:** parallel agent work MUST have independently useful
  boundaries, explicit ownership, frozen shared contracts, and an integration
  owner.
- **AINFRA-DEV-005:** model selection MUST consider task complexity, risk, and
  cost rather than applying one model class to every work package.

## Independent plan-conformance review

Before implementation, a reviewer other than the primary plan author checks
the plan against every applicable specification file, normative requirement,
schema, example, acceptance journey, governing decision, and company standard.
Decision records referenced for provenance do not add hidden requirements;
the accepted checked-in specification commit is the complete normative
baseline for this review.

Each requirement receives one result:

```text
covered
not applicable, with rationale
blocked by an unresolved decision
requires a specification change
deferred with explicit authority
```

The review detects missing work, invalid sequencing, overlapping ownership,
insufficient negative testing, undocumented compatibility effects, and
unjustified exclusions. Findings amend the plan before approval.

- **AINFRA-DEV-006:** implementation MUST NOT start until the revised plan has
  complete requirement disposition and an accepted conformance review.

## Controlled implementation waves

Baseline the accepted plan revision, ownership, evidence, exclusions, risks,
and integration order. Implement shared contracts and foundations first, then
independent components, then integration checkpoints. Run affected checks
after each wave and resolve interface drift before dependent work continues.

Each work package reports requirements addressed, files changed, tests run,
deviations, remaining risks, documentation impact, and proposed baseline
changes. Implementers do not expand scope silently.

## Specification discoveries and changes

When implementation exposes a specification defect or required change, stop
affected work and propose a baseline change. The proposal describes rationale
and impact on requirements, plan, schemas, examples, tests, security,
compatibility, documentation, acceptance, and roadmap.

Continue only after the governing process accepts or rejects the change. An
accepted change updates the exact baseline and plan; completed work is
rechecked against affected requirements.

- **AINFRA-DEV-007:** implementation convenience MUST NOT silently redefine the
  specification or retroactively weaken acceptance criteria.
- **AINFRA-DEV-008:** every accepted baseline change MUST trigger documented
  impact analysis and re-verification of affected completed work.

## Independent implementation-conformance review

After integration, an independent reviewer checks implementation and evidence
against the same complete baseline used for plan review. Inspect code, tests,
schemas, examples, generated output, documentation, phase notes, roadmap, and
release material rather than relying on implementer summaries.

Each requirement receives one result:

```text
satisfied
partially satisfied
not satisfied
not applicable
unverifiable
specification inconsistency
```

Verification covers negative, interruption, recovery, and partial-failure
behavior as well as successful journeys.

- **AINFRA-DEV-009:** an implementation phase MUST have an independent,
  requirement-complete conformance review before completion.

## Gap closure and phase completion

Every result other than satisfied or accepted non-applicability produces an
implementation correction, added evidence, documentation correction,
controlled specification change, authorized roadmap deferral, or blocked phase.
Repeat affected verification after correction.

A phase becomes `shipped` only when:

- implementation is integrated and required checks pass;
- no unexplained skips remain;
- the conformance matrix is accepted;
- specification, schemas, and examples reflect implemented behavior;
- user documentation describes observed behavior;
- the phase note records decisions, deviations, and validation;
- roadmap, changelog, and release material are updated; and
- follow-up work has explicit ownership and location.

- **AINFRA-DEV-010:** a gap report without correction, authorized deferral, or
  an explicit blocked state MUST NOT be treated as phase completion.
- **AINFRA-DEV-011:** roadmap status MUST NOT change to `shipped` until code,
  tests, specification, documentation, phase evidence, and release identity
  agree.

## Retrospective and cycle record

Significant phases review planned versus actual decomposition, agent and model
cost-effectiveness, integration friction, specification gaps, late defect
detection, and validation value. Reusable findings update company standards,
project guidance, processkit, or routing policy.

The retained phase record contains baseline, requirement inventory,
traceability matrix, approved plan, task graph, ownership, plan-conformance
result, implementation evidence, implementation-conformance result, accepted
baseline changes, gap closures, and completion approval.

- **AINFRA-DEV-012:** the phase record MUST be sufficient for a later
  maintainer to reconstruct what was intended, assigned, changed, verified,
  deferred, and accepted without relying on chat history.
