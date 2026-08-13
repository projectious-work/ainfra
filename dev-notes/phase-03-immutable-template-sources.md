# Phase 3: Immutable template sources

Status: in progress.

Phase 3 implements the immutable source boundary assigned by the canonical v1
specification. It starts from Phase 2 release `v1.0.0-alpha.2` and adds source
parsing, acquisition, locking, materialization, digest verification, and drift
diagnosis without beginning OpenTofu or Ansible lifecycle execution.

## Completion outcome

A user can lock or explicitly update a local or Git-subdirectory template
source, obtain a canonical `ainfra.lock`, materialize the exact locked content
inside a private contained cache, and have doctor report source, lock, digest,
or cache drift deterministically in text and JSON.

Normal lifecycle consumers use only the immutable commit and verified digest
from the lock. Mutable Git refs are resolved only by explicit lock/update
commands. Phase 3 does not create plans, apply infrastructure, generate
inventory, configure hosts, or destroy resources.

## Requirement and disposition matrix

| Normative requirement | Disposition | Phase 3 evidence target |
|---|---|---|
| `AINFRA-SOURCE-001`–`006` | implement | Structural local/Git source parser; approved-root containment; argument-array Git adapter; private cache; cleaned subdirectory; Git-owned redirect, SSH, and credential behavior. |
| `AINFRA-LOCK-001`–`006` | implement | Exclusive lock/update writers; canonical serialization; immutable Git commit or normalized local identity; specified SHA-256 tree digest; strict path, file-type, permission, and runtime-artifact handling. |
| `AINFRA-SEC-001`–`006` | implement | Locked identity and digest enforcement; visible source/ref changes; verified cache hits; contained materialization; no credential persistence; URL redaction. |
| `AINFRA-SEC-010`, `013`, `020`–`021`, `025` | supporting | No shell invocation; canonical contained working directories; continued inline-secret refusal; safe fixtures; owner-only operational storage. |
| `AINFRA-CLI-002`–`004`, `006`–`007`, `011`–`013` | supporting | Lock/update commands preserve established result, error, confirmation, help, and redaction contracts. |
| Source-lock doctor checks | implement | Deployment/template/all scopes report lock completeness, immutable identity, digest match, cache containment, unavailable checks, and exact remediation. |
| `AINFRA-TEST-001`–`003`, `014` | implement | Unit, component, compiled black-box, fuzz, archive/path traversal, symlink, special-file, cache-poisoning, interruption, and output-schema coverage. |
| OpenTofu/Ansible lifecycle and plan bindings | deferred | Phases 4–6 retain engine execution, reviewed plans, output, inventory, configuration, destruction, and recovery. |

## Delivery sequence

1. Freeze the Phase 3 command/result and lockfile schema contracts.
2. Implement structural source parsing and canonical source identities.
3. Implement the normative template-tree digest and adversarial fixtures.
4. Implement contained local-source materialization.
5. Implement the Git argument-array adapter and immutable revision resolution.
6. Implement private cache storage with verification on every hit.
7. Implement canonical lock creation and explicit update with visible diffs.
8. Integrate lock/materialization findings into doctor and reconciliation.
9. Add compiled CLI, fuzz, interruption, and cache-poisoning coverage.
10. Reconcile documentation, run an independent conformance review, close
    gaps, and publish the Phase 3 release before marking the phase shipped.

## Validation gates

- Default tests remain deterministic, credential-free, parallel-safe, and
  offline; Git network behavior uses controlled fake executables and local
  repositories.
- Every acquisition call records executable, argument boundaries, contained
  working directory, selected environment names, and sanitized diagnostics.
- Negative coverage includes absolute/traversal paths, Unicode normalization,
  empty segments, symlink and hard-link escape, devices/FIFOs/sockets,
  `.ainfra/` and engine artifacts, mutable-ref drift, digest mismatch, corrupt
  cache entries, credentials in URLs, cancellation, and partial writes.
- Lock and doctor JSON results validate against committed closed schemas.
- Phase completion follows Decision
  `DEC-20260813_0853-CoolFern-ship-phase-2-only-after-verified` as the general
  ordering precedent: review, gap closure, release preparation, publication,
  independent verification, evidence recording, then roadmap shipment.

## Initial risks

- Git URL/subdirectory syntax is ambiguous if parsed by string slicing;
  parsing must be grammar-driven and covered by adversarial cases.
- Filesystem digest portability depends on exact UTF-8 NFC path handling,
  executable-bit normalization, ordering, and byte framing.
- Cache reuse is a trust boundary, not a performance shortcut; every hit must
  be verified before materialization.
- Git and SSH may consult ambient configuration. The adapter must keep those
  mechanisms visible while preventing credentials from entering ainfra-owned
  files or diagnostics.

## Implementation record

### 2026-08-13 — Structural source contract

- Added a pure `internal/source` model for the two v1 source schemes. Parsing
  performs no filesystem, subprocess, cache, or network operation.
- Canonicalized slash-separated local identities while leaving approved-root
  containment to the materialization policy layer.
- Structurally separated absolute HTTPS/SSH repository URLs, optional Git
  subdirectories, and the manifest's mandatory explicit Git ref.
- Added display-safe Git identities that redact URL userinfo and sensitive
  query values without changing the repository value passed to the future Git
  adapter.
- Integrated source validation and canonical identity output into deployment
  loading, so every later doctor, lock, and lifecycle consumer receives the
  same parsed contract.
- Narrowed the deployment JSON schema to `local:` or absolute
  `git::https://`/`git::ssh://` sources, requiring `ref` only for Git.
- Added table-driven, integration, and fuzz regression coverage for unknown
  schemes, missing or invalid refs, absolute paths, separators, empty path
  segments, fragments, control characters, credentials, and sensitive query
  values.

This slice intentionally does not acquire content or write `ainfra.lock`. The
next dependency-ordered slice is the normative template-tree digest, followed
by contained local materialization and then the Git adapter.
