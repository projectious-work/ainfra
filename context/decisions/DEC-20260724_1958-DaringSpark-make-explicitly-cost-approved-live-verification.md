---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1958-DaringSpark-make-explicitly-cost-approved-live-verification
  created: '2026-07-24T19:58:38+00:00'
spec:
  title: Make explicitly cost-approved live verification the final delivery gate
  state: accepted
  decision: Run live disposable-environment verification only as the final delivery
    step after every schema, unit, adapter, OpenTofu, cloud-init, Ansible, policy,
    secret, and documentation gate has passed locally. Require an explicit user approval
    immediately before creating resources, with a clear cost warning naming the Hetzner
    project, environment, template, requested resources, estimated maximum cost, and
    intended lifetime. The initial ceiling is one smallest supported x86 server, zero
    workers, no public IPv4, and an intended maximum lifetime of two hours. Credentials
    remain operator-owned and environment-referenced. Tag resources with a unique
    run ID, owner, creation time, and expiry. Generate a destroy plan after verification,
    but require separate explicit approval before destroy. Keep raw evidence ignored
    under .ainfra/evidence and promote only explicitly sanitized reports.
  context: The Hetzner baseline needs real apply, idempotence, evidence, and ownership-scoped
    destroy verification. These checks create billable cloud resources and therefore
    must not run as an ordinary local test or without the user's informed approval.
  rationale: Placing live verification last prevents cloud spend on changes that already
    fail local gates. Immediate explicit approval makes the cost and target visible
    to the person authorizing the run. Minimal resources and no public IPv4 reduce
    both cost and exposure while still proving the lifecycle.
  alternatives:
  - option: Run live verification as part of ordinary test-all or validate-all
    rejected_because: Routine local checks must not incur unannounced cloud costs
      or require cloud credentials.
  - option: Run live verification earlier in development
    rejected_because: That would spend money before cheaper local failures have been
      eliminated.
  - option: Infer approval from invoking a test script
    rejected_because: The user must explicitly approve the named billable resources
      and estimated cost immediately before creation.
  - option: Automatically destroy resources in a hidden cleanup hook
    rejected_because: Destruction must remain visible, ownership-scoped, and separately
      approved.
  consequences: The CLI needs a distinct final verification command and a two-stage
    approval flow for creation and destruction. Local gates must prove readiness before
    that command proceeds. Cost estimates are advisory rather than billing guarantees.
    Doctoring must report discoverable expired verification resources without deleting
    them.
  decided_at: '2026-07-24T19:58:38+00:00'
---
