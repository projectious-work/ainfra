#!/bin/sh
set -eu

template_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf -- "$test_root"' EXIT HUP INT TERM

cp -R "$template_root/tofu" "$test_root/tofu"
cp -R "$template_root/ansible" "$test_root/ansible"
cp -R "$template_root/examples" "$test_root/examples"
cp -R "$template_root/tests/fixtures" "$test_root/fixtures"

tofu -chdir="$test_root/tofu" fmt -check
tofu -chdir="$test_root/tofu" init -input=false
tofu -chdir="$test_root/tofu" validate
tofu -chdir="$test_root/tofu" plan \
  -input=false \
  -var-file="$test_root/examples/minimal/terraform.tfvars" \
  -out=plan.tfplan
tofu -chdir="$test_root/tofu" apply -input=false plan.tfplan
tofu -chdir="$test_root/tofu" output -json ainfra_inventory \
  | jq -S . >"$test_root/output.json"
jq -S . "$test_root/fixtures/output.json" >"$test_root/expected-output.json"
cmp "$test_root/expected-output.json" "$test_root/output.json"

ansible-playbook \
  -i "$test_root/fixtures/inventory.yaml" \
  -e "@$test_root/examples/minimal/ansible-vars.yaml" \
  "$test_root/ansible/site.yml"
ansible-playbook \
  --check \
  -i "$test_root/fixtures/inventory.yaml" \
  -e "@$test_root/examples/minimal/ansible-vars.yaml" \
  "$test_root/ansible/site.yml"

tofu -chdir="$test_root/tofu" plan \
  -destroy \
  -input=false \
  -var-file="$test_root/examples/minimal/terraform.tfvars" \
  -out=destroy.tfplan
tofu -chdir="$test_root/tofu" apply -input=false destroy.tfplan
