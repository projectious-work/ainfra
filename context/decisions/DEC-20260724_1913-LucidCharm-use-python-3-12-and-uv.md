---
apiVersion: processkit.projectious.work/v2
kind: DecisionRecord
metadata:
  id: DEC-20260724_1913-LucidCharm-use-python-3-12-and-uv
  created: '2026-07-24T19:13:57+00:00'
spec:
  title: Use Python 3.12 and uv for the ainfra v0.1 wrapper
  state: accepted
  decision: Implement the ainfra v0.1 command-line wrapper in Python 3.12, package
    and execute it with uv, and isolate public contracts and execution adapters so
    a future implementation can preserve the CLI and file formats.
  context: The initial ainfra wrapper primarily performs schema validation, safety-policy
    enforcement, subprocess orchestration, output sanitization, and diagnostics. The
    project needs a low-cost milestone-one implementation without committing the public
    contract to one implementation language.
  rationale: Python minimizes bootstrap time for validation and orchestration work,
    has mature JSON Schema and testing support, and makes fake-executable integration
    tests straightforward. The architectural boundary is the versioned CLI and schemas,
    not the implementation language.
  alternatives:
  - option: Implement v0.1 in Rust
    rejected_because: A portable binary is attractive, but compilation and ecosystem
      setup add cost before the contracts and safety behavior have been proven.
  - option: Use shell scripts as the wrapper
    rejected_because: Structured validation, safe secret handling, redaction, and
      cross-platform testing would become brittle.
  consequences: The repository will require Python 3.12 and uv for wrapper development
    and execution. Packaging and adapters must avoid leaking Python-specific details
    into public contracts. A future Rust rewrite remains possible but must preserve
    behavior and compatibility.
  decided_at: '2026-07-24T19:13:57+00:00'
---
