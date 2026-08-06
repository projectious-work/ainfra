---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260806_1623-ActiveGarden-adopt-a-long-lived-v1-version
  created: '2026-08-06T16:23:06+00:00'
spec:
  title: Adopt a long-lived v1 version-line branching and promotion model
  state: accepted
  decision: Create v1.x-release, v1.x-pre-release, and v1.x-dev from the same main
    commit as long-lived branches. Merge short-lived feat/* and ordinary pre-release
    fix/* branches into v1.x-dev. Promote tested candidates from v1.x-dev to v1.x-pre-release,
    then to v1.x-release, and finally to main. Tag alpha and beta prereleases on v1.x-pre-release,
    release candidates on v1.x-release, and the final stable SemVer version on the
    validated v1.x-release commit before merging it to main. Do not create v1.x-maintenance
    while v1 is the only active version line; create it from the latest stable v1
    tag only when v2 development begins.
  context: 'The project will be completely rewritten for v1 against the sole authoritative
    product specification in GitHub pull request #46. The project needs a version-line
    branching model parallel to v0.x while v1 is developed, prereleased, released,
    and supported.'
  rationale: This makes development, staged validation, release stabilization, and
    published history explicit without prematurely maintaining an additional support
    branch. It preserves a clear promotion path and prevents production incident fixes
    from unintentionally carrying unreleased v1 development work.
  alternatives:
  - option: Use a single v1.x-dev branch and release directly to main
    rejected_because: It does not provide separate staged validation and release-stabilization
      surfaces.
  - option: Create v1.x-maintenance immediately
    rejected_because: It duplicates v1.x-release while no newer major line is active.
  - option: Tag beta builds on v1.x-release
    rejected_because: Alpha and beta belong to prerelease validation; release candidates
      on the release branch more clearly signal final stabilization.
  consequences: Branch protection and CI should enforce promotion through the stated
    path. Production incident fixes after v1 general availability start from the current
    stable v1.x-release commit or tag, then are forwarded into v1.x-dev when applicable.
    The v1 product-spec PR must be rebased or retargeted from v0.x-dev to v1.x-dev
    before becoming the implementation baseline.
  decided_at: '2026-08-06T16:23:06+00:00'
---
