## Goals

ainfra is an agent-native infrastructure execution boundary. MCP is the
primary agent experience, the CLI remains complete for direct human, CI,
recovery, and break-glass operation, and both adapt the same application use
cases. The rationale and division of labour are defined in
[Agent-native product posture](19-agent-native-posture.md).

- **AINFRA-PROD-001:** ainfra MUST provide one understandable command surface
  for acquiring a template, planning infrastructure, applying a reviewed plan,
  configuring hosts, inspecting status, and destroying owned infrastructure.
- **AINFRA-PROD-002:** ainfra MUST keep OpenTofu configuration and variables
  native and directly usable with the `tofu` CLI.
- **AINFRA-PROD-003:** ainfra MUST keep Ansible playbooks, roles, collections,
  and variables native and directly usable with Ansible tools.
- **AINFRA-PROD-004:** a template author MUST be able to add a conforming
  provider or hosting-service template without changing ainfra source code.
- **AINFRA-PROD-005:** every mutation MUST be attributable to an explicit
  canonical deployment root and identity, template version, operation, and
  reviewed input.
- **AINFRA-PROD-006:** direct engine commands MUST remain a documented recovery
  and diagnosis path.
- **AINFRA-PROD-007:** Ansible owns host inspection, facts, fact caching, task
  semantics, check mode, and configuration changes. ainfra MUST NOT maintain a
  parallel model of desired or current host state, interpret fact-cache
  contents, or independently determine host convergence. It MAY retain
  sanitized Ansible Runner events and summarize only outcomes reported by
  Ansible as ainfra execution evidence.
- **AINFRA-PROD-008:** OpenTofu is the mandatory provisioning and reviewed
  plan/apply/destroy engine in v1; Ansible is conditional. An Ansible-only mode
  is outside v1 because it would require a new trusted-inventory input,
  different command applicability, and provisioning/teardown semantics not
  present in the current lifecycle contract. It MUST NOT be simulated with an
  empty OpenTofu root module.
- **AINFRA-PROD-009:** ainfra MUST expose agent-readable discovery, planning,
  execution, status, evidence, and recovery through bounded typed operations
  without adding an agent-specific infrastructure language.
- **AINFRA-PROD-010:** MCP, CLI, and future adapters MUST call one interface-
  neutral application core and MUST NOT implement divergent lifecycle rules.

## Target users

Primary users are AI agents operating for infrastructure-capable developers,
consultants, and platform teams, plus those humans operating ainfra directly.
They want reviewed, repeatable deployments without adopting a mandatory remote
infrastructure control plane. CI systems and portals are additional callers of
the same typed lifecycle.

Users are expected to understand the cost and ownership implications of the
chosen template. ainfra improves repeatability and safety; it does not replace
provider knowledge.

## Core use cases

1. Apply a locally developed template to a deployment.
2. Apply a template pinned to a Git repository commit and subdirectory.
3. Update a deployment to a newly reviewed template revision.
4. Plan and apply infrastructure through OpenTofu.
5. Generate inventory from declared non-secret OpenTofu outputs.
6. When the template declares configurable hosts, configure and verify them
   through Ansible Runner.
7. Destroy exactly the infrastructure covered by a reviewed destroy plan.
8. Author and validate a new template for a provider or hosting service.
9. Reproduce operations inside a user-built container when desired.
10. Let an authorized agent inspect, plan, explain, execute, monitor, and
    recover the same bounded lifecycle without shell wrappers or prose parsing.

## Non-goals

- **AINFRA-NONGOAL-001:** ainfra is not a Terraform/OpenTofu language compiler.
- **AINFRA-NONGOAL-002:** ainfra does not merge, transform, or generate operator
  tfvars from another configuration language.
- **AINFRA-NONGOAL-003:** ainfra does not replace Ansible with a Go automation
  engine.
- **AINFRA-NONGOAL-004:** ainfra does not provide a hosted control plane, job
  queue, web UI, RBAC service, or multi-tenant secrets store.
- **AINFRA-NONGOAL-005:** ainfra does not install workloads merely because a
  template creates Kubernetes-ready hosts. Workload installation belongs to a
  separately identified template or downstream system.
- **AINFRA-NONGOAL-006:** ainfra does not publish or operate a container image.
- **AINFRA-NONGOAL-007:** Windows is not a supported execution or release
  target for v1.
- **AINFRA-NONGOAL-008:** compatibility with Rust-era project files, run
  records, CLI flags, or implementation structure is not required.
- **AINFRA-NONGOAL-009:** ainfra does not embed OpenTofu or Ansible as
  libraries.

## Ownership boundary

| Owner | Responsibilities |
|---|---|
| ainfra application core | Source acquisition, locking, validation, lifecycle sequencing, safe subprocess execution, plan binding, inventory generation, run evidence, diagnostics, redaction, teardown guards |
| MCP and CLI adapters | Typed input conversion, capability presentation, result rendering, transport behavior, and delegation to identical application use cases |
| Template | Provider resources, OpenTofu variables and outputs, Ansible content, supported topology, provider-specific security, template documentation |
| Deployment | Template selection, pointers to native engine input files, credential references, approvals |
| OpenTofu | Infrastructure dependency graph, provider execution, plan, apply, state, locking |
| Ansible Runner/Core | Host inspection and configuration, task semantics, facts and fact caching, check mode, collection behavior, and engine-reported outcomes |
| Operator | Credentials, plan review, cost approval, SSH trust, destructive approval, independent provider verification |

## Supported delivery

ainfra v1 publishes statically linked binaries where practical for Linux
`amd64`/`arm64` and macOS `amd64`/`arm64`. The project additionally maintains a
Dockerfile that users MAY build for convenience. The Dockerfile is not the
normative runtime and no official prebuilt image is promised.
