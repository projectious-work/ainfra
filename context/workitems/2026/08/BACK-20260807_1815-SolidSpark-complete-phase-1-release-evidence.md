---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  created: '2026-08-07T18:15:11+00:00'
  labels:
    phase: '1'
    area: release-engineering
    blocked_by: validation-environment
  updated: '2026-08-10T16:29:15+00:00'
spec:
  title: Complete Phase 1 release-engineering evidence
  state: in-progress
  type: task
  priority: high
  description: 'Implemented the PR #55 two-stage host container gate locally for macOS
    and Linux. The minimal Bash launcher selects an exact owner-approved uv-managed
    Python 3.13.14 by digest, creates runtime/bootstrap exclusively, records its manifest/log
    and host OS, clears inherited configuration, and hands off through a fixed offline
    uv invocation. Platform policy now uses account-database home discovery, macOS
    Library/Caches or Linux .cache state, shasum or sha256sum, fixed Homebrew/Linuxbrew/system
    executable paths, and OS-specific installation guidance. Standard Homebrew symlinks
    are accepted at reviewed fixed paths while executable digests remain recorded
    and validated. The Python entrypoint validates bootstrap layout, identities, hashes,
    environment, immutable input, and authoritative evidence. Documentation, normative
    release text, ShellCheck, and 19 adversarial tests are integrated. scripts/test-all
    and scripts/validate-all pass using isolated temporary Go caches. The controlled
    bundle still requires renewed owner review and real execution on each supported
    Docker-capable host; this restricted harness received no container-runtime authority.'
  started_at: '2026-08-07T19:55:41+00:00'
---

## Transition note (2026-08-07T19:55:41+00:00)

Phase 1 release evidence retry executed.


## Transition note (2026-08-07T19:55:41+00:00)

Blocked on unresolved normative release evidence and release identity, detailed in the WorkItem description.


## Transition note (2026-08-10T16:29:15+00:00)

macOS arm64 host gate passed for 1.0.0-alpha.1 at run 20260810T161257Z-087ae53f636934f2a09a52e347264007; native/container version checks, Docker hardening, SBOM, zero-match Grype scan, and exact cleanup evidence retained.
