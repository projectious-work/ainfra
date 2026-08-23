## Security posture

ainfra executes third-party code with infrastructure credentials. A template,
its OpenTofu modules/providers, and its Ansible content are executable supply
chain inputs. Schema validation does not make them trusted.

The default posture is explicit trust, least exposure, immutable review
bindings, and refusal on ambiguity.

## Threats

The v1 threat model includes:

- malicious or replaced Git template content;
- mutable tags and branches;
- archive or path traversal;
- symlink escape and special files;
- shell/argument injection;
- malicious OpenTofu modules and providers;
- malicious Ansible roles and collections;
- secrets in tfvars, backend files, environment, state, plans, outputs, logs,
  errors, inventories, and subprocess arguments;
- confused-deputy destruction against another deployment or backend;
- SSH interception;
- interrupted mutations and false success claims;
- poisoned local caches;
- unsafe optional Dockerfile dependencies.

## Trust and source controls

- **AINFRA-SEC-001:** normal lifecycle execution MUST use locked immutable
  source identity and verified content digest.
- **AINFRA-SEC-002:** changing a requested source/ref MUST require an explicit
  lock/update command and a visible diff.
- **AINFRA-SEC-003:** cache hits MUST be digest-verified before use.
- **AINFRA-SEC-004:** template materialization MUST reject absolute paths,
  traversal, escaping symlinks, devices, FIFOs, sockets, and unsafe permission
  transitions.
- **AINFRA-SEC-005:** source credentials MUST be handled by Git/SSH credential
  mechanisms and MUST NOT be persisted in ainfra-owned documents.
- **AINFRA-SEC-006:** source URLs in human or machine output MUST redact
  userinfo and sensitive query parameters.

## Subprocess controls

- **AINFRA-SEC-010:** subprocesses MUST be invoked with executable plus argument
  array, never a shell command string.
- **AINFRA-SEC-011:** executable discovery MUST resolve a regular file and
  record its version; mutation MUST refuse a changed executable binding after
  plan unless explicitly allowed by a new plan.
- **AINFRA-SEC-012:** child environment MUST be built from an allowlist plus
  documented engine credential namespaces; unrelated parent variables SHOULD
  be excluded.
- **AINFRA-SEC-013:** working directories MUST be canonical, contained, and
  recorded before execution.
- **AINFRA-SEC-014:** stdout and stderr MUST be streamed without placing secret
  values into command diagnostics. Raw retained logs are sensitive.
- **AINFRA-SEC-015:** no secret value may appear in a process argument when a
  supported file, stdin, agent, or environment mechanism exists.

## Secret handling

The complete deployment-credential acquisition, delivery, lifecycle, and
scenario contract is defined in
[Deployment credential acquisition and delivery](18-deployment-credentials.md).
That contract distinguishes credentials authorizing ainfra's engines from
bootstrap secrets placed by templates and workload secrets consumed after
handover.

- **AINFRA-SEC-020:** ainfra MUST NOT accept inline secret fields in
  `ainfra.yaml` or the template manifest.
- **AINFRA-SEC-021:** committed examples MUST use placeholders and `.invalid`
  domains, never syntactically plausible live credentials.
- **AINFRA-SEC-022:** ainfra MUST classify plans, state, backend files, tfvars,
  engine logs, Ansible artifacts, and crash diagnostics as potentially
  sensitive regardless of user declarations.
- **AINFRA-SEC-023:** standardized output and inventory MUST undergo both
  structural allowlisting and secret-pattern scanning.
- **AINFRA-SEC-024:** redaction MUST cover exact known secret values and common
  credential shapes. Redaction MUST be tested across chunk boundaries.
- **AINFRA-SEC-025:** permissions for local operational directories and files
  MUST default to owner-only where supported.
- **AINFRA-SEC-026:** structured engine events are potentially sensitive and
  MUST undergo the same containment, permission, retention, and explicit raw-
  access controls as stdout, stderr, plans, and Runner artifacts.
- **AINFRA-SEC-027:** engine evidence profiles MUST be embedded, reviewed
  ainfra assets in v1; templates and deployments MUST NOT supply or override
  parsing, classification, or redaction rules.
- **AINFRA-SEC-028:** a template or deployment MUST NOT select a credential
  provider or contain provider authentication material; it may declare or bind
  only the symbolic contract defined by the credential specification.
- **AINFRA-SEC-029:** ainfra MUST acquire deployment credentials only through
  supported bounded adapters or documented user-managed native mechanisms.

## Engine state and destructive safety

- **AINFRA-SEC-030:** infrastructure state lifecycle, storage, locking,
  migration, and recovery belong entirely to OpenTofu and the selected
  backend. ainfra MUST NOT impose a local/remote or disposable/durable policy.
- **AINFRA-SEC-031:** ainfra MUST NOT parse, modify, migrate, copy, or directly
  inspect OpenTofu state. Security and recovery recommendations belong in
  template and operator documentation and remain engine-native policy.
- **AINFRA-SEC-032:** the ordered native input pointers and file bytes,
  including backend configuration files, MUST be bound to the reviewed plan as
  opaque execution inputs.
- **AINFRA-SEC-033:** ainfra MUST never implement unreviewed `tofu destroy` as a
  shortcut.
- **AINFRA-SEC-034:** ownership tags or labels MUST be defined by templates and
  tested where providers support them.

## SSH safety

- **AINFRA-SEC-040:** host key checking MUST not be disabled by ainfra or a
  certified template.
- **AINFRA-SEC-041:** ainfra MUST not generate long-lived private keys.
- **AINFRA-SEC-042:** SSH agent forwarding and socket mounts MUST be explicit
  user choices and documented as sensitive delegation.
- **AINFRA-SEC-043:** templates implementing temporary bastions MUST document
  creation, restricted ingress, use, closure, and teardown verification.

## Dependency and artifact security

Go dependencies, OpenTofu providers/modules, Ansible collections/roles, and
release tools MUST be pinned through their native lock or checksum mechanisms.
Release checks include reachable-vulnerability scanning, repository secret
scanning, SBOM generation, artifact scanning, and signed checksums or
attestations where the release process supports them.

## Optional Dockerfile

The Dockerfile MUST use pinned base-image identity for releases, avoid embedded
credentials, create a non-root user, verify downloaded tools, and remain a
convenience. ainfra neither publishes the resulting image nor defines how the
user starts, mounts, or enters it.
