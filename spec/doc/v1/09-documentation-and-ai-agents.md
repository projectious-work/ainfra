## Documentation system

The top-level `docs/` remains user-oriented product documentation. This
top-level `spec/` tree is the normative product and engineering specification.
User documentation MAY summarize requirements but MUST link back when precise
contract behavior matters.

## Documentation follows implementation

ainfra is implemented in small, documented roadmap phases so a maintainer can
trace how the system grew. The normative specification describes intended
behavior; code and tests establish what was implemented; phase notes record how
and why the implementation reached that result; user documentation distills
the verified behavior.

Implementation phase notes live in the top-level `dev-notes/` directory, close
to the source they describe. A note uses the roadmap phase ID in its filename,
for example `dev-notes/phase-02-contracts-and-doctor.md`. It evolves with the
phase rather than being reconstructed after the work is forgotten.

Every phase note identifies:

- roadmap phase, status, scope, and intended reader;
- implemented behavior and important source/package boundaries;
- decisions, alternatives, and deviations from the specification;
- tests and manual validation performed, including known gaps;
- security, compatibility, migration, and operational consequences;
- user and template-author documentation added or changed; and
- unresolved follow-up work linked to issues rather than anonymous TODOs.

A phase is `shipped` only when its implementation, tests, phase note, affected
specification, and user documentation agree. The roadmap entry then links its
`devNote`. User-facing pages MUST describe observed released behavior rather
than planned behavior. If implementation exposes a flawed specification, the
specification is amended explicitly; documentation MUST NOT silently redefine
the contract to match an accidental implementation.

- **AINFRA-DOC-010:** every non-trivial implementation change MUST update the
  active phase note or state why it has no phase-note impact in its review.
- **AINFRA-DOC-011:** phase notes MUST explain decisions and observable results,
  not narrate commits or restate self-evident code.
- **AINFRA-DOC-012:** a roadmap phase with status `shipped` MUST reference an
  existing checked-in `dev-notes/` file.
- **AINFRA-DOC-013:** release review MUST verify code, tests, phase notes,
  specification, user documentation, examples, and schemas for consistency.

## Required user documentation

- installation and prerequisite versions;
- quickstart using the reference template;
- deployment creation and native variable editing;
- local and Git template sources;
- lock/update workflow;
- plan review, apply, deploy, status, recovery, and destroy;
- security and secrets model;
- machine output and exit codes;
- direct-engine recovery;
- optional Dockerfile build and enter instructions;
- supported Linux/macOS platforms;
- release, compatibility, and migration policy;
- console output, logging, and CLI configuration reference;
- complete template-authoring guide; and
- standard `docs/variables.md` authoring and drift-checking guidance.

No documentation may imply that ainfra publishes a container image or supports
Windows binaries.

## AI-agent authoring package

The repository MUST provide a concise, self-contained instruction entry point
for AI agents creating templates. It SHOULD contain:

1. objective and non-goals;
2. required directory tree;
3. manifest schema and example;
4. native variable rules;
5. standard native-variable reference format;
6. standard output schema;
7. security checklist;
8. required documentation sections;
9. local validation commands;
10. disposable live acceptance protocol;
11. completion checklist and evidence format.

- **AINFRA-DOC-001:** an agent instruction MUST not depend on hidden context or
  unstated ainfra implementation knowledge.
- **AINFRA-DOC-002:** all referenced files and commands MUST exist and be
  checked by documentation tests.
- **AINFRA-DOC-003:** examples MUST be internally consistent and validate
  against current schemas.
- **AINFRA-DOC-004:** supported values MUST be enumerated, not described with
  ambiguous phrases such as “and similar.”
- **AINFRA-DOC-005:** an agent MUST be told not to claim live support without
  disposable lifecycle evidence.

## Documentation quality gates

- Markdown formatting/style check;
- internal link and anchor check;
- external link check with controlled retry policy;
- schema validation of fenced/extracted examples where practical;
- command/help drift check;
- generated CLI reference drift check;
- spell check for prose without rewriting identifiers;
- site build for `docs/`;
- SVG validity and accessibility title/description checks.

## Traceability

Implementation changes SHOULD cite requirement IDs. Tests MAY encode IDs in
names or comments when the mapping is otherwise unclear. User documentation
must not be cluttered with every requirement identifier; traceability belongs
in engineering material and generated conformance reports.
