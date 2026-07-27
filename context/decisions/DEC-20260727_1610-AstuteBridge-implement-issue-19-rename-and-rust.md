---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260727_1610-AstuteBridge-implement-issue-19-rename-and-rust
  created: '2026-07-27T16:10:47+00:00'
spec:
  title: Implement issue 19 rename and Rust CLI migration plan
  state: accepted
  decision: 'Execute the accepted implementation plan for issue #19, beginning with
    the bounded first milestone: repository rename and canonical URL repair, Python
    compatibility fixtures and destruction terminology correction, and a Rust CLI
    skeleton with dual-language local quality gates.'
  context: The overarching review approved renaming projectious-work/ainfra-templates
    to projectious-work/ainfra and rewriting the production CLI in Rust while preserving
    v1alpha1 contracts, safety properties, stable errors, templates, tests, and local-only
    release practices.
  rationale: The current Python implementation is mature enough to serve as a behavioral
    oracle but remains small enough for a fixture-driven rewrite. Completing rename
    and compatibility capture before lifecycle porting minimizes product and infrastructure
    recovery risk.
  consequences: Implementation proceeds in ordered milestones. Lifecycle code is not
    ported before compatibility fixtures exist. No GitHub Actions are introduced,
    runtime checkout dependence must be removed, and legacy infrastructure must retain
    a proven management or recovery path.
  decided_at: '2026-07-27T16:10:47+00:00'
---
