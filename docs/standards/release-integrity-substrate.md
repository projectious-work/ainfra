# Release integrity substrate

This document defines organization-level control objectives for moving any
versioned output from an approved source state to a published destination. It
is intentionally independent of programming language, repository host,
artifact type, deployment model, and regulatory regime.

The substrate applies to compiled software, libraries, containers, mobile
applications, infrastructure modules, machine-learning models, datasets,
configuration bundles, policy packages, documentation sets, and other
controlled deliverables.

## The abstract problem

A release is a chain of custody, not merely a build command. Its integrity
depends on preserving one identity across five domains:

1. **Intent:** the authorized change set and release decision.
2. **Source:** the immutable source revision and dependency inputs.
3. **Production:** the process and environment that create outputs.
4. **Evidence:** tests, reviews, approvals, provenance, and signatures.
5. **Distribution:** the destination identity and bytes consumers receive.

Failures occur when these domains refer to different identities, or when a
mutable operation is allowed after expensive evidence has been approved.

## Minimum company control objectives

### CO-1 — Canonical release identity

Every release must have one immutable identity before artifact production.
The identity should include the source revision, source-tree digest, version,
release channel, and relevant configuration or dependency-lock digests.

Projects without Git should use an equivalent immutable snapshot identifier.
Projects producing non-file outputs should bind the identity to a content
digest, database migration set, model checkpoint, or deployment manifest.

### CO-2 — Ordered state transitions

The delivery system must enforce a forward-only state machine:

```text
candidate → frozen → produced → verified → authorized → distributed
```

Each transition must require evidence from the prior state. Human approval
must occur only after all operations capable of changing output identity.

### CO-3 — Evidence binding

Evidence must identify what it proves. A test result that names only a branch,
version, environment, or timestamp is insufficient when those references can
move. Evidence should bind to immutable source and output identities.

### CO-4 — Separation of mutation and authorization

The actor or system authorizing distribution should not silently alter the
candidate being authorized. High-risk deliverables should additionally
separate producer, approver, and publisher identities.

### CO-5 — Fail-closed verification

Missing, ambiguous, stale, malformed, or mismatched evidence must stop the
workflow. A system must not infer success from partial logs, the presence of
an artifact, or a previous run for the same version.

### CO-6 — Recoverable publication

Distribution must be idempotent and resumable. Recovery must distinguish at
least these states:

- destination identity absent;
- identity created but output transfer incomplete;
- outputs transferred but verification incomplete;
- complete and independently verified.

Retries must verify existing state before continuing. They must never replace
an existing immutable identity with different content.

### CO-7 — Independent consumer verification

Completion requires retrieving outputs through the consumer-facing path and
verifying their digests and authenticity. Verifying only the producer's local
directory does not prove successful distribution.

### CO-8 — Environment and credential isolation

Build caches, trust roots, signing identities, repository credentials, and
publisher credentials need separate explicit locations and lifetimes. A fix
for one tool must not implicitly replace another tool's identity or policy.

### CO-9 — Evidence retention and auditability

The organization must define retention, access, privacy, and disposal rules
for release state, provenance, approvals, logs, vulnerability findings,
signatures, and distribution receipts.

### CO-10 — Exception governance

Emergency releases and waivers must remain visible state transitions. They
need an accountable actor, reason, scope, expiry where applicable, and a
follow-up obligation. An emergency path must not be an undocumented bypass.

## Repository conformance checks

An internal repository baseline can evaluate the following technology-neutral
questions:

| Area | Required question |
| --- | --- |
| Identity | Is one immutable candidate established before production? |
| Ordering | Can merge, rebase, dependency update, or generation occur after approval? |
| Inputs | Are source, dependencies, build policy, and environment identified? |
| Outputs | Are all deliverables content-addressed or checksummed? |
| Evidence | Do tests and reviews bind to immutable input and output identities? |
| Authorization | Is approval explicit, attributable, scoped, and non-replayable? |
| Publication | Can interruption be resumed without replacing trusted content? |
| Verification | Are published outputs retrieved through the consumer path? |
| Credentials | Are producer, signer, and publisher credentials independently scoped? |
| Audit | Can an investigator reconstruct the complete transition history? |
| Exceptions | Are deviations time-bound, attributable, and reviewable? |
| Usability | Are expensive human actions placed after automated failure points? |

Repositories should report `pass`, `not applicable`, or a risk-accepted
exception with evidence. A uniform toolchain is not required; uniform control
outcomes are.

## Risk-based profiles

### Baseline

- immutable source identity;
- reproducible or deterministic packaging where practical;
- automated tests bound to the candidate;
- output checksums;
- protected publication credentials;
- post-publication checksum verification.

### Enhanced

- signed provenance and artifacts;
- independent approval;
- software bill of materials or equivalent component inventory;
- vulnerability and policy gates;
- resumable publication state;
- retained consumer-path verification evidence.

### High assurance

- hosted, isolated, and hardened production environment;
- non-exportable or workload-bound signing identity;
- two-person or policy-engine authorization;
- hermetic or fully declared inputs;
- transparency or append-only audit records;
- independently reproduced or verified outputs;
- tested revocation, rollback, and incident-response procedures.

Profiles should be selected by impact, exposure, reversibility, data
sensitivity, consumer population, and regulatory obligations—not by language
or repository size.

## Examples outside conventional software releases

- **Data pipeline:** freeze query definitions, source snapshots, schema, and
  transformation image; sign the dataset manifest; verify the published table
  or object-store prefix.
- **Machine-learning model:** bind training code, dataset identity,
  hyperparameters, evaluation evidence, model digest, and deployment card.
- **Infrastructure policy:** bind policy sources, compiler version, test
  corpus, generated bundle, and target control-plane revision.
- **Mobile application:** bind source, dependency locks, signing profile,
  store metadata, binary digest, review result, and storefront download.
- **Documentation:** bind source revision, renderer, link/check output, final
  archive or site digest, approval, and public retrieval result.
- **Database change:** bind migration set, compatibility evidence, approval,
  target schema identity, execution receipt, and rollback limitation.

## Standards alignment

This substrate is a control model, not a compliance claim. Organizations can
map it to their applicable frameworks:

- [NIST Secure Software Development Framework](https://csrc.nist.gov/pubs/sp/800/218/final)
  provides outcome-oriented secure-development practices, including
  protecting software and collecting release provenance.
- [SLSA build-track guidance](https://slsa.dev/spec/v1.2/build-track-basics)
  provides progressively stronger expectations for build provenance, hosted
  production, and hardened environments.
- [OpenSSF software supply-chain initiatives](https://openssf.org/technical-initiatives/software-supply-chain/)
  provide complementary practices and tooling for signing, provenance,
  repository policy, and component intelligence.
- [in-toto](https://in-toto.io/) provides a model for recording and verifying
  authorized steps across a software supply chain.

Company policy owners should maintain a separate mapping from these control
objectives to contractual, regulatory, audit, and risk-management controls.
Repository teams should consume a stable profile rather than independently
interpreting every external framework.

## Adoption model for an internal repository

1. Inventory deliverable types and distribution destinations.
2. Identify every operation capable of changing deliverable identity.
3. Place the freeze boundary after the last such operation.
4. Define required evidence and bind it to immutable identities.
5. Move human authorization after automated checks.
6. Make distribution resumable and independently verifiable.
7. Select a risk profile and record justified exceptions.
8. Continuously test interruption, credential failure, replay, and mismatch
   scenarios—not only successful delivery.

The reusable organizational asset is the control contract and evidence
schema. Individual repositories may implement it with CI workflows, release
services, package registries, deployment controllers, or local tools.
