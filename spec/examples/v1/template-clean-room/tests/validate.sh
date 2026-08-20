#!/bin/sh
set -eu

template_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf -- "$test_root"' EXIT HUP INT TERM

cp -R "$template_root/tofu" "$test_root/tofu"
cp -R "$template_root/examples" "$test_root/examples"

grep -q 'variable "environment_name"' "$template_root/tofu/variables.tf"
grep -q '`environment_name`' "$template_root/docs/variables.md"
if grep -R -E 'provider[[:space:]]+"|source[[:space:]]*=' \
  "$template_root/tofu" >/dev/null
then
  echo "clean-room proof unexpectedly declares an external dependency" >&2
  exit 1
fi

tofu -chdir="$test_root/tofu" fmt -check
tofu -chdir="$test_root/tofu" init -input=false
tofu -chdir="$test_root/tofu" validate
tofu -chdir="$test_root/tofu" plan -input=false \
  -var-file="$test_root/examples/minimal/terraform.tfvars" \
  -out=plan.tfplan
tofu -chdir="$test_root/tofu" apply -input=false plan.tfplan
tofu -chdir="$test_root/tofu" plan -destroy -input=false \
  -var-file="$test_root/examples/minimal/terraform.tfvars" \
  -out=destroy.tfplan
tofu -chdir="$test_root/tofu" apply -input=false destroy.tfplan
