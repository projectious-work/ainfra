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

### 2026-08-13 — Normative template-tree digest

- Implemented the exact `AINFRA-LOCK-005` byte framing: domain separator and
  NUL, byte-sorted normalized paths, unsigned 64-bit big-endian path/content
  lengths, executable-bit marker, and unmodified content bytes.
- Required valid UTF-8 NFC relative paths with slash separators and rejected
  empty, current-directory, and parent-directory segments.
- Excluded `.git/` metadata while refusing `.ainfra/`, `.terraform/`, state
  files, symlinks, hard links, FIFOs, and other special files.
- Bound content, normalized path, and executable status into the digest and
  detected file-size changes during streaming.
- Added an independent normative-framing oracle, mutation assertions,
  adversarial filesystem cases, and deterministic fuzz coverage.

The digest remains a pure read-only primitive. Contained local materialization
is the next slice and will verify this digest before making a cache entry
available to lockfile or doctor consumers.

### 2026-08-13 — Contained local materialization

- Added private digest-addressed cache entries below
  `<cache>/templates/sha256/<digest>` with private staging and atomic rename.
- Copied only validated regular template files, excluded `.git/`, preserved
  only the executable/non-executable distinction, and removed group/other
  permissions from materialized content.
- Recomputed the normative digest after copying and before publication.
- Verified every existing cache hit before reuse and failed closed on poisoned,
  symlinked, non-directory, or otherwise invalid entries.
- Rejected overlapping source/cache roots to prevent recursive or ambiguous
  materialization.
- Added tests for first publication, deterministic reuse, permission policy,
  Git exclusion, cache poisoning, unsafe input, and root overlap.

This primitive accepts explicit source and cache roots produced by policy; it
does not consult ambient configuration or yet expose a CLI mutation. The next
slice resolves `local:` identities against the deployment/repository approved
roots and composes this primitive into lock creation.

### 2026-08-13 — Approved local roots and initial lock creation

- Resolved `local:` identities relative to the deployment and proved them
  contained by either the deployment root or its nearest parent Git repository.
- Rejected traversal beyond those approved roots, symlinked path components,
  invalid Git root markers, missing paths, and non-directory sources.
- Added strict canonical `ainfra.lock` parsing, comparison, and private atomic
  publication with an exclusive writer sentinel.
- Composed approved-root resolution, verified digest-addressed materialization,
  materialized-template contract validation, and initial lock publication into
  `ainfra template lock`.
- Made repeated identical locks deterministic and non-mutating while refusing
  changed existing bindings with an exact `ainfra template update` remediation.
- Added text/JSON command results plus policy, cache, lock, application, and CLI
  regression coverage. Git sources remain explicitly unavailable until the
  argument-array Git adapter is implemented.

The next dependency-ordered slice implements controlled Git acquisition and
immutable commit resolution. It will then share this lock writer and verified
cache path with explicit `template update`.

### 2026-08-13 — Controlled Git acquisition

- Added a shell-free Git adapter that invokes an immutable executable through
  structured argument arrays and an explicit environment inside private
  acquisition staging.
- Fetched the requested ref without configuring a persistent remote, resolved
  `FETCH_HEAD` to a full SHA-1 or SHA-256 commit, and checked out that exact
  detached commit before content use.
- Cleaned and contained optional Git subdirectories, rejecting traversal before
  checkout and relying on the normative tree policy to reject symlinks, hard
  links, special files, runtime state, and poisoned content.
- Materialized selected Git content through the same verified digest-addressed
  cache used by local sources, verifying every cache hit.
- Recorded display-safe executable, argument, and working-directory boundaries
  while redacting repository credentials from records and child diagnostics.
- Added controlled-runner and controlled-executable tests for argument
  boundaries, credential redaction, immutable revision validation, failure,
  cancellation, subdirectory selection, and ref injection attempts.

The next slice composes Git acquisition into `template lock`, introduces the
explicit `template update` comparison flow, and resolves configured Git and
cache policy from the existing trusted configuration layers.

### 2026-08-13 — Git locking and explicit update

- Composed immutable Git acquisition into the same `template lock` path used
  by approved local sources without making Git a prerequisite for local locks.
- Resolved the host Git executable lazily, fingerprinted it before execution,
  and selected only the host environment needed for Git, SSH agent access, and
  non-interactive credential behavior.
- Recorded requested Git refs separately from their immutable resolved commits
  while retaining the canonical source, selected subdirectory, version, tree
  digest, and resolution time in `ainfra.lock`.
- Added `template update` as the explicit path for replacing a changed binding;
  it refuses a missing or malformed existing lock and leaves identical bindings
  untouched.
- Added application, command, and compiled black-box coverage for Git lock
  fields, local content updates, canonical JSON results, and unchanged updates.

The next slice integrates lock and verified-cache findings into deployment and
template doctor scopes, then closes remaining configuration and interruption
coverage before the independent Phase 3 conformance review.

### 2026-08-13 — Lock and cache doctor integration

- Added deployment findings for lock presence and structure, requested-source
  binding, verified cache safety, cached digest equality, and local source drift.
- Reported exact `template lock` or `template update` remediation for missing,
  malformed, stale, unsafe, or poisoned state while keeping doctor read-only.
- Avoided resolving mutable Git refs during doctor: Git-backed deployments check
  the locked immutable commit and verified cache and direct users to the explicit
  update command when they want to resolve the requested ref again.
- Replaced the placeholder template result in `doctor all` with full template
  contract checks whenever the locked cache entry is verified; unavailable or
  invalid resolved templates remain deterministic skip/fail findings.
- Added regression coverage for missing locks, fully verified local locks,
  resolved template checks, and local source drift after locking.

The next slice closes trusted configuration and interruption/corruption edge
coverage, followed by the independent Phase 3 conformance and gap review.
