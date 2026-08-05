# 9. Documentation and AI agents

## Documentation system

The top-level `docs/` remains user-oriented product documentation. This
top-level `spec/` tree is the normative product and engineering specification.
User documentation MAY summarize requirements but MUST link back when precise
contract behavior matters.

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
- complete template-authoring guide.

No documentation may imply that ainfra publishes a container image or supports
Windows binaries.

## AI-agent authoring package

The repository MUST provide a concise, self-contained instruction entry point
for AI agents creating templates. It SHOULD contain:

1. objective and non-goals;
2. required directory tree;
3. manifest schema and example;
4. native variable rules;
5. standard output schema;
6. security checklist;
7. required documentation sections;
8. local validation commands;
9. disposable live acceptance protocol;
10. completion checklist and evidence format.

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
