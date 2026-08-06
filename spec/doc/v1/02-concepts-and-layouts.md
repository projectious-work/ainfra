## Concepts

### Template

A template is an immutable-at-execution directory containing a small ainfra
manifest, native OpenTofu root module, native Ansible project, examples, and
operator documentation.

### Deployment

A deployment is a user-owned directory that selects one template and contains
one concrete set of native inputs. A deployment is the unit of planning,
application, configuration, status, destruction, and local execution evidence.
v1 has no separate environment entity or named input-set selector: one
deployment directory represents one independently operated deployment.

### Run

A run is one locally recorded lifecycle attempt bound to a deployment,
template digest, native input bytes, engine versions, operation, and—where
applicable—a saved OpenTofu plan. This is ainfra execution state, not
infrastructure state.

Names such as `development`, `staging`, and `production` MAY be deployment
directory names or native template variable values. ainfra treats those names
as paths or opaque engine input, never as a second execution identity. Teams
that need several instances create several deployment directories:

```text
deployments/
├── development/
│   ├── ainfra.yaml
│   └── .ainfra/
├── staging/
│   ├── ainfra.yaml
│   └── .ainfra/
└── product-a/
    └── development/
        ├── ainfra.yaml
        └── .ainfra/
```

## Deployment layout

```text
my-deployment/
├── ainfra.yaml
├── ainfra.lock
├── terraform.tfvars
├── ansible-vars.yaml
├── backend.hcl                 # optional native OpenTofu input
├── known_hosts                 # optional; normally ignored
└── .ainfra/                    # ignored local operational state
    ├── cache/
    └── runs/
```

- **AINFRA-LAYOUT-001:** `ainfra.yaml` MUST be a regular, non-symlink file in
  the deployment root.
- **AINFRA-LAYOUT-002:** files listed under `inputs.tofu.variableFiles` and
  `inputs.tofu.backendConfigFiles` MUST be passed to OpenTofu without semantic
  transformation.
- **AINFRA-LAYOUT-003:** files listed under `inputs.ansible.variableFiles` MUST
  be passed to Ansible as native extra-vars files without semantic
  transformation.
- **AINFRA-LAYOUT-004:** `ainfra.lock` MUST contain immutable resolved template
  identity and MUST be committed for a committed deployment.
- **AINFRA-LAYOUT-005:** `.ainfra/`, plans, state, generated inventory, engine
  caches, and run logs MUST NOT be committed by default.
- **AINFRA-LAYOUT-006:** every native input pointer MUST be a relative path
  contained by the deployment root. The filenames are otherwise unrestricted.
- **AINFRA-LAYOUT-007:** one deployment directory MUST contain exactly one
  `ainfra.yaml` and one deployment-local `.ainfra/` execution-evidence tree.
  Coordinating several deployment directories is outside v1.

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
- **AINFRA-LAYOUT-014:** infrastructure state and ainfra execution state MUST
  NOT be part of template source.

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

The minimal conforming, provider-free reference template is
[`../../examples/v1/template-example/`](../../examples/v1/template-example/).
It exercises the complete localhost lifecycle without claiming certified cloud
provider support. The adjacent deployment is a second invocation example of
the same template contract.
