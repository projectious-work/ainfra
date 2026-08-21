---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_1330-MerryVale-use-a-fixed-external-cloudflare-fixture
  created: '2026-08-21T13:30:44+00:00'
spec:
  title: Use a fixed external Cloudflare fixture for Phase 9 certification
  state: accepted
  decision: Use an operator-managed dedicated Cloudflare Tunnel, fixed test hostname,
    DNS record, and Access policy as the external Phase 9 certification fixture. The
    automated disposable lifecycle will create and destroy only Hetzner and K3s resources,
    consume the external tunnel connector and Access credentials, test connectivity,
    and never require Cloudflare DNS edit permission.
  context: The user rejected domain-wide DNS edit access for a regression test and
    prefers to provision a fixed test subdomain. Phase 9 already treats Cloudflare
    resources as externally managed.
  rationale: This preserves least privilege, keeps the provider template ownership
    boundary accurate, permits repeatable tunnel regression tests, and avoids granting
    automation mutation rights across an important DNS zone.
  alternatives:
  - option: Per-run automated tunnel and DNS record
    reason: Rejected because DNS Edit is scoped to the domain rather than a single
      record.
  - option: TryCloudflare quick tunnel
    reason: Rejected because it does not exercise the remotely managed production
      tunnel/token design.
  - option: Manual DNS mutation for every run
    reason: Rejected because it prevents repeatable regression automation.
  consequences: 'The fixed Cloudflare fixture must be created once and maintained
    separately. Certification evidence must distinguish functional verification from
    ownership and teardown: Hetzner resources are independently proven absent, while
    the external Cloudflare fixture remains intentionally present. Tunnel and Access
    credentials must be supplied through protected environment or variable files and
    rotated independently.'
  related_workitems:
  - BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
  decided_at: '2026-08-21T13:30:44+00:00'
---
