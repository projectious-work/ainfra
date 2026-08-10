---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260810_1447-AstuteFjord-bind-host-release-handovers-to-semver
  created: '2026-08-10T14:47:16+00:00'
spec:
  title: Bind host release handovers to SemVer
  state: accepted
  decision: Release-host preparation will require a strict SemVer, build all four
    supported binaries in the development container, and create an immutable version-scoped
    run. The host command will require the same version, resolve the newest concrete
    unused run for exactly that version, and perform host tests and evidence collection
    without compiling binaries.
  context: A generic latest run did not identify the intended release, while host-side
    preparation incorrectly required Go and crossed the container-to-host handover
    boundary.
  rationale: Version binding makes selection explicit and auditable. Container-side
    cross-compilation keeps build tooling inside the development environment, while
    host execution is limited to native and container validation. Concrete unique
    runs preserve evidence immutability.
  alternatives:
  - option: Generic latest run
    reason: Rejected because it can select a run for the wrong intended release.
  - option: Prepare and compile on the host
    reason: Rejected because compilation belongs inside the development container.
  consequences: Preparation and host execution become two commands in different environments.
    The shared workspace transfers the version-scoped run. Completed or attempted
    runs are retained; only incomplete preparation state may be cleaned safely. Go
    is no longer a host prerequisite.
  related_workitems:
  - BACK-20260807_1815-SolidSpark-complete-phase-1-release-evidence
  decided_at: '2026-08-10T14:47:16+00:00'
---
