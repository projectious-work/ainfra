# ainfra compatibility corpus

This directory defines the language-neutral behavior that the Python and Rust
implementations must share during the rewrite.

The Python implementation is the executable oracle until a case is explicitly
marked as implemented in Rust. A Rust command is not considered compatible
merely because it accepts the same arguments.

## Case format

CLI cases live in `cli/cases.yaml`. Each case declares:

- `name`: stable case identifier;
- `args`: arguments after the `ainfra` executable;
- `expect.exit`: process exit code;
- `expect.stdout_contains`: stable stdout fragments;
- `expect.stderr_contains`: stable stderr fragments.

`<ROOT>` is replaced with the absolute checkout root by the test harness.
Paths, randomly generated plan IDs, timestamps, local executable versions, and
parser-specific whitespace are not compatibility fields unless a case says
otherwise.

## Compatibility rules

- Assert stable identifiers and semantic fragments, not Python exception text.
- Keep `v1alpha1` schemas and policy identifiers unchanged during the port.
- Never normalize away security-relevant values such as digests, operations,
  template identities, or approval IDs.
- Run the same cases against Rust as each command becomes implemented.
- Keep legacy records literal. Substitute only documented placeholders such as
  `<ROOT>` and recomputed fixture digests during setup.

The initial executable slice covers help, version, validation, and refusal of
an unknown reviewed plan. Further cases are added before their corresponding
Rust behavior is ported.
