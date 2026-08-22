# Templates


Templates are self-contained, reviewable infrastructure environments for
specific AI-agent workloads. Phase 8 defines the released authoring contract;
Phase 9 ships the first provider-backed production candidate after offline
conformance and a cost-approved disposable live certification.

## Hetzner private K3s production candidate

`templates/hetzner-kubernetes-baseline` provisions private-management Hetzner
nodes, bootstraps pinned K3s, and runs a connector for an externally managed
Cloudflare Tunnel. Public SSH exists only through an opt-in temporary bastion
with a narrow source allowlist.

The candidate has complete native-variable documentation, offline conformance
coverage, and Phase 9 certification evidence. The certified lifecycle created
three private K3s control-plane nodes, verified convergence and tunnel-only
administration, removed the temporary bastion, destroyed the exact reviewed
resources, and independently confirmed that no owned Hetzner resources
remained. Every new deployment still requires its own authorization, trust
bootstrap, validation, and teardown evidence.

Start by copying the
[reference template](https://github.com/projectious-work/ainfra/tree/v1.x-dev/spec/examples/v1/template-example).
Keep the applicable authoring layout intact:

```text
README.md
docs/variables.md
tofu/{versions,variables}.tf
tofu/outputs.tf                           # when the template emits outputs
ansible/                                  # when hosts are configurable
tests/{README.md,validate.sh,fixtures/}   # output fixture when inventory exists
```

The template manifest declares identity, engine requirements, inventory and
input/output contracts. OpenTofu owns infrastructure resources; Ansible owns
host configuration. Keep native tool configuration inside those engines rather
than duplicating it in ainfra.

## Authoring workflow

1. Copy the reference template and give the manifest a unique name and version.
2. Declare OpenTofu variables, outputs and required provider versions.
3. Add Ansible inventory/playbooks and map host facts deliberately.
4. Document every supported input in `docs/variables.md`.
5. Write the lifecycle guide in `README.md`, including prerequisites,
   architecture, network exposure, costs, failure handling and teardown.
6. Create a clean-room fixture and make `tests/validate.sh` exercise it.
7. Run the local conformance gate and native validation:

   ```sh
   ainfra doctor template . --format json
   ./tests/validate.sh
   ```

`ainfra doctor template` validates the portable authoring contract. It reports
native OpenTofu and Ansible validation as delegated, so the clean-room script
must run compatible tools itself. A passing doctor report is not authorization
to create billable resources; obtain explicit approval immediately before any
live provider lifecycle.

## Documentation contract

`docs/variables.md` contains these H2 sections, in order: **OpenTofu
variables**, **Ansible variables**, **Cross-variable rules**, and **Examples**.
The variable tables identify name, type or shape, required/default values,
constraints, sensitivity and description.

The template `README.md` covers prerequisites, architecture, variables,
network, cost, failure, teardown, compatibility and validation. Never put a
credential in the manifest, fixture, README or generated evidence.

For a complete machine-oriented checklist, see the
[AI template-authoring guide](/docs/template-authoring-ai/). The normative
schema and lifecycle details remain in the
[template contract](https://github.com/projectious-work/ainfra/blob/v1.x-dev/spec/doc/v1/08-template-authoring.md).


---
Source: https://projectious-work.github.io/ainfra/docs/templates/index.md
