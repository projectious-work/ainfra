---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260924_1509-RapidDove-require-an-end-to-end-mcp
  created: '2026-09-24T15:09:11+00:00'
  updated: '2026-09-25T15:33:02+00:00'
spec:
  title: Require an end-to-end MCP agent workflow before v1 beta
  state: superseded
  decision: Treat a demonstrated AI-agent MCP workflow for template creation, template
    management, native engine validation, reviewed plan/apply/configure, and exact
    reviewed destroy as a v1 beta entry criterion. Preserve independent authorization
    and the existing plan-bound lifecycle. Phases 10 and later are not beta prerequisites
    by default.
  context: Phase 9 and v1.0.0-alpha.9 shipped, but the current deployment-bound MCP
    server does not expose a complete path from authoring a new template through native
    OpenTofu and Ansible validation to independently authorized deployment and removal.
    The project owner explicitly identified useful MCP exposure of the template, OpenTofu,
    and Ansible tooling as necessary for a consumable beta.
  rationale: The existing Phase 7 MCP adapter handles a prepared deployment, but does
    not by itself support an agent creating and validating a new native template.
    A beta that cannot exercise the owner's required agent workflow would be functionally
    incomplete despite Phase 9 certification.
  alternatives:
  - option: Start beta after Phase 9 alone
    rejected_because: The documented MCP authoring path still depends on external
      file and native-tool operations, so the owner's required agent workflow is unproven.
  - option: Require all Phase 10+ roadmap ideas before beta
    rejected_because: Consumer handover, additional providers, catalog, coordination,
      and confidential-computing work are separable expansions and would delay a usable
      first beta.
  consequences: Define a narrow beta gate and acceptance test for an agent to author
    and validate a disposable template, manage its binding, review and execute an
    apply plan, configure, inspect, review and execute destroy, and verify teardown
    through MCP-accessible tools. The precise authoring API and approval issuer remain
    design choices; no generic shell or implicit infrastructure approval is authorized.
    The earlier Phase 9-only beta boundary is superseded.
  decided_at: '2026-09-24T15:09:11+00:00'
  supersedes: DEC-20260817_1602-HopefulGarnet
  related_workitems:
  - BACK-20260925_1518-ValiantSummit
  superseded_by: DEC-20260925_1532-CleverSpring
---
