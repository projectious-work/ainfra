# Acceptance matrix

| Requirement | Milestone 0 evidence | Later evidence |
|---|---|---|
| Versioned strict contracts | Schema and fixture tests | Compatibility tests |
| Six wrapper commands | CLI surface and guarded exits | Adapter tests |
| No template hardcoding in wrapper | Contract-driven loader | Two fixtures |
| Pinned Hetzner template | Contract fields | OpenTofu lock and tests |
| Secure SSH and keys | Schema prohibition | Plan and host tests |
| Safe remote state | Capability representation | Policy and integration |
| Secret-free output | Strict output schema | Redaction and scans |
| Local validation gates | `scripts/validate-all` | Tool-specific gates |
| Disposable lifecycle | Documented final gate | Approved live evidence |
| Clear portfolio boundaries | README and architecture | Review checklist |
| No unsafe donor behavior | Security invariants | Negative plan tests |
| No GitHub workflows | Repository-policy test | Release audit |

Every issue criterion must eventually map to an automated local check or a
documented manual verification with sanitized evidence.
