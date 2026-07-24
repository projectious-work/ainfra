---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_2008-FocusedSea-use-version-specific-development-branches-and
  created: '2026-07-24T20:08:43+00:00'
spec:
  title: Use version-specific development branches and release from main
  state: accepted
  decision: Keep main release-only. For each planned minor release, create a version-specific
    development branch such as v0.1-dev from the latest released main. Branch feature/<name>
    branches from the active version development branch and merge them back through
    review, using squash merges. When the release is ready and all local gates pass,
    merge the version development branch to main through a release PR and create an
    annotated semantic-version tag such as v0.1.0 from the exact release commit. Create
    the next version development branch from the newly released main and retire the
    previous one. Branch unreleased fixes from the active development branch. Branch
    emergency fixes for released code from main, release them as patch versions, and
    propagate them back into the active development branch. Do not commit features
    directly to main or the active development branch.
  context: The repository needs a branch model that keeps main release-only while
    allowing feature integration and explicit pre-1.0 release lines.
  rationale: Version-specific development branches make the intended release line
    explicit and avoid an ambiguous long-lived v0.x-dev branch. A release-only main
    provides a clear deployable history, while feature branches and a reviewed integration
    branch support incremental work. Propagating hotfixes prevents released fixes
    from being lost in later versions.
  alternatives:
  - option: Use one v0.x-dev branch for every pre-1.0 release
    rejected_because: Its target version becomes ambiguous and it accumulates history
      across multiple release lines.
  - option: Use trunk-based development directly on main
    rejected_because: The owner wants main to contain releases only and explicit release
      promotion.
  - option: Create feature branches directly from main
    rejected_because: Features for the upcoming release need to integrate and validate
      together before promotion to release-only main.
  consequences: Branch protection and documentation must distinguish release, development,
    feature, and hotfix flows. Local validation runs on feature integration and again
    before the release PR. Release scripts must tag the exact main commit after merge.
    Each new minor line requires creating and later retiring its own development branch.
  related_workitems:
  - BACK-20260724_2005-FastWren-deliver-secure-ainfra-templates
  decided_at: '2026-07-24T20:08:43+00:00'
---
