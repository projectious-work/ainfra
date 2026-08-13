---
title: References
weight: 80
---

## Doctor commands

```text
ainfra doctor [TARGET]
ainfra doctor all [TARGET]
ainfra doctor environment
ainfra doctor deployment [TARGET]
ainfra doctor template [LOCAL_TEMPLATE]
ainfra doctor run [TARGET]
```

Bare `doctor` is an exact alias of `doctor all`. The aggregate selects the
applicable environment, deployment, already-resolved template, and latest-run
checks. Missing optional evidence is `skip`, never `pass`.

Every finding has a stable code, scope, check ID, status, severity, message,
component or path, and reconciliation state. Failed and skipped checks also
provide a next action. `--format json` emits exactly one
`ainfra.result/v1` object on stdout; diagnostics and confirmation prompts do
not contaminate JSON stdout.

## Local reconciliation

`--reconcile` can repair only registered, ainfra-owned local artifacts. The
currently registered repair creates the deployment `.ainfra` directory or
restricts it to owner-only mode (`0700`). It does not edit manifests, native
inputs, templates, state, credentials, or remote infrastructure.

Interactive use prints the complete plan and asks for confirmation. Automation
must provide all three options:

```sh
ainfra doctor deployment path/to/deployment \
  --reconcile --non-interactive --yes
```

ainfra locks the deployment, rechecks plan preconditions, applies eligible
actions, and reruns affected checks. Failed actions remain in the result as
`failed` or `still_failing`; stale plans are refused before writing.

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | Successful command; warnings and skips may still be present. |
| `1` | Operation, child-tool, rendering, or reconciliation failure. |
| `2` | Invalid invocation, configuration, or input contract. |
| `3` | Missing or incompatible required dependency. |
| `4` | Security or trust refusal. |
| `5` | Stale reviewed-plan binding (reserved for lifecycle commands). |
| `6` | Interrupted or ambiguous mutation (reserved for lifecycle commands). |

Machine-output schemas and examples are maintained under `spec/schemas/v1/`
and `spec/examples/v1/machine-output/`.
