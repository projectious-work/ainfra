---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260730_0705-ZestfulRose-replace-docsy-with-hextra-for-the
  created: '2026-07-30T07:05:20+00:00'
spec:
  title: Replace Docsy with Hextra for the documentation site
  state: accepted
  decision: Adopt Hextra as the documentation theme and remove Docsy completely. Commit
    the current Hextra migration as one coherent change, then continue content and
    design work on the Hextra site.
  context: The existing documentation site used Docsy. The Hextra migration is implemented
    and intentionally changes the theme submodule, configuration, assets, layouts,
    scripts, and documentation content.
  rationale: Hextra is the selected foundation for the upcoming documentation redesign.
    Keeping both themes would preserve obsolete dependencies and split maintenance
    across two incompatible site structures.
  alternatives:
  - option: Retain Docsy alongside Hextra
    rejected_because: Creates duplicate theme dependencies and ambiguity about the
      supported documentation stack.
  - option: Continue improving the Docsy site
    rejected_because: Conflicts with the confirmed plan to build future documentation
      content and design on Hextra.
  consequences: Docsy files, npm tooling, and theme references are removed. Hextra
    becomes the sole supported documentation theme; future content and visual work
    must target its structure and extension points.
  decided_at: '2026-07-30T07:05:20+00:00'
---
