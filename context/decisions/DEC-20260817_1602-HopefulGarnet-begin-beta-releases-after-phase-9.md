---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260817_1602-HopefulGarnet-begin-beta-releases-after-phase-9
  created: '2026-08-17T16:02:07+00:00'
spec:
  title: Begin beta releases after Phase 9 end-to-end acceptance
  state: accepted
  decision: Use completion of Phase 9 as the v1.0.0-beta.1 boundary. Phase 9 completion
    includes successful create, configure, convergence, and destroy execution against
    a real disposable environment with the certified template, plus closure of release-blocking
    Phase 0–9 conformance gaps.
  context: Phases 0–7 provide the guarded lifecycle and automation interface, Phase
    8 completes template authoring and conformance, and Phase 9 supplies the first
    certified Kubernetes-ready template. The release specification defines alpha,
    beta, RC, and stable mechanics but did not assign maturity transitions to roadmap
    phases.
  rationale: 'At this boundary the intended v1 product slice becomes feature-complete
    and externally testable: users can select a certified template, provide native
    inputs, review an immutable plan, receive a configured environment, verify convergence,
    and tear it down. Beta can then focus on field validation and defect closure instead
    of adding the missing core product experience.'
  alternatives:
  - option: Begin beta after Phase 8
    rejected_because: Authoring conformance alone does not prove the promised product
      through a certified real deployment.
  - option: Delay beta until Phase 10 or later
    rejected_because: Phases 10+ are currently idea-stage expansion and should be
      reviewed for content and order before becoming release blockers.
  - option: Keep issuing alpha releases without a phase boundary
    rejected_because: This leaves release maturity subjective and obscures when the
      v1 feature set is ready for external stabilization.
  consequences: Phase 8 and Phase 9 remain the critical path to beta. Phase 9 requires
    a disposable real-environment lifecycle acceptance gate. Beta releases are for
    external validation, compatibility work, and defect closure; material scope expansion
    requires explicit review. Phases 10+ remain non-blocking idea-stage roadmap entries
    until their content and ordering are reviewed. The roadmap or release specification
    should document the beta boundary.
  decided_at: '2026-08-17T16:02:07+00:00'
---
