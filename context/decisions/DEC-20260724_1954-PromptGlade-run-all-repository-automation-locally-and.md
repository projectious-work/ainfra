---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1954-PromptGlade-run-all-repository-automation-locally-and
  created: '2026-07-24T19:54:39+00:00'
spec:
  title: Run all repository automation locally and prohibit GitHub Actions workflows
  state: accepted
  decision: Do not add or use GitHub Actions or any files under .github/workflows.
    Every validation, test, plan, apply, destroy, evidence, and release-support action
    must be runnable locally through documented scripts or the ainfra CLI. Scripts
    must return reliable non-zero exit statuses suitable for human-operated local
    execution and other explicitly chosen external runners, without depending on GitHub-hosted
    workflow behavior.
  context: The project requires reproducible validation and infrastructure operations
    but does not use GitHub Actions or GitHub workflow automation.
  rationale: Local-first automation keeps the toolchain provider-neutral, inspectable,
    and usable in the same environment where operators review and execute infrastructure
    changes. It also enforces the issue's thin-wrapper and direct-tool principles.
  alternatives:
  - option: Add GitHub Actions for validation only
    rejected_because: The owner explicitly prohibits GitHub Actions and workflows,
      including validation-only use.
  - option: Document raw commands without stable scripts
    rejected_because: That would make complete, repeatable local validation difficult
      and invite gate drift.
  consequences: The repository must maintain comprehensive local entry points such
    as scripts/validate-all and scripts/test-all. Documentation and acceptance criteria
    must never imply hosted CI. Any future hosted or remote runner integration requires
    a new decision that supersedes this one.
  decided_at: '2026-07-24T19:54:39+00:00'
---
