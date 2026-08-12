---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260810_1724-TenderSpark-use-keyless-sigstore-signing-for-ainfra
  created: '2026-08-10T17:24:13+00:00'
  updated: '2026-08-10T17:29:12+00:00'
spec:
  title: Use keyless Sigstore signing for ainfra releases
  state: superseded
  decision: Sign ainfra release artifacts keylessly with Cosign using the public identity
    ainfra@projectious.work, Sigstore's interactive OIDC issuer https://oauth2.sigstore.dev/auth,
    GitHub as the upstream login provider, Fulcio certificates, and Rekor transparency
    logging.
  context: The project needs an auditable signing identity for locally executed releases
    while hosted release automation remains prohibited.
  rationale: A dedicated public release identity avoids exposing a private personal
    address and binds signatures to an OIDC-authenticated identity without maintaining
    a long-lived signing key.
  alternatives:
  - option: Encrypted project Cosign key
    reason_not_chosen: Requires secure long-term private-key custody and rotation.
  - option: GitHub Actions workload identity
    reason_not_chosen: Conflicts with the current prohibition on hosted release automation.
  consequences: The release operator must authenticate interactively through Sigstore
    using a GitHub account whose verified email claim is ainfra@projectious.work.
    The identity and signing event become public in the certificate/transparency record.
    Verification must pin both the email identity and Sigstore OIDC issuer.
  decided_at: '2026-08-10T17:24:13+00:00'
  superseded_by: DEC-20260810_1729-LucidHaven-use-info-projectious-work-for-ainfra
---
