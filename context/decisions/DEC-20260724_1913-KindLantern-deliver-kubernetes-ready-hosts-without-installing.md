---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1913-KindLantern-deliver-kubernetes-ready-hosts-without-installing
  created: '2026-07-24T19:13:57+00:00'
spec:
  title: 'Deliver Kubernetes-ready hosts without installing Kubernetes in issue #1'
  state: accepted
  decision: The first Hetzner reference template will deliver and label only Kubernetes-ready
    infrastructure. It will not install or claim to operate a Kubernetes cluster.
    Kubernetes distribution installation must be separately scoped after the baseline
    passes security, idempotence, recovery, and ownership-scoped destroy verification.
  context: 'Issue #1 establishes secure infrastructure templates and a standardized
    target hand-off. Installing Kubernetes would introduce cluster lifecycle, CNI,
    certificates, kubeconfig handling, upgrades, and conformance responsibilities
    before the infrastructure baseline is proven.'
  rationale: This preserves the portfolio boundary, keeps the first topology concrete,
    and allows the project to validate infrastructure and host-hardening guarantees
    without absorbing workload-platform lifecycle responsibilities.
  alternatives:
  - option: 'Install a minimal Kubernetes distribution in issue #1'
    rejected_because: It expands scope into cluster lifecycle and credential management
      and would delay validation of the core infrastructure contract.
  - option: Describe unconfigured nodes as a Kubernetes cluster
    rejected_because: That would misrepresent the delivered capability and repeat
      an explicitly prohibited donor-project behavior.
  consequences: The output target type remains kubernetes-ready. Documentation must
    state what readiness guarantees and what remains absent. Kubeconfig references
    are not emitted until a separately scoped component actually creates them.
  decided_at: '2026-07-24T19:13:57+00:00'
---
