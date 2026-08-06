---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260806_1605-BriskSun-use-pull-request-46-as-the
  created: '2026-08-06T16:05:28+00:00'
spec:
  title: 'Use pull request #46 as the sole v1 product specification'
  state: accepted
  decision: 'Treat GitHub pull request #46, “docs: specify ainfra v1 Go rewrite,”
    as the sole authoritative product specification for the v1 rewrite. Disregard
    all other product specifications, proposals, and drafts.'
  context: 'The project will undergo a complete v1 rewrite of code and documentation.
    The owner specified that the fully reviewed open GitHub pull request #46 is the
    only product specification to work against.'
  rationale: The reviewed PR provides one explicit, reviewable source of truth for
    the v1 product boundary, architecture, validation, and release expectations. Restricting
    scope to it prevents conflicting inputs during a full rewrite.
  alternatives:
  - option: 'Use existing v1 drafts or other product specifications alongside PR #46'
    rejected_because: The owner explicitly excluded all other sources from scope.
  - option: Create a new consolidated product specification before implementation
    rejected_because: It would duplicate the accepted reviewed PR and delay the rewrite
      without authorization.
  consequences: 'All v1 planning, implementation, documentation, and acceptance work
    must trace to PR #46. The PR currently targets v0.x-dev, so it must be rebased
    or retargeted to v1.x-dev before it becomes the baseline for the new version line.'
  decided_at: '2026-08-06T16:05:28+00:00'
---
