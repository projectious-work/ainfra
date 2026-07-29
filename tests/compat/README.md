# ainfra compatibility corpus

This directory preserves the language-neutral behavior frozen before the Rust
cutover.

The Rust implementation is the production executable. These fixtures retain
the former prototype's stable output fragments, exit codes, and security
boundaries without retaining its runtime.

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

- Assert stable identifiers and semantic fragments, not implementation text.
- Keep `v1alpha1` schemas and policy identifiers unchanged during the port.
- Never normalize away security-relevant values such as digests, operations,
  template identities, or approval IDs.
- Run every retained case against Rust.
- Keep legacy records literal. Substitute only documented placeholders such as
  `<ROOT>` and recomputed fixture digests during setup.

The corpus covers stable command behavior and remains a regression boundary
for future versions.
