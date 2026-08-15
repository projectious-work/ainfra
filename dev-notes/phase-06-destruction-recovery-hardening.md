# Phase 6: Destruction, recovery, and hardening

Status: in progress

Phase 6 closes the destructive lifecycle and retained-evidence boundaries. It
does not declare shipment: the roadmap moves to `shipped` only after
`v1.0.0-alpha.6` is published and independently verified.

## Requirement disposition

| Requirement area | Normative IDs | Implementation evidence |
| --- | --- | --- |
| Reviewed destruction | AINFRA-DESTROY-001..004, AINFRA-SEC-032..034 | `internal/app/destroy.go`, intent-bound review in `internal/run/review.go`, lifecycle tests in `internal/app/plan_test.go`, and provider verification guidance in `docs/content/docs/reviewed-plans.md`. |
| Recovery and resumption | AINFRA-APPLY-004..005, AINFRA-ANS-005 | Durable terminal and inspection-required events, typed status recovery, doctor guidance, and exclusive mutation starts in `internal/run` and `internal/app`. |
| Retained run browsing | AINFRA-LOGS-001..007 | Combined chronological evidence, versioned Runner classification, typed unavailable OpenTofu filtering, guarded raw access, and symlink refusal in `internal/app/evidence.go` and its tests. |
| Operational logging | AINFRA-LOG-001..005 | Shared pre-sink redaction, deterministic private rotation, surfaced sink failures, and serialized correlated records in `internal/logging` plus compiled-CLI coverage. |
| Security hardening | AINFRA-SEC-001..006, 010..015, 020..027 | Revalidated bindings, owner-only cache trees, directory-scoped file operations, strict retained-artifact identity, hostile fixtures, and fuzzed chunk-safe redaction. |
| Verification | AINFRA-TEST requirements and Phase 6 roadmap | `scripts/validate-all`, `scripts/test-all`, bounded redaction fuzzing, the documentation build, black-box logging coverage, container-gate tests, `govulncheck`, and `gosec`. |

## Final normative sweep

The implementation sweep completed on 2026-08-15. It closed the remaining
concurrent structured-log integrity, combined retained-timeline, typed
OpenTofu filtering-unavailable, and symlinked evidence gaps. The processkit
release audit reports no errors for either the live or shipped context tree.

Phase 6 remains `in_progress`. Release packaging, host verification, signing,
publication, and independent verification for `v1.0.0-alpha.6` are separate
release gates; only their successful completion may move the roadmap phase to
`shipped`.

## Delivery slices

1. Add exact reviewed destroy-plan execution and durable destructive evidence.
2. Complete interruption diagnosis, status, and safe resumption guidance.
3. Add retained `ainfra logs` browsing with closed source and raw-access
   contracts.
4. Harden redaction, bindings, caches, permissions, and hostile fixtures.
5. Document independent provider-side teardown verification for every
   certified template and run the final Phase 6 requirements sweep.

## Non-goals

- Parsing, copying, migrating, or interpreting OpenTofu state.
- Unreviewed `tofu destroy` or implicit replanning during destruction.
- Treating operational file or syslog sinks as authoritative run evidence.
- Marking Phase 6 shipped before release publication and independent
  verification.
