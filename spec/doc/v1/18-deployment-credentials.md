# Deployment credential acquisition and delivery

## Scope and terminology

Infrastructure engines need authority credentials before they can create,
inspect, change, or destroy infrastructure. Examples include `HCLOUD_TOKEN`,
cloud workload identity, backend credentials, and an SSH agent delegated to
Ansible. This specification calls these **deployment credentials**.

Deployment credentials are distinct from:

| Class | Purpose | ainfra responsibility |
|---|---|---|
| Deployment authority credential | Authorize an engine against a provider, backend, or managed host. | Acquire or accept, deliver narrowly, redact, renew when supported, and clean up. |
| Infrastructure bootstrap secret | Configure a service that the template creates. | Treat as a sensitive native engine input or external reference; never invent its domain model. |
| Workload secret | Used by a workload deployed onto resulting infrastructure. | Exclude values from ainfra results; expose only approved symbolic references. |

The secret authority is always operationally independent of the deployment
subject. If an ainfra template deploys OpenBao, an OpenBao instance used to
supply that deployment's provider credentials is a separate installation
operated by a platform team or equivalent authority. This is an authority
boundary, not a bootstrap exception.

## Design boundary

ainfra orchestrates credential acquisition and delivery but is not a vault,
key-management service, or general secret broker. It implements a small,
versioned provider capability contract. Templates declare what an engine
needs and the engine-native delivery destination; deployments bind those slots
to symbolic credential references. Provider configuration and authentication
remain outside the template and deployment repository.

The exact manifest and deployment schema additions are delivered with roadmap
Phase 10. Until that phase ships, documented ambient engine variables remain
the compatibility path and MUST continue to satisfy the existing subprocess,
redaction, and evidence rules.

An illustrative future contract is:

```yaml
# Template manifest: no provider choice and no secret value.
spec:
  credentials:
    - name: hcloud-api
      phases: [init, plan, apply, destroy]
      delivery:
        engine: tofu
        environment: HCLOUD_TOKEN

# Deployment: symbolic binding only.
spec:
  credentialBindings:
    hcloud-api: operator-hcloud
```

The binding `operator-hcloud` resolves through operator-controlled execution
configuration, not through template code. A template MUST NOT select SOPS,
OpenBao, a file path, a broker endpoint, or an arbitrary acquisition command.

- **AINFRA-CRED-001:** a credential slot MUST declare a stable symbolic name,
  required lifecycle phases, consuming engine, and supported native delivery
  form without containing or selecting a credential value or provider.
- **AINFRA-CRED-002:** a deployment credential binding MUST contain only a
  symbolic reference; provider configuration and authentication MUST remain
  outside committed deployment and template documents.
- **AINFRA-CRED-003:** credential adapters MUST implement a bounded ainfra-
  owned capability contract and MUST NOT expose arbitrary template hooks or
  shell execution.
- **AINFRA-CRED-004:** ainfra MUST NOT become a persistent secret store or
  offer secret creation, editing, rotation policy, or access-policy authoring.

## Initial providers

Phase 10 supports these sources:

| Provider | Intended use | Preconditions | Security posture |
|---|---|---|---|
| User-managed native source | Compatibility and operator-managed automation. | Required engine variables, files, agents, sockets, or workload identity already exist in the execution environment. | Security depends on the operator and execution platform; ainfra still narrows child exposure and redacts evidence. |
| SOPS | Interactive and controlled local execution, including an ainfra devcontainer whose host supplies the decrypted value. | SOPS is installed where ainfra executes; an authorized age, PGP, cloud-KMS, or OS-keychain-backed key is available. | Encrypted at rest; plaintext exists only just in time in process memory or a protected ephemeral delivery object. Not a good default for unattended renewal. |
| External OpenBao | Headless, remote, CI, Kubernetes, and stronger local use. | A separately operated OpenBao service, policy, authentication method, trusted TLS identity, and reachable endpoint. | Preferred initial managed provider: short-lived, scoped credentials and auditable access where the upstream system supports them. |

SOPS is a file-format and decryption provider in this architecture. The SOPS
binary runs where ainfra executes, not inside infrastructure being provisioned.
Its decryption key is supplied by the selected standard SOPS integration; it
is never stored in `ainfra.yaml`, a template, or a run record.

OpenBao integration uses its supported API or Agent interface. ainfra accepts
an already authenticated workload identity or obtains a lease through a
configured authentication method. It does not deploy, initialize, unseal, or
administer the credential-source OpenBao instance.

- **AINFRA-CRED-010:** provider profiles MUST identify only non-secret
  connection, authentication-method, and secret-reference metadata; static
  authentication values MUST use an external native mechanism.
- **AINFRA-CRED-011:** SOPS plaintext MUST NOT be written persistently. When an
  engine requires a path, ainfra MUST use an owner-only ephemeral file or FIFO,
  unlink it at the earliest supported point, and register cleanup for normal,
  error, signal, and recovery paths.
- **AINFRA-CRED-012:** OpenBao leases MUST be requested with the narrowest
  configured policy and lifetime, renewed only while the consuming operation
  requires them, and revoked or allowed to expire after use.
- **AINFRA-CRED-013:** an OpenBao credential source MUST be independent from
  all infrastructure resources owned by the deployment that consumes it.

