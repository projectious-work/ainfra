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
    tofu:
      variableFiles:
        - terraform.tfvars
      backendConfigFiles:
        - backend.hcl
    ansible:
      variableFiles:
        - ansible-vars.yaml
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

## Native engine inputs

The input entries in `ainfra.yaml` are ordered lists of pointers to native
engine files. `tofu.variableFiles` maps to repeated OpenTofu `-var-file`
arguments during planning, `tofu.backendConfigFiles` maps to repeated
`-backend-config` arguments during initialization, and
`ansible.variableFiles` maps to native Ansible extra-vars files. Missing lists
are equivalent to empty lists. List order is preserved because the owning
engine may use it for precedence.

The contents of these files are deliberately a loose, template-owned contract.
A template MAY introduce any variables or backend configuration needed for its
provider, topology, operating system, or application bootstrap. It MAY also
integrate configuration directly and require no corresponding deployment file.
The template documents which files and values are mandatory or optional.
ainfra standardizes pointer handling and execution, not the domain model.

The template defines its input API through native mechanisms: OpenTofu variable
declarations, types, defaults, validation blocks, and descriptions; and Ansible
playbooks, role defaults, role argument specifications where applicable, and
documentation. Variable additions and breaking changes follow the template's
version and migration policy.

- **AINFRA-CONTRACT-020:** ainfra MUST pass every declared OpenTofu variable
  file as `-var-file=<path>`, once and in declaration order, during planning.
- **AINFRA-CONTRACT-021:** ainfra MUST NOT parse tfvars to implement
  provider-specific validation. It MAY hash the bytes and run OpenTofu
  validation.
- **AINFRA-CONTRACT-022:** ainfra MUST pass every declared Ansible variable
  file as a native extra-vars file, once and in declaration order.
- **AINFRA-CONTRACT-023:** reserved common names MUST be prefixed `ainfra_` and
  documented. Templates MUST NOT require a generated common-variable object.
- **AINFRA-CONTRACT-024:** ainfra MUST treat native variable contents as opaque
  bytes except for hashing, safe file handling, redaction, and delegation to the
  owning engine; it MUST NOT implement template-specific semantic validation.
- **AINFRA-CONTRACT-025:** ainfra MUST NOT merge, rename, default, generate, or
  translate native variables between deployment, OpenTofu, and Ansible files.
- **AINFRA-CONTRACT-026:** every required template variable MUST be declared and
  documented through the owning engine's native mechanisms; the ainfra manifest
  MUST NOT duplicate a variable schema.
- **AINFRA-CONTRACT-027:** unknown, missing, mistyped, or invalid native values
  are reported by OpenTofu or Ansible and surfaced by ainfra without claiming
  independent semantic interpretation.
- **AINFRA-CONTRACT-028:** ainfra MUST pass every declared backend
  configuration file to `tofu init` as `-backend-config=<path>`, once and in
  declaration order. It MUST treat its contents as opaque and MUST NOT infer or
  enforce backend type, state location, locking, durability, or lifecycle.
- **AINFRA-CONTRACT-029:** native input lists MAY be empty or omitted. Whether a
  template requires a file or value is part of the template's documented
  contract and is ultimately validated by the owning engine.

This looseness ends at the cross-template boundary. ainfra-owned manifests,
locks, machine results, and standardized OpenTofu output remain strict,
versioned schemas so inventory generation and lifecycle safety are portable.

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

Additional source schemes are outside v1 and require a separate future product
decision.

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

The lockfile records the requested source, resolved immutable Git commit or
normalized local-source identity, selected subdirectory, template manifest
version, complete template digest, and resolution time. A local identity is not
itself immutable; its digest provides the content binding.

- **AINFRA-LOCK-001:** `ainfra template lock` and `template update` are the only
  commands that MAY rewrite the lockfile.
- **AINFRA-LOCK-002:** plan MUST fail when source content differs from the
  locked digest.
- **AINFRA-LOCK-003:** apply and destroy MUST use the materialized template and
  digest bound into the reviewed plan record, not reacquire a mutable ref.
- **AINFRA-LOCK-004:** lockfile comparison MUST use canonical serialization.
- **AINFRA-LOCK-005:** a template tree digest MUST be SHA-256 over the
  following byte stream: the ASCII bytes `ainfra-template-tree-v1` followed by
  one NUL byte, then every included regular file sorted by its normalized
  UTF-8 relative-path bytes. Each file contributes an unsigned 64-bit
  big-endian path length, the path bytes using `/` separators, one byte that is
  `1` when any executable bit is set and `0` otherwise, an unsigned 64-bit
  big-endian content length, and the unmodified content bytes.
- **AINFRA-LOCK-006:** digest paths MUST be valid UTF-8 in Unicode NFC form,
  relative, traversal-free, and contain no empty segment. `.git/`, `.ainfra/`,
  and engine runtime artifacts are not template content: `.git/` is excluded,
  while the others are rejected. Symlinks and special files are rejected
  rather than hashed.
