# Conformance tests

The fixtures cover the standardized output, deterministic inventory, and final
Ansible Runner statistics expected from the minimal deployment.

A complete validation executes these steps from `examples/minimal/`:

1. validate `ainfra.yaml` and `ainfra-template.yaml` against the v1 schemas;
2. run `tofu fmt -check`, `tofu init`, `tofu validate`, and `tofu plan`;
3. apply the saved plan and compare `ainfra_inventory` with `output.json`;
4. compare generated inventory with `inventory.yaml`;
5. run `ansible-runner` normally and in check mode;
6. require the check-mode stats represented by `ansible-stats.json`; and
7. apply a reviewed destroy plan.

Run `./validate.sh` to exercise the available native-tool portion in a
disposable directory. It requires `tofu`, `ansible-playbook`, and `jq`. The Go
CLI implementation will add schema, ainfra inventory-generation, Runner-event,
saved-plan-binding, and command-surface checks when those lifecycle commands
exist.
