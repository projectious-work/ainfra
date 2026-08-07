---
title: "Phase 0 shipped: the ainfra v1 specification"
date: 2026-08-07
description: >-
  The product contract for the ainfra v1 rewrite is accepted, integrated, and
  ready to guide implementation.
authors:
  - ainfra project
tags:
  - roadmap
  - specification
  - v1
toc: true
---

ainfra v1 has crossed its first roadmap boundary. Phase 0, **Product
specification**, is shipped. The specification has been reviewed, accepted, and
integrated into the v1 development line.

{{< pj-callout type="success" title="Phase 0 shipped" >}}
The product contract is in place. Phase 1, **Go project foundation**, is now in
progress.
{{< /pj-callout >}}

## What phase 0 established

The specification defines the intended product before implementation begins.
It covers:

- the v1 product boundary and lifecycle;
- native-file and template contracts;
- the security and secrets model;
- the Go architecture and source conventions;
- schemas, examples, and acceptance journeys;
- documentation and phase-note requirements; and
- the release-engineering process for the v1 line.

This gives implementation work a stable reference. Code, tests, templates, and
user documentation can now be reviewed against the same contract instead of
developing separate assumptions.

## What shipped means

`shipped` is the roadmap's completed state. A shipped phase has an aligned
specification, validation evidence, documentation, and development note.

Phase 0 delivered no runtime behavior. Its output is the normative contract
that later phases implement. The detailed record is available in the
[phase 0 development note](https://github.com/projectious-work/ainfra/blob/v1.x-dev/dev-notes/phase-00-product-specification.md).

## Next: the Go project foundation

Phase 1 establishes the Go module, command shell, typed results, process and
filesystem boundaries, test fixtures, developer tooling, and Linux/macOS build
targets.

Follow current progress on the [ainfra roadmap]({{< relref
"/docs/roadmap.md" >}}). The page is generated directly from the authoritative
roadmap YAML, so phase and status changes appear in the next documentation
build.
