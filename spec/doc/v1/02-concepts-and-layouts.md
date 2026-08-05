## Concepts

### Template

A template is an immutable-at-execution directory containing a small ainfra
manifest, native OpenTofu root module, native Ansible project, examples, and
operator documentation.

### Deployment

A deployment is a user-owned directory that selects one template and contains
environment-specific native inputs. A deployment is the unit of planning,
application, configuration, status, and destruction.

### Run

A run is one locally recorded lifecycle attempt bound to a deployment,
template digest, inputs, engine versions, operation, and—where applicable—a
saved OpenTofu plan.

### Environment

An environment is a named input set within a deployment. v1 SHOULD prefer one
deployment directory per independently owned state boundary. Multiple named
environments MAY share a deployment only when their state, credentials, and
teardown ownership remain unambiguous.

## Deployment layout

```text
my-deployment/
├── ainfra.yaml
├── ainfra.lock
├── terraform.tfvars
├── ansible-vars.yaml
├── backend.hcl                 # optional; normally ignored when secret
├── known_hosts                 # optional; normally ignored
└── .ainfra/                    # ignored local operational state
    ├── cache/
    └── runs/
```

- **AINFRA-LAYOUT-001:** `ainfra.yaml` MUST be a regular, non-symlink file in
  the deployment root.
- **AINFRA-LAYOUT-002:** `terraform.tfvars` MUST be passed to OpenTofu without
  semantic transformation.
- **AINFRA-LAYOUT-003:** `ansible-vars.yaml` MUST be passed to Ansible as a
  native extra-vars file without semantic transformation.
- **AINFRA-LAYOUT-004:** `ainfra.lock` MUST contain immutable resolved template
  identity and MUST be committed for a committed deployment.
- **AINFRA-LAYOUT-005:** `.ainfra/`, plans, state, generated inventory, engine
  caches, and run logs MUST NOT be committed by default.
- **AINFRA-LAYOUT-006:** custom filenames MAY be declared in `ainfra.yaml`, but
  all paths MUST remain relative to and contained by the deployment root.

## Template layout

```text
template-name/
├── ainfra-template.yaml
├── README.md
├── tofu/
│   ├── versions.tf
│   ├── providers.tf
│   ├── variables.tf
│   ├── outputs.tf
│   └── ...
├── ansible/
│   ├── ansible.cfg
│   ├── requirements.yml
│   ├── site.yml
│   └── roles/
├── examples/
│   └── minimal/
│       ├── terraform.tfvars
│       └── ansible-vars.yaml
└── tests/
```

- **AINFRA-LAYOUT-010:** the directory basename and manifest template name MUST
  match after canonical normalization.
- **AINFRA-LAYOUT-011:** engine working directories MUST be declared and
  contained by the template root.
- **AINFRA-LAYOUT-012:** a template MUST contain at least one complete,
  secret-free example.
- **AINFRA-LAYOUT-013:** a template MUST document direct OpenTofu and Ansible
  commands equivalent to ainfra execution.
- **AINFRA-LAYOUT-014:** runtime state MUST NOT be part of template source.

## Run layout

```text
.ainfra/runs/<run-id>/
├── run.json
├── events.jsonl
├── plan.tfplan                # plan/apply/destroy workflows
├── plan.json                  # sanitized structural summary
├── output.json                # standardized non-secret output
├── inventory.yaml             # generated
├── engine/
│   ├── tofu.stdout.log
│   ├── tofu.stderr.log
│   └── ansible-artifacts/
└── workspace/                 # immutable materialized template
```

- **AINFRA-LAYOUT-020:** run directories MUST be created with restrictive
  permissions before invoking engines.
- **AINFRA-LAYOUT-021:** `run.json` MUST be immutable after the first mutating
  stage starts, except for explicitly append-only outcome references.
- **AINFRA-LAYOUT-022:** events MUST be append-only and monotonically ordered.
- **AINFRA-LAYOUT-023:** raw engine logs MUST be treated as sensitive even when
  ainfra applies redaction.
- **AINFRA-LAYOUT-024:** an interrupted or ambiguous mutation MUST require
  inspection before automatic continuation.

## Reference example

The normative structural example is under
[`../../examples/v1/`](../../examples/v1/). Example provider values are
illustrative; the directory and handoff contracts are normative.
