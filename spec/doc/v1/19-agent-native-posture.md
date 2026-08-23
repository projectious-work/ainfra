# Agent-native product posture

## Why ainfra exists in an agentic software world

An AI agent can already generate OpenTofu and Ansible and invoke their command
lines. That capability does not remove the need for an infrastructure product.
Infrastructure operations retain privileged credentials, persistent state,
financial and destructive consequences, delayed external effects, approval
boundaries, organizational policy, recovery obligations, and multiple actors
over time.

Moving those constraints into an unconstrained prompt would recreate software
as ambiguous prose. ainfra instead gives agents and humans a small, versioned,
testable execution boundary built from standard tools. It does not replace
agent reasoning or provider knowledge.

The product posture is:

> ainfra is the agent-native infrastructure execution boundary for reviewed,
> reproducible OpenTofu and Ansible deployments.

The precise promise is not that cloud infrastructure itself is deterministic.
Provider APIs, networks, and remote systems remain stateful and fallible.
ainfra makes the process by which an actor changes them explicit, reviewable,
reproducible, attributable, and recoverable.

- **AINFRA-AGENT-001:** ainfra MUST treat natural-language interpretation as
  an upstream source of intent, never as mutation authority or execution
  truth.
- **AINFRA-AGENT-002:** the product MUST preserve a versioned, deterministic
  control protocol around validation, plan binding, authorization, credential
  delivery, execution, recovery, and evidence.
- **AINFRA-AGENT-003:** ainfra MUST NOT accept arbitrary natural-language
  desired state, shell commands, hooks, or executable arguments as a shortcut
  around templates and native engine contracts.

## Division of labour

| Activity | Primary mode | Required boundary |
|---|---|---|
| Interpret the operator's goal | Probabilistic agent reasoning | Proposed intent is inspectable and has no authority by itself. |
| Select, create, or adapt a template | Agent-assisted authoring | Native OpenTofu and Ansible plus the versioned ainfra template contract. |
| Interpret company and security standards | Agent-assisted reasoning | Accepted requirements become versioned templates, policy, schemas, or approvals rather than hidden prompt state. |
| Validate a template and deployment | Deterministic tooling | Schemas, conformance checks, native validation, and stable diagnostics. |
| Produce and explain a change proposal | Deterministic plan with agent explanation | The saved plan is authoritative; explanation is derived and non-authorizing. |
| Approve a mutation | Human or independently authorized policy | Approval is explicit, attributable, time-bound, and bound to the exact operation and plan. |
| Acquire and deliver credentials | Deterministic security mechanism | External provider, least exposure, phase scoping, redaction, and cleanup. |
| Apply, configure, verify, or destroy | Deterministic execution protocol | Exact reviewed inputs, native engines, locking, evidence, and recovery. |
| Explain failure and propose remediation | Agent-assisted reasoning | Facts come from typed diagnostics and sanitized evidence; another mutation requires fresh authorization. |

Templates are compiled institutional knowledge: they capture provider
experience, company standards, security posture, lifecycle semantics, access,
and teardown in reusable native artifacts. Agents can create them quickly, but
deployment uses an immutable validated revision. This converts probabilistic
synthesis into governed operational capability.

## Product and interface identity

The internal application use cases are the durable product boundary. MCP and
CLI are equally authoritative adapters over that boundary:

- MCP is the primary agent interaction surface;
- CLI is the primary direct human, CI, recovery, and break-glass surface; and
- native OpenTofu and Ansible commands remain the documented diagnosis and
  recovery escape path.

The external posture is MCP-first, but ainfra is not "an MCP server with a
convenience CLI." MCP is a replaceable protocol integration. No protocol type,
session, conversational state, or tool annotation may enter domain semantics.

- **AINFRA-AGENT-010:** MCP and CLI MUST expose the same application use cases,
  lifecycle invariants, authorization rules, result semantics, and recovery
  state wherever their capabilities overlap.
- **AINFRA-AGENT-011:** an agent MUST be able to discover the permitted next
  action from typed capabilities, resources, results, and diagnostics without
  parsing terminal prose or repository implementation source.
- **AINFRA-AGENT-012:** losing an MCP session MUST NOT lose, ambiguously repeat,
  or silently authorize a mutation; durable operation identity and status
  remain outside transport-session state.
- **AINFRA-AGENT-013:** CLI operation MUST remain complete enough to validate,
  authorize, execute, inspect, recover, and destroy without an MCP client.

## Agent experience contract

The agent-facing surface SHOULD organize tools around bounded use cases, not
command execution. Its conceptual vocabulary includes:

- inspect templates, deployments, schemas, capabilities, and requirements;
- diagnose the environment and deployment;
- validate and lock a template revision;
- create, retrieve, and explain an apply or destroy plan;
- submit independent authorization for an exact reviewed operation;
- execute an authorized plan;
- inspect operation status, sanitized result, evidence, and recovery advice;
  and
- configure and verify hosts when applicable.

Exact MCP tool names and schema versions are implementation contracts. There
MUST NOT be an arbitrary `run_command`, `execute_shell`, or `deploy_text` tool.
Tool discovery is an authorization boundary: read-only is the default and
mutating capability groups require explicit server configuration.

MCP resources SHOULD expose version-matched:

- product and template schemas;
- the template-authoring contract and supported versions;
- template metadata and documentation;
- credential-provider capabilities without secret values;
- sanitized deployment status and result contracts;
- reviewed plan summaries; and
- sanitized run evidence and recovery guidance.

- **AINFRA-AGENT-020:** every tool MUST represent one bounded application use
  case with a strict input schema, stable outcome codes, and explicit side-
  effect classification.
- **AINFRA-AGENT-021:** every result MUST identify applicability, blocking and
  advisory findings, approval requirements, durable operation identity,
  permitted next actions, and recovery state where relevant.
- **AINFRA-AGENT-022:** agent-facing resources MUST be version-aligned with the
  running binary and MUST NOT supply executable instructions from an untrusted
  template as server or tool policy.
- **AINFRA-AGENT-023:** plans and evidence are factual inputs to explanation;
  generated summaries MUST NOT replace, modify, or authorize their source.

## Authorization and actor separation

ainfra distinguishes the actor that authored a template, requested a plan,
approved a plan, supplied deployment authority, and executed the operation.
One identity MAY occupy several roles only when explicit policy permits it.
Conversation text such as "approved" is not an approval artifact.

Production-capable workflows SHOULD support separate plan and apply identities
or an independent policy decision. Authorization binds the canonical project,
operation, plan digest, caller, intent, expiry, and permitted credential
authority. Agent confidence and MCP client confirmation UI are not substitutes
for server-side verification.

- **AINFRA-AGENT-030:** an agent that creates a plan MUST NOT self-authorize
  its execution unless an independently configured policy explicitly grants
  that identity both roles.
- **AINFRA-AGENT-031:** every mutation MUST remain attributable across intent,
  plan, approval, credential authority, execution, and evidence without
  recording secret values or private chain-of-thought.

## Remote and protocol evolution

Local stdio remains the initial MCP transport. A future remote MCP service is
a distinct security and operational phase requiring authenticated identities,
TLS, authorization, rate and concurrency limits, durable operation lookup,
credential-provider integration, deployment and template allowlists, tenant
isolation where applicable, centralized audit evidence, and an explicit remote
deployment threat model.

Remote MCP MUST NOT be implemented by exposing the local stdio assumptions on
a network socket. The application and result contracts remain transport
independent so a future protocol can replace or complement MCP without a new
execution lifecycle.