## Acquisition and delivery pipeline

Credential handling follows a fixed pipeline:

1. Resolve the reviewed deployment's symbolic binding.
2. Authenticate to the configured external provider without exposing its
   authentication material to templates.
3. Acquire the credential just in time on the machine running ainfra.
4. Deliver it only to the engine and lifecycle phase that declared the slot.
5. Observe expiry and renew only where the provider and operation require it.
6. Remove ephemeral delivery objects and release or revoke leases.
7. Retain only redacted, non-secret provenance and outcome evidence.

Supported delivery forms are engine-native environment entries, protected
temporary files or FIFOs, stdin where the engine has a documented protocol,
agents or sockets, workload identity, and direct provider APIs. A credential
value MUST NOT be passed in argv. Environment delivery is process scoped: the
value is provided directly to the engine child, not exported into a user's
shell or unrelated child processes.

ainfra MAY run on a host, in a development container, in CI, or as a
Kubernetes job. Acquisition always happens in that execution environment. A
host-side tool such as aibox may inject the SOPS-decrypted credential or key
into the ainfra process without making ainfra depend on aibox.

- **AINFRA-CRED-020:** acquisition MUST occur just in time where ainfra and its
  child engines execute; ainfra MUST NOT copy credentials to a deployment
  target merely to invoke a local engine.
- **AINFRA-CRED-021:** a credential MUST be visible only to declared consumers
  and phases. The default parent environment MUST NOT be inherited wholesale.
- **AINFRA-CRED-022:** delivery MUST use the consuming engine's documented
  native mechanism and MUST preserve direct-tool operability.
- **AINFRA-CRED-023:** a failure to acquire, renew, deliver, or clean up a
  required credential MUST be explicit and MUST stop the affected mutation.
- **AINFRA-CRED-024:** ainfra MUST prevent credential values from entering
  configuration, source locks, plans, argv, retained logs, run events,
  evidence, standardized results, or diagnostic output.

## Reviewed-plan binding and rotation

The plan record binds each required slot to non-secret acquisition identity:
the slot and binding name, provider kind, source identity, provider-exposed
version or lease identity, intended principal and scope when available, and
delivery form. It never records the credential value or a digest derived from
that value.

Apply resolves the same symbolic binding again. A changed source identity,
principal, scope, or incompatible version invalidates the reviewed plan.
Normal credential rotation MAY remain valid only when the provider exposes a
stable verified identity and policy explicitly permits renewal or rotation
without changing authority. Destroy reacquires required credentials under the
same rules; it never depends on retained plaintext from apply.

- **AINFRA-CRED-030:** plan evidence MUST bind non-secret credential provenance
  sufficiently to detect substitution without hashing or retaining the value.
- **AINFRA-CRED-031:** apply and destroy MUST fail closed when credential
  authority changes beyond the reviewed binding or cannot be verified.
- **AINFRA-CRED-032:** renewal and rotation MUST NOT silently widen principal,
  provider, account, project, role, policy, or permission scope.

## Supported scenarios

| Execution scenario | Recommended initial source | Delivery | Preconditions and limitations |
|---|---|---|---|
| Interactive local host | SOPS; external OpenBao for stronger policy | Process environment, protected ephemeral file/FIFO, agent/socket, or native API | Local key access for SOPS, or external identity and network access for OpenBao. |
| Interactive development container | Host- or container-side SOPS; external OpenBao | Direct injection into the ainfra process and then its engine child; protected file/socket where required | The decryption tool and key access must exist where acquisition occurs. Container images must not contain keys. |
| Remote execution host | External OpenBao preferred; SOPS only with a deliberately provisioned remote key | Engine-native delivery on the remote executor | ainfra runs remotely through a secure execution channel; plaintext is not staged from a local deployment bundle. |
| Headless local or CI runner | External OpenBao or user-managed workload identity | Short-lived process environment, agent/socket, file, or native API | Non-interactive authentication and renewal are mandatory. SOPS is acceptable only when its key is supplied by a protected machine identity. |
| Kubernetes execution job | External OpenBao or platform workload identity | Agent/sidecar socket or file, projected identity, process environment, or native API | Kubernetes Secret objects are optional user-managed compatibility, not a required ainfra secret store. |
| Confidential-computing executor | Future KBS/attestation-aware provider | Release only after verified attestation, then native delivery | Deferred to the confidential-computing phase; Phase 10 does not claim live attestation. |

External OpenBao is available in every scenario when its prerequisites are
met; it is not restricted to headless operation. SOPS is also possible in
headless execution when a protected non-interactive key source exists, but it
does not by itself provide dynamic leases, policy enforcement, or renewal.

## Future providers

Phase 23 adds providers only through the same bounded capability contract.
Candidates include other Vault-compatible services, cloud secret managers,
KMS-backed decryption, and OIDC credential brokers. Attestation-gated KBS
integration remains part of Phase 24. Compatibility is capability based:
similar HTTP shapes do not justify claiming one universal secrets API.

Provider additions MUST document authentication, acquisition, identity and
scope binding, expiry, renewal, revocation, delivery compatibility, audit
evidence, failure behavior, and conformance tests. Provider-specific concepts
remain in adapters and operator configuration; they MUST NOT enter templates,
the core lifecycle, or standardized infrastructure results.
