# Phase 6: Destruction, recovery, and hardening

Status: in progress

Phase 6 closes the destructive lifecycle and retained-evidence boundaries. It
does not declare shipment: the roadmap moves to `shipped` only after
`v1.0.0-alpha.6` is published and independently verified.

## Requirement disposition

| Requirement area | Normative IDs | Planned evidence |
| --- | --- | --- |
| Reviewed destruction | AINFRA-DESTROY-001..004, AINFRA-SEC-032..034 | Exact destroy-intent plan loading, complete binding revalidation, shell-free `tofu apply <plan>`, durable lifecycle evidence, and certified-template teardown guidance. |
| Recovery and resumption | AINFRA-APPLY-004..005, AINFRA-ANS-005 | Typed interruption outcomes, inspection-required events, doctor/status guidance, and refusal to automatically repeat a mutating stage. |
| Retained run browsing | AINFRA-LOGS-001..007 | Version-checked retained evidence, source/error selection, explicit raw controls, and separation from operational sinks. |
| Operational logging | AINFRA-LOG-001..005 | Shared pre-sink redaction, deterministic private sinks, surfaced sink failures, and correlated concurrent records. |
| Security hardening | AINFRA-SEC-001..006, 010..015, 020..027 | Revalidated source/cache/executable/input bindings, hostile cache and path fixtures, chunk-safe redaction, private evidence, and no shell or state inspection. |
| Verification | AINFRA-TEST requirements and Phase 6 roadmap | Unit, integration, black-box, schema, documentation, and negative security fixtures followed by a full normative sweep. |

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
