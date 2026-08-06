## Versioning

ainfra uses Semantic Versioning. Contract versions evolve independently and
appear in their `apiVersion` or schema-version fields.

- patch: compatible fixes and documentation corrections;
- minor: compatible functionality and template/source capabilities;
- major: breaking CLI or contract behavior;
- prerelease: `v1.0.0-alpha.1`, `-beta.1`, and `-rc.1`.

## Branching and promotion

ainfra follows the projectious.work branching and release-promotion standard.
While the v1 rewrite proceeds beside the stable product, its long-lived
branches are:

| Branch | Purpose |
|---|---|
| `v1.x-dev` | Reviewed integration of short-lived `feat/*` and `fix/*` branches. |
| `v1.x-pre-release` | Exact alpha/beta staging pointer. |
| `v1.x-release` | Exact release-candidate and final-release pointer. |
| `main` | Published stable history. |

The three v1 branches begin at the same stable `main` commit. Topic branches
merge only into `v1.x-dev`, using the repository's normal reviewed integration
policy. The pre-release and release branches receive no direct commits and no
topic-branch merges.

Promotions MUST preserve commit identity. An approved source tip advances the
next branch only by a fast-forward equivalent to `git merge --ff-only`:

```text
stable baseline on main
        │
        ├── v1.x-dev ── feat/* and fix/* integration
        │       │
        │       └── fast-forward ──> v1.x-pre-release
        │                                  │
        │                                  └── alpha / beta tags
        │                                           │
        │                       fast-forward ──> v1.x-release
        │                                                   │
        │                                                   └── rc / final tags
        │                                                           │
        └────────── fast-forward final commit into main <───────────┘
```

Squash, rebase, and merge commits are prohibited during promotion. If a
fast-forward is impossible, the release stops; divergence is reconciled on
`v1.x-dev` and the affected validation is repeated. Force-pushing a long-lived
branch or moving a published tag is prohibited.

The first alpha/beta promotion starts feature freeze for that release train.
Until the final commit reaches `main`, `v1.x-dev` accepts only approved release-
scope fixes, tests, documentation, and release preparation. A stabilization
fix branches from `v1.x-dev`, returns there through review, and is re-promoted
through every applicable stage.

Alpha and beta tags are created on validated `v1.x-pre-release` commits. RC and
stable tags are created on validated `v1.x-release` commits. If tracked content
changes after an RC, the result receives a higher RC and repeats validation.
The stable tag, `v1.x-release`, and `main` MUST resolve to the same final commit.

For an incident in the current stable major, `fix/<topic>` starts from the
affected stable tag, not from development. After the patch reaches `main`,
`main` is merged into the same-major development line to retain ancestry, and
the fix is separately forward-ported to affected newer majors. A
`vX.x-maintenance` branch is created from the latest stable vX tag only when a
newer major is stable and major X remains explicitly supported.

## Changelog

`CHANGELOG.md` follows Keep a Changelog structure:

```markdown
## [Unreleased]

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
```

Every release moves entries into a dated version section. Conventional Commits
MAY support automation, but a curated changelog is authoritative. Breaking
changes include explicit migration instructions.

## Release targets and artifacts

Required binary archives:

```text
ainfra_<version>_linux_amd64.tar.gz
ainfra_<version>_linux_arm64.tar.gz
ainfra_<version>_darwin_amd64.tar.gz
ainfra_<version>_darwin_arm64.tar.gz
```

Also publish:

- source archive/reference to signed tag;
- SHA-256 checksums;
- SBOMs;
- signatures or attestations according to release infrastructure;
- release notes derived from the curated changelog;
- installation and verification instructions.

Do not publish Windows binaries or an ainfra container image. The repository
Dockerfile is source and is validated, not published as an image artifact.

## Pre-release gate

1. Confirm intended version, clean release branch, and accepted scope.
2. Review all changes since the previous tag.
3. Run formatting, vet, static analysis, and curated linting.
4. Run unit, race, coverage, fuzz-smoke, process-contract, and local-engine
   integration tests.
5. Run `govulncheck`, `gosec`, OSV, and secret scans.
6. Review direct/transitive dependencies and licenses.
7. Validate schemas and every example.
8. Build user documentation, check links, and verify CLI reference drift.
9. Validate template-authoring instructions with a clean-room exercise or
   fixture.
10. Build and lint the optional Dockerfile; scan the locally built image, then
    discard it.
11. Cross-build all four supported targets.
12. Smoke-test archives on representative Linux/macOS systems.
13. Generate checksums and SBOMs; scan final artifacts.
14. Verify changelog, migration notes, phase notes, supported versions, and
    security docs against the implemented behavior.
15. Run disposable live acceptance when provider/template behavior changed.
16. Create signed tag and release from the exact validated commit.
17. Download published artifacts and independently verify checksums,
    signatures, `ainfra version`, and `ainfra help`.

## Documentation release checklist

Review and update when applicable:

- README status and installation;
- prerequisites and supported tool versions;
- CLI reference and machine interface;
- deployment/template/lock contracts;
- OpenTofu, output/inventory, and Ansible integration;
- security and recovery;
- template-authoring and AI-agent guidance;
- optional Dockerfile usage;
- reference template and live evidence;
- compatibility/migration;
- implementation phase notes and roadmap status;
- specification status/version;
- changelog and release notes.

## Release standard

- **AINFRA-REL-001:** no release is built from a dirty worktree.
- **AINFRA-REL-002:** artifacts MUST be derived from the signed release tag.
- **AINFRA-REL-003:** release checks MUST fail closed on errors; documented
  warnings require explicit review.
- **AINFRA-REL-004:** published checksums, SBOMs, and signatures MUST be
  verified after publication.
- **AINFRA-REL-005:** a release MUST not claim provider support beyond current
  redacted acceptance evidence.
- **AINFRA-REL-006:** security fixes MUST use the changelog Security section and
  coordinated disclosure where appropriate.
- **AINFRA-REL-007:** promotion branches MUST be exact fast-forward pointers
  and MUST NOT contain unique commits.
- **AINFRA-REL-008:** release tags MUST be annotated, SHOULD be signed, and
  MUST NOT be moved or reused after publication.
- **AINFRA-REL-009:** agents MUST identify the active version line, source,
  target, intended version, and operation type before changing branches or
  opening a pull request.
