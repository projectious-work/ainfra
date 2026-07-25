# Architecture

## Boundaries

OpenTofu owns infrastructure desired state. Ansible owns host configuration.
The `ainfra` wrapper validates contracts and visibly orchestrates those tools.
Every wrapper operation has a documented direct-tool equivalent.

The first reference implementation is Hetzner-specific and produces
Kubernetes-ready hosts. It does not install or claim to operate Kubernetes.
Provider abstraction and workload deployment are outside the initial scope.

## Public contracts

Three strict JSON Schema draft 2020-12 contracts form the hand-off boundary:

1. `InfrastructureTemplate` declares engine locations, contract files,
   capabilities, and mandatory security invariants.
2. `TemplateInput` carries non-secret configuration and references to
   credentials or ignored local configuration.
3. `InfrastructureOutput` carries stable target identity, non-secret network
   facts, inventory metadata, and references to credentials.

Unsupported versions, kinds, capabilities, and unknown fields fail before
mutation. Template metadata is not an executable DSL and cannot declare
arbitrary shell hooks.

## Wrapper behavior

The intended command surface is:

```text
ainfra validate <path>
ainfra doctor
ainfra plan <template> [--input <path>]
ainfra apply <template> [--input <path>] [--approve]
ainfra destroy <template> [--input <path>] [--approve]
ainfra outputs <template> [--input <path>] [--format json|yaml]
```

Milestone 0 implements contract validation and the CLI boundary. Lifecycle
commands remain guarded until execution, approval, redaction, and
ownership-scoping tests exist.
