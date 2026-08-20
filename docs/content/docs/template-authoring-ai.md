---
title: AI template-authoring guide
weight: 65
---

This is the concise execution guide for an AI agent authoring an ainfra
template. It is a local-conformance workflow, not permission to provision a
live environment.

## Objective

Produce a portable template that passes the Phase 8 authoring checks and whose
native OpenTofu and Ansible checks can run from a clean room. Do not create
provider resources, expose secrets, or claim production certification.

## Required deliverables

Copy the
[reference template](https://github.com/projectious-work/ainfra/tree/v1.x-dev/spec/examples/v1/template-example),
then retain and complete:

```text
ainfra-template.yaml
README.md
docs/variables.md
tofu/versions.tf
tofu/variables.tf
tofu/outputs.tf
ansible/
tests/README.md
tests/validate.sh
tests/fixtures/output.json
```

Document OpenTofu variables, Ansible variables, cross-variable rules and
examples in that order. Declare secret inputs as sensitive and ensure no
secret value is checked into the template, fixture or evidence.

## Completion checklist

- [ ] Manifest identity, version, supported tools and input/output contracts
      are accurate.
- [ ] The README explains prerequisites, architecture, variables, network,
      cost, failure behavior, teardown, compatibility and validation.
- [ ] `docs/variables.md` has the required headings and complete tables.
- [ ] A clean-room fixture exists and `tests/validate.sh` validates it.
- [ ] `ainfra doctor template . --format json` has no `fail` findings.
- [ ] `./tests/validate.sh` passes using compatible native tools.
- [ ] Any proposal to create billable resources has explicit, immediate
      lifecycle approval and records sanitized evidence afterwards.

The doctor command intentionally marks native tool execution as delegated.
Treat a skip for that check as an instruction to run the template test script,
not as a conformance success by itself.
