---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260821_1517-NimbleSea-generate-ephemeral-ssh-and-k3s-credentials
  created: '2026-08-21T15:17:52+00:00'
spec:
  title: Generate ephemeral SSH and K3s credentials for Phase 9 certification
  state: accepted
  decision: The Phase 9 certification workflow will generate a fresh ephemeral Ed25519
    SSH keypair and a cryptographically random ephemeral K3s cluster token inside
    the protected run directory. It will use them only for the disposable certification
    lifecycle and destroy them with the protected local certification state after
    independent Hetzner teardown verification. The operator remains responsible only
    for provider access and the externally managed Cloudflare fixture credentials.
  context: Phase 9 live certification needs SSH access to disposable Hetzner hosts
    and a K3s cluster token. The operator has clarified that these are run-scoped
    certification credentials and will not provide them.
  rationale: Generation within the protected, disposable certification boundary avoids
    requiring persistent operator SSH material, keeps the credentials scoped to one
    lifecycle, and satisfies the template rule that credentials are external to template
    and infrastructure generation. The ainfra template itself still does not generate
    or own secrets.
  alternatives:
  - option: Require operator-provided SSH keys and K3s token
    rejected_because: The operator explicitly declined to provide persistent credentials
      for a disposable certification run.
  - option: Reuse shared credentials across certification runs
    rejected_because: This expands credential lifetime and blast radius without providing
      certification value.
  consequences: The certification runner must create the key and token under umask
    077, avoid logging private material, inject only the public SSH key into OpenTofu,
    pass secrets through protected runtime files, retain them only while recovery
    remains necessary, and securely remove the run directory after teardown verification.
    Container restart requires regenerating a new run-scoped pair and token unless
    a protected incomplete-run directory is intentionally recovered.
  related_workitems:
  - BACK-20260821_0208-TidyBird-implement-release-phase-nine-production-template
  decided_at: '2026-08-21T15:17:52+00:00'
---
