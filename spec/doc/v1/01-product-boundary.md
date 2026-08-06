## Goals

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
  deployment, template version, environment, operation, and reviewed input.
- **AINFRA-PROD-006:** direct engine commands MUST remain a documented recovery
  and diagnosis path.
- **AINFRA-PROD-007:** Ansible owns host inspection, facts, fact caching, task
  semantics, check mode, and configuration changes. ainfra MUST NOT maintain a
  parallel model of desired or current host state, interpret fact-cache
  contents, or independently determine host convergence. It MAY retain
  sanitized Ansible Runner events and summarize only outcomes reported by
  Ansible as ainfra execution evidence.

## Target users

Primary users are infrastructure-capable developers, consultants, and small
platform teams who want reviewed, repeatable deployments without adopting a
remote infrastructure control plane. Secondary users are AI coding agents
working under human review.

Users are expected to understand the cost and ownership implications of the
chosen template. ainfra improves repeatability and safety; it does not replace
provider knowledge.

## Core use cases

1. Apply a locally developed template to a deployment.
2. Apply a template pinned to a Git repository commit and subdirectory.
3. Update a deployment to a newly reviewed template revision.
4. Plan and apply infrastructure through OpenTofu.
5. Generate inventory from declared non-secret OpenTofu outputs.
6. Configure and verify hosts through Ansible Runner.
7. Destroy exactly the infrastructure covered by a reviewed destroy plan.
8. Author and validate a new template for a provider or hosting service.
9. Reproduce operations inside a user-built container when desired.

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
| ainfra CLI | Source acquisition, locking, validation, lifecycle sequencing, safe subprocess execution, plan binding, inventory generation, run evidence, diagnostics, redaction, teardown guards |
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
