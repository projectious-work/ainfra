# Clean-room validation

`validate.sh` copies the native OpenTofu module and example to a temporary
directory, checks the documented native variable and absence of external
dependencies, then runs formatting, initialization, validation, plan, apply,
destroy plan and destroy apply without contacting a provider.

The test is intentionally inventory-free, so Ansible and standard output
fixtures are not applicable.
