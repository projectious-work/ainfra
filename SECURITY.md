# Security Policy

## Supported versions

ainfra is being rewritten for v1. Security and correctness fixes are released
on the latest published version line; unsupported historical lines do not
receive parallel fixes.

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older releases | No |

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability or exposed secret.
Use GitHub's private vulnerability reporting form:

https://github.com/projectious-work/ainfra/security/advisories/new

Include the affected ainfra version, operating system, reproduction steps,
impact, and any suggested mitigation. Remove live tokens, private keys,
personal data, and customer data from reports and attachments.

The maintainers will acknowledge a complete report within seven calendar days,
triage severity and affected versions, and coordinate disclosure after a fix
is available. Response and release timing depend on severity and
reproducibility.

## Scope

Security reports may cover the Go CLI, OpenTofu and Ansible orchestration,
configuration and lock contracts, release binaries, documentation deployment,
and ainfra-owned credential or isolation behavior.

The normative trust boundaries and security requirements are maintained in
[`spec/doc/v1/05-security-and-trust.md`](spec/doc/v1/05-security-and-trust.md).
