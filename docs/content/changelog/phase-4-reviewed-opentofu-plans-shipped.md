---
title: "Phase 4 shipped: Reviewed OpenTofu plans"
date: 2026-08-14
description: >-
  OpenTofu changes can be planned, reviewed, bound to immutable inputs, and
  applied only from the exact saved plan.
badges:
  - label: Phase 4
  - label: v1.0.0-alpha.4
    variant: accent
tags: [roadmap, release, v1]
toc: true
---

Phase 4 delivered the first guarded infrastructure mutation workflow.

## What shipped

- Controlled OpenTofu initialization and planning.
- Sanitized structural summaries for human review.
- Immutable bindings across template, inputs, tools, and saved-plan bytes.
- Exact plan-ID application without implicit replanning.
- Durable evidence for started, terminal, cancelled, and ambiguous outcomes.

Read the complete
[Phase 4 development note](https://github.com/projectious-work/ainfra/blob/v1.x-dev/dev-notes/phase-04-reviewed-opentofu-plans.md)
or inspect
[v1.0.0-alpha.4](https://github.com/projectious-work/ainfra/releases/tag/v1.0.0-alpha.4).
