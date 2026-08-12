---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260810_1804-FluentOak-pin-github-as-ainfra-sigstore-certificate
  created: '2026-08-10T18:04:39+00:00'
spec:
  title: Pin GitHub as ainfra Sigstore certificate issuer
  state: accepted
  decision: Verify ainfra keyless Cosign release signatures with certificate identity
    info@projectious.work and certificate OIDC issuer https://github.com/login/oauth.
  context: An owner-run production Sigstore preflight showed that the interactive
    authorization begins at https://oauth2.sigstore.dev/auth, while the Fulcio certificate
    records the selected upstream GitHub connector as https://github.com/login/oauth.
  rationale: Verification must pin the claims actually embedded in the signing certificate.
    The browser authorization endpoint is not the certificate issuer claim.
  alternatives:
  - option: https://oauth2.sigstore.dev/auth
    reason_not_chosen: Production verification reported an issuer mismatch and disclosed
      https://github.com/login/oauth as the certificate issuer.
  consequences: Release tooling and consumer verification instructions must use https://github.com/login/oauth.
    The signing browser may still open at oauth2.sigstore.dev.
  decided_at: '2026-08-10T18:04:39+00:00'
  supersedes: DEC-20260810_1729-LucidHaven-use-info-projectious-work-for-ainfra
---
