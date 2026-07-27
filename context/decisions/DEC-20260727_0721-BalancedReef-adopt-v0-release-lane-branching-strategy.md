---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260727_0721-BalancedReef-adopt-v0-release-lane-branching-strategy
  created: '2026-07-27T07:21:31+00:00'
spec:
  title: Adopt v0 release-lane branching strategy
  state: accepted
  decision: Use main as the default branch and latest stable lane. Use v0.x-maintenance
    for supported v0 fixes, v0.x-dev for normal integration, v0.x-pre-release for
    alpha/beta/RC validation, and v0.x-release for general-availability release preparation.
    Promote tested commits forward through pull requests and clean obsolete branches.
  context: ainfra-templates currently has no v1 line. The user requested adoption
    of ProcessKit's branching model adapted to v0 only, restoration of main as default,
    and cleanup of other branches.
  rationale: Separate stable, maintenance, development, prerelease, and release responsibilities
    while preserving an explicit promotion path and immutable published tags.
  alternatives:
  - option: Keep v0.1-dev as the default branch
    reason_rejected: It conflates integration and repository default/stable state.
  - option: Use only main and feature branches
    reason_rejected: It lacks isolated maintenance and release-validation lanes.
  consequences: Feature branches target v0.x-dev. Prerelease tags originate only from
    v0.x-pre-release. Stable v0 tags originate only from v0.x-release and are then
    merged to main. Direct development on release-integration branches is avoided.
  decided_at: '2026-07-27T07:21:31+00:00'
---
