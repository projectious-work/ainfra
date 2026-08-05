## `ainfra.yaml`

`ainfra.yaml` contains orchestration metadata only. It MUST NOT duplicate
OpenTofu or Ansible variables.

```yaml
apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: git::https://github.com/company/templates.git//hetzner/kubernetes
    ref: v2.3.1
  inputs:
    tofu: terraform.tfvars
    ansible: ansible-vars.yaml
  state:
    backendConfig: backend.hcl
```

- **AINFRA-CONTRACT-001:** unknown fields in ainfra-owned documents MUST fail
  validation.
- **AINFRA-CONTRACT-002:** secrets MUST be external references or engine-native
  environment variables, never fields containing secret values.
- **AINFRA-CONTRACT-003:** the schema at
  [`../../schemas/v1/ainfra.schema.json`](../../schemas/v1/ainfra.schema.json)
  defines the structural contract.

## Template manifest

The manifest declares identity, compatibility, engine directories, required
standard outputs, and optional prerequisites. It is not an executable DSL and
MUST NOT contain arbitrary commands or hooks.

```yaml
apiVersion: ainfra.projectious.work/v1
kind: Template
metadata:
  name: hetzner-kubernetes
  version: 1.0.0
spec:
  engines:
    tofu:
      directory: tofu
      version: ">=1.10.0 <2.0.0"
    ansible:
      directory: ansible
      playbook: site.yml
      version: ">=2.18.0 <3.0.0"
  outputs:
    inventory: ainfra_inventory
```

- **AINFRA-CONTRACT-010:** the manifest MUST stay provider-neutral at the CLI
  layer; provider-specific behavior belongs to source and documentation.
- **AINFRA-CONTRACT-011:** engine ranges MUST be checked before mutation.
- **AINFRA-CONTRACT-012:** declared paths MUST be regular contained paths.
- **AINFRA-CONTRACT-013:** required Ansible collections MUST be declared in
  the template's native `requirements.yml`, not duplicated in the manifest.

## Native variables

- **AINFRA-CONTRACT-020:** ainfra MUST call OpenTofu with
  `-var-file=<deployment tfvars>`.
- **AINFRA-CONTRACT-021:** ainfra MUST NOT parse tfvars to implement
  provider-specific validation. It MAY hash the bytes and run OpenTofu
  validation.
- **AINFRA-CONTRACT-022:** ainfra MUST call Ansible Runner with the deployment
  `ansible-vars.yaml` as a native extra-vars file.
- **AINFRA-CONTRACT-023:** reserved common names MUST be prefixed `ainfra_` and
  documented. Templates MUST NOT require a generated common-variable object.

## Standard OpenTofu output

Every configurable-host template MUST expose a non-sensitive output named by
the manifest (default `ainfra_inventory`) with this logical shape:

```json
{
  "schema_version": "1",
  "groups": {
    "control_plane": {
      "hosts": {
        "cp-1": {
          "ansible_host": "10.42.0.10",
          "ansible_user": "ainfra",
          "vars": {"ainfra_role": "control-plane"}
        }
      }
    }
  }
}
```

- **AINFRA-CONTRACT-030:** the output MUST be marked non-sensitive by
  OpenTofu and MUST contain no credentials, private keys, tokens, passwords, or
  secret references that resolve to values.
- **AINFRA-CONTRACT-031:** host and group names MUST satisfy Ansible inventory
  naming rules enforced by ainfra.
- **AINFRA-CONTRACT-032:** host variables MUST be JSON scalar, array, or object
  values safe to serialize to YAML.
- **AINFRA-CONTRACT-033:** ainfra MUST reject unknown output schema versions.
- **AINFRA-CONTRACT-034:** generated inventory MUST be deterministic: sorted
  groups, hosts, and map keys with stable YAML serialization.
- **AINFRA-CONTRACT-035:** templates without configurable hosts MAY declare
  inventory mode `none`; configure and verify commands then MUST refuse with an
  explanatory result rather than succeed silently.

## Source schemes

v1 MUST support:

- `local:<relative-path>` for development;
- `git::<https-or-ssh-url>//<subdirectory>` with an explicit `ref`.

OCI template artifacts MAY be added later but are not required for v1.

- **AINFRA-SOURCE-001:** source strings MUST be parsed structurally, never by
  shell evaluation.
- **AINFRA-SOURCE-002:** local paths MUST remain within policy-approved roots;
  the default is the deployment root and its parent repository.
- **AINFRA-SOURCE-003:** a Git ref MAY be a tag, branch, or commit during
  `ainfra template update`, but normal plan/apply MUST use the immutable commit
  recorded in `ainfra.lock`.
- **AINFRA-SOURCE-004:** acquisition MUST use argument arrays and a private
  cache directory. It MUST NOT interpolate URLs into shell commands.
- **AINFRA-SOURCE-005:** subdirectories MUST be cleaned and checked for
  traversal after checkout.
- **AINFRA-SOURCE-006:** redirects, SSH host trust, and Git credential handling
  MUST remain visible and governed by Git rather than copied into ainfra.yaml.

## Lockfile

The lockfile records the requested source, resolved commit, selected
subdirectory, template manifest version, complete template digest, and
resolution time.

- **AINFRA-LOCK-001:** `ainfra template lock` and `template update` are the only
  commands that MAY rewrite the lockfile.
- **AINFRA-LOCK-002:** plan MUST fail when source content differs from the
  locked digest.
- **AINFRA-LOCK-003:** apply and destroy MUST use the materialized template and
  digest bound into the reviewed plan record, not reacquire a mutable ref.
- **AINFRA-LOCK-004:** lockfile comparison MUST use canonical serialization.
