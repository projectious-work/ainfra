---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_2005-WarmPond-license-ainfra-templates-under-mit
  created: '2026-07-24T20:05:10+00:00'
spec:
  title: License ainfra-templates under MIT
  state: accepted
  decision: License the ainfra-templates repository under the MIT License.
  context: 'GitHub issue #1 requires a LICENSE file for the reusable infrastructure
    template and wrapper repository, but did not choose a license.'
  rationale: The owner selected MIT. It is concise, permissive, and suitable for reuse
    of infrastructure templates and supporting tooling.
  alternatives:
  - option: Apache License 2.0
    rejected_because: The owner explicitly selected MIT instead.
  - option: Copyleft license
    rejected_because: The owner selected a permissive license without reciprocal distribution
      conditions.
  consequences: Milestone 0 must add the canonical MIT license text with the correct
    copyright holder and year. Source distributions and documentation must preserve
    the license notice.
  decided_at: '2026-07-24T20:05:10+00:00'
---
