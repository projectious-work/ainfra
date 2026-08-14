---
title: Output, inventory, and Ansible
weight: 45
description: Validate OpenTofu output and configure hosts from a reviewed run.
---

{{< pj-callout type="warning" title="Development capability" >}}
This Phase 5 workflow is available on `v1.x-dev`. The latest published alpha,
`v1.0.0-alpha.4`, stops after reviewed OpenTofu apply.
{{< /pj-callout >}}

Phase 5 keeps the OpenTofu-to-Ansible handoff narrow. A template declares one
standardized, non-sensitive OpenTofu output. ainfra validates only that value,
converts it to deterministic inventory, and passes template-specific settings
through native Ansible variable files rather than an ainfra variable model.

## Run the stages separately

Start with a successfully applied reviewed run:

```sh
ainfra output ./deployments/example --run RUN_ID
ainfra inventory ./deployments/example --run RUN_ID
ainfra configure ./deployments/example --run RUN_ID
ainfra configure ./deployments/example --run RUN_ID --check
```

`output` writes private `output.json`; `inventory` derives private
`inventory.yaml` from that validated artifact. Normal `configure` performs the
declared playbook. The separate `--check` execution verifies that every
expected host was processed with zero changes, failures, and unreachable
hosts.

Use `--format json` for the versioned machine result. Results contain only
artifact digests, engine attribution, and evidence paths. Ansible Runner
artifacts remain marked sensitive.

## Run the composed pipeline

`deploy` applies the exact reviewed plan and then runs every applicable Phase
5 stage in order:

```sh
ainfra deploy ./deployments/example --plan RUN_ID
```

It never creates an implicit plan. Infrastructure-only templates return typed
`not-applicable` results for output and inventory, skip Ansible, and do not
require `ansible-runner` to be installed.

## SSH trust

SSH inventory requires `spec.ssh.knownHosts` to name a populated native file
that was bound into the reviewed run. ainfra forces host-key checking and
passes the bound snapshot to Ansible. It does not discover, accept, or repair
host keys automatically.

## Recovery rules

Output validation and inventory conversion are non-mutating and may be rerun.
A normal or check-mode configuration stage starts at most once for a retained
run. If it fails or is interrupted, inspect its Runner artifacts and run
evidence before creating a new reviewed plan; ainfra does not automatically
repeat a potentially mutating playbook.
