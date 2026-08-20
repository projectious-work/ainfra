---
title: "Phase 0 shipped: the ainfra v1 specification"
date: 2026-08-07
description: >-
  The product contract for the ainfra v1 rewrite is accepted, integrated, and
  ready to guide implementation.
badges:
  - label: Phase 0
  - label: Specification
    variant: accent
tags:
  - roadmap
  - specification
  - v1
toc: true
---

ainfra v1 has crossed its first roadmap boundary. Phase 0, **Product
specification**, is shipped. The specification has been reviewed, accepted, and
integrated into the v1 development line.

{{< callout type="success" title="Phase 0 shipped" >}}
The product contract is in place. Phase 1, **Go project foundation**, is the
next implementation slice.
{{< /callout >}}

## What Phase 0 established

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
[Phase 0 development note](https://github.com/projectious-work/ainfra/blob/v1.x-dev/dev-notes/phase-00-product-specification.md).

Follow current progress on the [ainfra roadmap]({{< relref
"/docs/roadmap.md" >}}). It is generated directly from the authoritative
roadmap YAML.
