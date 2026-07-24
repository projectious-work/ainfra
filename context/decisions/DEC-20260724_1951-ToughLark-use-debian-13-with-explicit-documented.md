---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1951-ToughLark-use-debian-13-with-explicit-documented
  created: '2026-07-24T19:51:12+00:00'
spec:
  title: Use Debian 13 with explicit documented image choices and reviewed major upgrades
  state: accepted
  decision: Default the initial Hetzner template to the official Debian 13 major-release
    image selector. Do not use latest or an unversioned selector and do not perform
    automatic major upgrades. Resolve and expose the exact provider image ID in plan
    evidence, reject deprecated or architecture-incompatible images, apply current
    security updates during initial configuration, and enable unattended security
    updates with an explicit reboot policy. Treat a Debian major-version change as
    a reviewed template-version change with migration and disposable-environment testing.
    Wherever users can select an image, include adjacent comments or equivalent schema/UI
    descriptions enumerating and explaining every allowed option; initially the only
    supported option may be Debian 13.
  context: The Hetzner baseline needs a secure and reproducible OS default without
    relying on provider image IDs that vary by architecture or may be retired. Users
    must also be able to understand every supported image value where they configure
    it.
  rationale: Pinning the major release gives a stable reviewed platform while allowing
    Hetzner to refresh minor-release security content. Capturing the resolved image
    preserves operational evidence without binding the template indefinitely to a
    removable numeric ID. Point-of-choice documentation prevents hidden enums and
    makes constraints discoverable without requiring users to search separate documentation.
  alternatives:
  - option: Pin one numeric Hetzner image ID indefinitely
    rejected_because: Image IDs can be architecture-specific and eventually become
      unavailable after deprecation.
  - option: Use a latest or unversioned Debian selector
    rejected_because: Major-version changes could enter plans without deliberate review
      or migration testing.
  - option: Support several distributions immediately
    rejected_because: It multiplies hardening, networking, package, and idempotence
      test paths before the reference baseline is proven.
  - option: Document allowed images only in the operator guide
    rejected_because: Users should see all valid choices and their meaning where the
      selection is made.
  consequences: Schemas and examples initially expose a constrained Debian 13 choice
    with inline or adjacent descriptions. Tests must cover unknown, deprecated, and
    architecture-incompatible selections. Plans and evidence capture the resolved
    image ID. Adding another OS or major release requires its own hardening validation,
    documentation, and disposable lifecycle tests.
  decided_at: '2026-07-24T19:51:12+00:00'
---
