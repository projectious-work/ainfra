---
title: Acceptance matrix
weight: 40
description: Evidence expected for the initial secure template release.
---

| Requirement | Evidence |
|---|---|
| Versioned strict contracts | Schema and positive/negative fixture tests |
| Visible wrapper commands | CLI, runner, lifecycle, and adapter tests |
| No template hardcoding | Contract-driven discovery and fixtures |
| Pinned Hetzner template | Provider lock, image enum, validation |
| Secure SSH and keys | Schema, policy, plan, and host tests |
| Safe state | Backend capability validation |
| Secret-free output | Strict schema, redaction, Gitleaks |
| Local validation gates | `scripts/validate-all`, `scripts/test-all` |
| Disposable lifecycle | Approved live apply, idempotence, and teardown |
| Clear portfolio boundaries | README and architecture review |
| No unsafe donor behavior | Negative policy and plan tests |
| No GitHub workflows | Repository-policy test |

Every release criterion maps to an automated local check or documented manual
verification with sanitized evidence. A claim without one of those forms of
evidence remains unverified.
