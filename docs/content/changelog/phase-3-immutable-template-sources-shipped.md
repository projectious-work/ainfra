---
title: "Phase 3 shipped: Immutable template sources"
date: 2026-08-13T12:00:00Z
description: >-
  Local and Git-subdirectory template sources can be locked, materialized,
  verified by digest, and checked for drift.
badges:
  - label: Phase 3
  - label: v1.0.0-alpha.3
    variant: accent
tags: [roadmap, release, v1]
toc: true
---

Phase 3 introduced the immutable template boundary consumed by every lifecycle
operation.

## What shipped

- Local and Git-subdirectory source resolution.
- Canonical lock records with immutable revisions and tree digests.
- Contained, private template materialization and verified cache reuse.
- Explicit source updates with visible changes.
- Deterministic source, lock, digest, and cache drift diagnostics.

Read the complete
[Phase 3 development note](https://github.com/projectious-work/ainfra/blob/v1.x-dev/dev-notes/phase-03-immutable-template-sources.md)
or inspect
[v1.0.0-alpha.3](https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.3).
