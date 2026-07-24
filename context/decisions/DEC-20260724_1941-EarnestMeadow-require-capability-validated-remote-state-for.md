---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1941-EarnestMeadow-require-capability-validated-remote-state-for
  created: '2026-07-24T19:41:52+00:00'
spec:
  title: Require capability-validated remote state for non-disposable environments
  state: accepted
  decision: Require non-disposable environments to use a remote backend that demonstrably
    provides encryption at rest, state locking, versioning or equivalent recovery,
    TLS transport, and access control. Provide a backend-neutral contract and an S3-compatible
    reference configuration using native lockfiles when supported. Supply credentials
    only through environment or conventional credential files. Keep backend configuration
    outside version control. Permit local state only for an explicitly declared disposable-development
    environment with a prominent warning. The wrapper validates declared capabilities
    but neither provisions nor migrates backends. Hetzner Object Storage may be documented
    as a reference only after integration tests prove all required capabilities and
    OpenTofu endpoint compatibility.
  context: The ainfra template must protect potentially sensitive OpenTofu state without
    hardcoding one storage vendor or allowing the wrapper to become a state-management
    service. S3-compatible services differ in encryption, locking, versioning, and
    conditional-write behavior.
  rationale: This establishes enforceable state safety while avoiding a premature
    vendor abstraction. Capability testing prevents an S3-compatible API label from
    being mistaken for proof of correct locking or recovery behavior. Keeping credentials
    out of backend arguments reduces leakage into local OpenTofu metadata and saved
    plans.
  alternatives:
  - option: Mandate AWS S3 for all non-disposable environments
    rejected_because: It would impose a second cloud vendor and unnecessarily narrow
      deployment choices.
  - option: Assume Hetzner Object Storage is sufficient because it is S3-compatible
    rejected_because: API compatibility alone does not prove locking, encryption,
      version recovery, or conditional-write semantics.
  - option: Let every operator select any remote backend without validation
    rejected_because: The issue explicitly requires encrypted, locked remote state
      and machine-validatable conservative defaults.
  - option: Have the wrapper provision or automatically migrate state backends
    rejected_because: That would hide consequential state operations and exceed the
      thin-wrapper scope.
  consequences: Schemas and validation must represent environment disposability and
    backend capability requirements. The repository needs positive and negative backend
    fixtures plus integration evidence for each documented reference service. Local-state
    warnings must be unavoidable. State migration remains an explicit operator procedure.
  decided_at: '2026-07-24T19:41:51+00:00'
---
