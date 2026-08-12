---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260810_1729-LucidHaven-use-info-projectious-work-for-ainfra
  created: '2026-08-10T17:29:00+00:00'
  updated: '2026-08-10T18:04:39+00:00'
spec:
  title: Use info@projectious.work for ainfra release signing
  state: superseded
  decision: Sign ainfra release artifacts keylessly with Cosign using identity info@projectious.work,
    authenticated by the existing GitHub user projectious through Sigstore's interactive
    Dex issuer https://oauth2.sigstore.dev/auth, with Fulcio certificates and Rekor
    transparency logging.
  context: The initially selected ainfra@projectious.work address is not associated
    with the projectious GitHub user. The existing verified primary address is info@projectious.work.
  rationale: Fulcio's email-based certificate identity must be supplied by the authenticating
    GitHub account. Using its existing verified primary email makes the intended identity
    reproducible without creating another account.
  alternatives:
  - option: ainfra@projectious.work
    reason_not_chosen: Not associated with the authenticating GitHub user.
  - option: Separate release GitHub account
    reason_not_chosen: Unnecessary additional account and credential lifecycle.
  consequences: Release verification must pin info@projectious.work and https://oauth2.sigstore.dev/auth.
    The email identity and signing event will be public in Sigstore transparency data.
  decided_at: '2026-08-10T17:29:00+00:00'
  supersedes: DEC-20260810_1724-TenderSpark-use-keyless-sigstore-signing-for-ainfra
  superseded_by: DEC-20260810_1804-FluentOak-pin-github-as-ainfra-sigstore-certificate
---
