---
apiVersion: processkit.projectious.work/v2
kind: WorkItem
metadata:
  id: BACK-20260810_1733-GentleGlade-automate-keyless-sigstore-release-signing
  created: '2026-08-10T17:33:20+00:00'
  updated: '2026-08-10T19:13:17+00:00'
spec:
  title: Automate keyless Sigstore release signing
  state: done
  type: task
  priority: high
  description: Extend the local ainfra release tooling to check Cosign availability,
    sign the published checksum manifest keylessly as info@projectious.work through
    GitHub/Sigstore, verify the bundle with pinned identity and issuer, preserve dry-run
    non-publishing semantics, and document/test the workflow.
  started_at: '2026-08-10T17:33:24+00:00'
  completed_at: '2026-08-10T19:13:17+00:00'
---

## Transition note (2026-08-10T17:33:24+00:00)

Implementation started.


## Transition note (2026-08-10T17:37:25+00:00)

Implementation and offline verification are complete; awaiting the owner-run host OIDC preflight to confirm the production certificate issuer.


## Transition note (2026-08-10T19:13:17+00:00)

Owner-run production preflight returned Verified OK for certificate identity info@projectious.work and issuer https://github.com/login/oauth. Automated verifier, documentation, and tests pin these exact claims.
