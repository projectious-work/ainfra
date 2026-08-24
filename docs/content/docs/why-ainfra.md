---
title: Why ainfra exists
weight: 15
---

AI agents can generate OpenTofu and Ansible and call their command lines. Why
put ainfra between an agent and those tools?

## The short answer

ainfra is the agent-native infrastructure execution boundary for reviewed,
reproducible OpenTofu and Ansible deployments.

Agents do the work where interpretation and synthesis help: understand intent,
select or create a template, explain a plan, and propose remediation. ainfra
does the work where stable controls matter: validate contracts, lock template
content, bind a reviewed plan, verify approval, scope credentials, execute
native tools, retain evidence, and recover after interruption.

ainfra does not make cloud infrastructure itself deterministic. Providers,
networks, and remote machines remain stateful and fallible. It makes the
process by which an agent or human changes them explicit, reviewable,
reproducible, attributable, and recoverable.

## Why not ask an agent to deploy directly?

A direct request such as “deploy a secure Kubernetes cluster” does not fully
describe network topology, cost limits, ingress, credentials, state and
locking, provider versions, ownership, approval, evidence retention, recovery,
or teardown. Those constraints still exist if they are moved into prompts and
Markdown. They become harder to validate, version, review, and repeat.

Infrastructure also differs from a one-off informational answer. A deployment
creates persistent state, spends money, uses privileged authority, changes
external systems, and may fail ambiguously. Several people and agents may
operate it over time. The execution boundary therefore deserves a stricter
contract than the reasoning that proposed the change.

## Division of labour

| Activity | Best mode | Boundary |
|---|---|---|
| Understand the desired outcome | Agent reasoning | A proposal has no mutation authority. |
| Select or author a template | Agent-assisted | Standard OpenTofu and Ansible plus the ainfra template contract. |
| Capture company and security standards | Agent-assisted | Accepted knowledge becomes versioned templates, policy, schemas, or approvals. |
| Validate source and inputs | Deterministic tooling | Strict contracts and stable diagnostics. |
| Create and explain a plan | Deterministic plan, agent explanation | The saved plan remains authoritative. |
| Approve a change | Human or independent policy | Exact, attributable, expiring authorization. |
| Acquire credentials | Deterministic security mechanism | External provider, least exposure, redaction, and cleanup. |
| Apply, verify, or destroy | Deterministic execution protocol | Reviewed inputs, native tools, locking, evidence, and recovery. |
| Explain failure and suggest a fix | Agent reasoning | Facts come from typed evidence; another mutation needs authorization. |

Templates are compiled institutional knowledge. They preserve provider
experience, company standards, security posture, access, lifecycle, and
teardown in inspectable native artifacts. An agent can author them quickly;
ainfra validates and executes an immutable reviewed revision repeatedly.

## MCP first, not MCP only

MCP is the primary agent interaction surface. It lets an agent discover
capabilities, inspect a deployment, create a reviewed plan, execute an
independently authorized operation, and retrieve typed status and recovery
guidance without wrapping shell commands or parsing terminal prose.

The CLI remains a complete interface for humans, CI, recovery, and break-glass
operation. MCP and CLI call the same application use cases and preserve the
same lifecycle and security rules. Native OpenTofu and Ansible commands remain
the final diagnosis and recovery escape path.

ainfra is therefore not an MCP server with a convenience CLI. MCP is the
leading agent interface over a protocol-independent execution core.

## What ainfra deliberately does not do

ainfra does not:

- accept arbitrary natural-language desired state as executable authority;
- provide a generic shell tool to agents;
- invent a language above OpenTofu and Ansible;
- treat a conversational “approved” as an authorization artifact;
- hide native plans, variables, state ownership, or recovery paths; or
- turn an agent session into an infrastructure control plane.

This boundary combines the speed of agent-assisted creation with the controls
expected when privileged software changes production infrastructure.

## Further context

Andrej Karpathy's
[Software Is Changing (Again)](https://www.youtube.com/watch?v=LCEmiRjPEtQ)
describes natural language and models as a new software interface. ainfra
adopts that interface while retaining explicit execution controls for
persistent privileged operations. The industry is moving similarly:
HashiCorp's
[Terraform MCP server](https://developer.hashicorp.com/terraform/mcp-server)
connects agents to infrastructure capabilities, while its
[security model](https://developer.hashicorp.com/terraform/mcp-server/security)
still requires validation and addresses hallucination, prompt injection, tool
poisoning, rug pulls, and tool shadowing.
