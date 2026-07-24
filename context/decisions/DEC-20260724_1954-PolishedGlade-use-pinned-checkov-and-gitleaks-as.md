---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1954-PolishedGlade-use-pinned-checkov-and-gitleaks-as
  created: '2026-07-24T19:54:39+00:00'
spec:
  title: Use pinned Checkov and Gitleaks as complementary local security gates
  state: accepted
  decision: Use an exactly pinned Checkov version for OpenTofu and infrastructure-policy
    scanning and an exactly pinned Gitleaks version for secret scanning. Keep ansible-lint,
    OpenTofu tests, cloud-init schema validation, schema checks, and bespoke negative
    tests as separate gates. Store suppressions in version control only with a specific
    rationale. Fail on unsuppressed findings and on missing required scanners. Scan
    committed sources and relevant generated plans, inventory, standardized outputs,
    and captured logs.
  context: The repository needs local IaC misconfiguration and secret-leak detection
    while retaining explicit ainfra-specific negative tests. One general scanner should
    not be treated as sufficient proof for every security invariant.
  rationale: Checkov and Gitleaks provide focused, auditable coverage for different
    failure classes. Dedicated ainfra tests remain necessary for domain invariants
    such as omitted IPv4 allocation, private-key prohibition, management CIDRs, backend
    safety, and ownership-scoped destruction.
  alternatives:
  - option: Use only Checkov for both IaC and secrets
    rejected_because: Secret-leak verification would be less explicit and dependent
      on one scanner's coverage.
  - option: Use only a general vulnerability scanner
    rejected_because: The primary risks are IaC policy violations and secret disclosure,
      which benefit from focused tools.
  - option: Rely only on custom tests
    rejected_because: Custom invariants do not replace maintained broad rule sets
      and secret-pattern detection.
  consequences: Local bootstrap and validation scripts must install or verify the
    pinned versions. Generated artifacts need safe temporary handling and cleanup.
    Suppression files require review and tests must prove missing tools cannot be
    silently ignored.
  decided_at: '2026-07-24T19:54:39+00:00'
---
