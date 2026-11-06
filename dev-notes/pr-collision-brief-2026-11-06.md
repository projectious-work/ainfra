# ainfra open PRs: collision resolution brief (2026-11-06)

**Audience:** the ainfra development container agent
**Owner rule:** where PRs collide, the newer PR takes precedence over the older one.
**Scope:** ainfra PRs #108–#112, all targeting `v1.x-dev` (head `bf16b0ae`, unchanged since 2026-08-23).
**Method:** pairwise `git merge-tree` from the coordination container, read-only. Nothing was pushed, and no ainfra scripts were run.

## Precedence (newest first)

| Rank | PR | Opened | Topic |
|---|---|---|---|
| 1 | #112 | 2026-10-26 | Adopt company engineering practice in the v1 spec |
| 2 | #111 | 2026-09-30 | Separate done roadmap phases from shipped releases |
| 3 | #110 | 2026-09-26 | Phase 10 agent-ready MCP beta gate |
| 4 | #109 | 2026-08-24 | Managed-container boundary (draft, stacked on #108) |
| 5 | #108 | 2026-08-23 | Agent-native execution and credential hardening (draft) |

## Collision matrix

| Pair | Result |
|---|---|
| #110 × #111 × #112 | Clean, both in pairs and merged together in sequence |
| #108/#109 × #111 | Clean |
| #108/#109 × #112 | Text conflict in `spec/doc/v1/README.md` |
| #108/#109 × #110 | Text conflicts in `spec/doc/v1/roadmap.yaml` (13 hunks) and `spec/doc/v1/15-mcp-server-mode.md`, plus a semantic clash over phase numbering |

Only #108/#109 have to adapt. #110, #111 and #112 can merge as they are.

## Merge order

1. #111 (no rebase needed).
2. #112 (merges cleanly after #111; rebase only if the branch protection requires it).
3. The processkit-removal PR, then #110 reduced to its spec and roadmap part (see owner decision 1). The reduced #110 still merges cleanly after #111 and #112; the archive move touches only `AGENTS.md` and the `context/` references in `spec/doc/v1/10-release-engineering.md`, which #110 also edits, so rebase #110 after the removal PR.
4. Fold #109 into #108, then rebase #108 onto the resulting `v1.x-dev` and resolve as described below. Close #109 with a link to #108.

## Resolving #108 (with #109 folded in)

### 1. Roadmap numbering (`spec/doc/v1/roadmap.yaml`)

#110's Phase 10, **Agent-ready MCP workflows**, keeps number 10 and stays the `v1.0.0-beta.1` gate. #108's inserted phases follow it. Every #108/#109 phase number N ≥ 10 becomes N + 1:

| Final | Phase | Source | Status |
|---|---|---|---|
| 10 | Agent-ready MCP workflows | #110 | planned |
| 11 | Agent-native product and interface hardening | #108 (was 10) | planned |
| 12 | Deployment credential acquisition and delivery hardening | #108 (was 11) | planned |
| 13 | Verifiable consumer target handover | base (#110: 11, #108: 12) | planned (#108's deliberate change, which #110 did not contest) |
| 14 | Complete major-cloud templates | base | idea |
| 15 | Specialized compute infrastructure templates | #109 title | idea |
| 16–24 | Compatibility and migration … Exchangeable infrastructure engine | base, unchanged order | idea |
| 25 | Additional deployment credential providers | #108 (was 24) | idea |
| 26 | Confidential computing and live attestation | base (#110: 23, #108: 25) | idea |

Also renumber every `dependsOn` reference, phase grouping and dev-note link, and keep #111's status schema (release metadata for `shipped`, dev notes for `done`/`shipped`).

### 2. Phase references in prose

The same rule (#108/#109 phase N ≥ 10 → N + 1) applies to every reference #108 adds outside the roadmap:
`01`, `03`, `04`, `05`, `08`, `11`, `12`, `13`, `15`, `18-deployment-credentials.md` and `19-agent-native-posture.md`.
Examples: "Roadmap Phase 10 additionally requires…" (hardening) becomes Phase 11, "Phase 11 supports these sources" becomes Phase 12, "Phase 24 adds providers" becomes Phase 25, and "part of Phase 25" becomes Phase 26.
#110's prose only references Phase 10, which stays correct.

### 3. `spec/doc/v1/15-mcp-server-mode.md`

- **Introduction:** keep #110's paragraph (Phase 7 introduced MCP; Phase 10 extends it to agent-ready authoring before the first beta). Append #108's point that the CLI stays complete and behaviorally equivalent for direct human, CI, recovery and break-glass operation, attributing the hardening to **Phase 11**.
- **Sections:** keep #110's "Phase 10 agent authoring and progressive guidance" section unchanged. Keep #108's "Agent experience" section after it. Where #108 overlaps #110 (version-matched resources, authoring guidance, labeling untrusted template content), #110's wording and requirements win; #108 adds only what #110 does not cover: durable operation identity, permitted next actions, actor and authorization separation, credential-provider capability resources and the adversarial suite.

### 4. `spec/doc/v1/README.md`

Both changes are additions to the list of decisions and standards. Keep #112's entry for the company engineering standards (revision 2026-10-26), followed by #108's BrightSail and RoyalShore entries.

### 5. Requirement IDs

There is one real ID collision. #110 adds `AINFRA-MCP-011` to `AINFRA-MCP-015`, and #108 adds `AINFRA-MCP-011` to `AINFRA-MCP-018` with different content. #110 keeps its IDs, and #108's MCP requirements become `AINFRA-MCP-016` to `AINFRA-MCP-023`, keeping their order. Update every cross-reference to them, including acceptance and test mappings.

Checked and free of collisions: #112's IDs (`DEV-013`–`018`, `DOC-014`, `REL-012`/`013`, `TEST-019`, `CODE-012`/`013`) and #108's other new families (`AGENT-*`, `CRED-*`, `STACK-*`). #108 edits `TEST-014`–`016`, which already exist in the base, without redefining them.

## Verification after each merge

- `scripts/validate-v1-spec`
- `git diff --check`
- Before #108 leaves draft: search the whole repo for `Phase [0-9]+` and check each hit against the final table.

## Owner decisions (2026-11-06)

1. **ainfra gives up processkit.** This follows the coordination repository, which removed it on 2026-10-29 (`DEC-20261029_1040-AgileFalcon`). Consequences:
   - Split #110 before merging. Keep its spec, roadmap and dev-note files (`spec/doc/v1/10-release-engineering.md`, `11-acceptance-migration.md`, `15-mcp-server-mode.md`, `roadmap.yaml`, `dev-notes/phase-09-initial-production-template.md`). Drop the processkit `context/` records and `.pk-doctor-allowlist.toml` from it. Move the aibox sync files (`aibox.toml`, `aibox.lock`, `context/templates/aibox-home/` snapshots) to a separate chore PR, or leave them out if the processkit removal makes them unnecessary.
   - Open a separate processkit-removal PR on `v1.x-dev`, modelled on the coordination repository: tag the last state `processkit-final`, move the entity records to `archive/processkit/` as plain read-only files, remove the processkit MCP server, skills, templates and pk-* commands, and rewrite links from `context/` to the archive.
   - Keep #110's decision records (`HopefulGarnet`, `RapidDove`, `CleverSpring`) in that archive, so the Phase 10 rationale stays traceable.
2. **Agent-native hardening after beta is acceptable.** #108's hardening stays Phase 11, after the Phase 10 `v1.0.0-beta.1` gate. Neither #110's beta gate nor #108's phase content needs to change.
