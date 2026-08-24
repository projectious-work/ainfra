## Market conclusion

ainfra does not create a new orchestration category. Open-source tools already
cover OpenTofu execution, reusable stacks, workflow orchestration, pull-request
automation, service catalogs, and topology-driven deployment. Its opportunity
is a deliberately smaller combination:

> A local-first runner for reviewed, versioned infrastructure templates that
> keeps OpenTofu and Ansible native while standardizing their secure handoff,
> execution evidence, diagnosis, and teardown.

The subprocess sequencing is not a defensible product by itself. The product
value is the combination of a small contract, conformance tooling, trustworthy
templates, and safe local execution. Without a useful template catalog and
strong conformance checks, ainfra would largely be custom glue.

## Competitive landscape

| Alternative | Overlap | Difference from ainfra |
|---|---|---|
| [Terramate](https://terramate.io/docs/concepts/orchestration) | Native OpenTofu execution, workflows, stacks, change detection | Broader graph and multi-stack orchestration; no prescribed OpenTofu-output-to-Ansible template handoff. |
| [Terragrunt](https://terragrunt.com/) | Remote sources, reusable infrastructure, backend configuration, stack execution | Adds Terragrunt configuration and primarily addresses OpenTofu/Terraform composition. |
| [Atmos](https://atmos.tools/intro/) | Components, environments, stacks, and workflows | Introduces YAML stack configuration above Terraform inputs, which ainfra deliberately avoids. |
| [Atlantis](https://www.runatlantis.io/) | Reviewed plans, exact-plan application, locking, and audit history | A self-hosted pull-request automation service rather than a local cross-engine template runner. |
| [Kubitect](https://kubitect.io/latest/getting-started/introduction/) | Go CLI coordinating Terraform and Kubernetes configuration | A vertical Kubernetes/libvirt product with a higher-level configuration model, not a provider-extensible native-file contract. |
| [Backstage](https://backstage.io/docs/features/software-templates/) | Curated templates and repeatable developer self-service | A developer portal, catalog, scaffolder, and organizational integration platform; it can present or invoke ainfra rather than replace its local execution contract. |
| [TOSCA](https://docs.oasis-open.org/tosca/TOSCA/v2.0/TOSCA-v2.0.pdf) and [xOpera](https://xlab-si.github.io/xopera-docs/07-examples.html) | Packaged service templates, lifecycle orchestration, and Ansible artifacts | TOSCA defines a topology and lifecycle metamodel interpreted by an orchestrator; ainfra explicitly refuses that additional infrastructure language. |

Terramate is the closest general architectural alternative. A team could
reproduce much of ainfra with Terramate scripts and repository conventions,
but would still need to design and enforce the cross-engine contract,
certification rules, deterministic inventory, and deployment-focused safety
model.

TOSCA is the clearest contrasting design. Its typed nodes, relationships,
requirements, capabilities, interfaces, policies, and workflows target semantic
portability and full lifecycle orchestration. ainfra instead accepts less
semantic portability in exchange for native OpenTofu and Ansible files, direct
diagnosis, and a much smaller orchestrator.

## Initial market wedge

ainfra is well positioned for development, test, sandbox, demonstration, and
small-team infrastructure environments when it offers reviewed golden paths
without requiring a platform control plane or a new configuration language.
The target operator should be able to select a template, fill native inputs,
review an OpenTofu plan, and receive a configured environment without manually
wiring OpenTofu output into Ansible inventory.

This promise is not “no infrastructure knowledge” and not “no dependencies.”
Operators still need the provider concepts, credentials, OpenTofu, Ansible,
Git, and SSH applicable to their template. The simplification is fewer custom
scripts, fewer organization-specific conventions, one documented lifecycle,
and a direct escape path through standard tools.

The initial golden-path wedge remains valid, but the leading product posture is
now:

> The agent-native infrastructure execution boundary for reviewed,
> reproducible OpenTofu and Ansible deployments.

The supporting promise is:

> Agents create and adapt standard templates. ainfra validates, locks, plans,
> authorizes, executes, and records them through MCP or CLI.

This does not claim deterministic cloud outcomes or infrastructure without
expert knowledge. It distinguishes ainfra from an agent directly invoking a
shell: the differentiated product is reusable institutional knowledge plus the
governed boundary around authority, credentials, reviewed change, execution,
evidence, and recovery.

Production use can remain possible, but the initial promise should emphasize
fast, understandable, disposable or reconstructible environments. Certified
production templates require a materially higher security, recovery, upgrade,
and operational evidence bar.

## Relationship to Backstage

The comparison to Backstage is useful at the product-experience level, not the
architecture level. Backstage provides a centralized catalog and software
templates that collect parameters, execute scaffolder actions, publish source,
and register components. ainfra provides a local, infrastructure-specific
execution contract and retains no service catalog.

They are complementary:

- Backstage can expose an approved ainfra template as an organizational golden
  path and invoke a controlled ainfra execution action;
- ainfra can remain useful without Backstage for individuals and small teams;
- ainfra templates remain directly usable and diagnosable outside any portal;
  and
- a future portal integration MUST NOT become a required control plane or a
  second variable model.

## Kubernetes bootstrap boundary

Delivering a Kubernetes-ready environment, including an initial K3s or similar
installation, directly supports the development-environment promise. It turns
provisioned machines into a usable infrastructure environment. The boundary is
initial bootstrap, not ongoing cluster operations.

The lean v1 approach is to implement bootstrap in template-owned Ansible
content when Ansible can express it safely and idempotently. ainfra already
owns the OpenTofu-to-inventory-to-Ansible sequence, so this requires no new CLI
engine, manifest language, or lifecycle state model.

A dedicated external bootstrap tool MAY be added later when real templates
show that it provides important behavior Ansible should not duplicate. Such an
addition must use a narrow, named adapter with a versioned input/output and
execution-evidence contract. It must not introduce arbitrary manifest hooks or
a generic workflow language.

Whether implemented through Ansible or a later adapter, the boundary is:

- install an initial, pinned Kubernetes distribution on provisioned hosts;
- establish and verify the initial control-plane endpoint and node membership;
- handle join tokens and kubeconfig as sensitive engine-owned material;
- emit only the minimum sanitized result needed for the operator; and
- stop before workload deployment, continuous reconciliation, upgrades,
  backup operation, policy management, or general day-two cluster maintenance.

The [official K3s documentation](https://docs.k3s.io/quick-start) describes
installation as creating a persistent system service and kubeconfig, with
additional agents joined using a server URL and token. Those behaviors make
bootstrap security and teardown explicit template responsibilities rather than
generic ainfra state.

## Product test

Each proposed feature should answer yes to all of these questions:

1. Does it remove repeated integration work across multiple templates?
2. Can ainfra implement it without understanding provider or host topology?
3. Do OpenTofu, Ansible, or another named child tool remain authoritative?
4. Can the native equivalent command remain documented and usable?
5. Does it strengthen template conformance, safety, or evidence rather than add
   a competing infrastructure language?

If not, the feature belongs in a template, an external tool, or a portal such
as Backstage—not in the ainfra core.
