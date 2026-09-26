---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260925_1532-CleverSpring-make-agent-ready-mcp-workflows-phase
  created: '2026-09-25T15:32:59+00:00'
  updated: '2026-09-25T15:34:42+00:00'
spec:
  title: Make agent-ready MCP workflows Phase 10 and the v1 beta gate
  state: accepted
  decision: Insert a planned Phase 10 for agent-ready MCP template authoring, native
    validation, guided progressive disclosure, and the independently authorized deployment/removal
    journey. Make Phase 10 completion the v1.0.0-beta.1 entry gate. Renumber current
    phases 10 through 22 as 11 through 23, retaining their idea status and content.
  context: The v1 roadmap marks Phase 9 shipped and places verifiable consumer handover
    at Phase 10, while the owner requires an AI agent to create and edit templates,
    use native OpenTofu and Ansible validation, and deploy or remove infrastructure
    through a useful MCP workflow before beta. The owner also requested progressive
    how-to guidance for using agents.
  rationale: This groups the remaining beta-critical MCP work into a reviewable vertical
    slice while allowing consumer handover, provider breadth, and later ecosystem
    ideas to follow beta. A small on-demand guide helps agents discover correct next
    steps without loading full documentation or granting authority.
  alternatives:
  - option: Keep the current Phase 10 number for consumer handover and add MCP as
      an unnumbered gate
    rejected_because: The MCP work has its own implementation and acceptance journey
      and should be visible in roadmap sequencing.
  - option: Require consumer handover and all later roadmap ideas before beta
    rejected_because: They are separable expansions that delay the owner's usable
      agent workflow.
  consequences: Amend the checked-in v1 roadmap, MCP specification, release engineering
    beta gate, and acceptance journeys. The guide remains read-only and progressively
    discloses version-matched topics and resources; it does not authorize or perform
    mutations. Existing Phase 10 consumer handover becomes Phase 11 and former Phase
    22 confidential computing becomes Phase 23. Supersede the prior beta-boundary
    decision while preserving Phase 9 as shipped.
  decided_at: '2026-09-25T15:32:59+00:00'
  supersedes: DEC-20260924_1509-RapidDove
  related_workitems:
  - BACK-20260925_1518-ValiantSummit
---
