#!/bin/sh
set -eu

template_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
repo_root=$(CDPATH= cd -- "$template_root/../.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf -- "$test_root"' EXIT HUP INT TERM

cp -R "$template_root/tofu" "$test_root/tofu"
cp -R "$template_root/cloud-init" "$test_root/cloud-init"
rm -rf "$test_root/tofu/.terraform"

python3 "$template_root/tests/policy.py"
tofu -chdir="$test_root/tofu" fmt -check
tofu -chdir="$test_root/tofu" init -backend=false -input=false
tofu -chdir="$test_root/tofu" validate

ANSIBLE_LOCAL_TEMP=${ANSIBLE_LOCAL_TEMP:-/tmp/ainfra-ansible-local} \
ANSIBLE_REMOTE_TEMP=${ANSIBLE_REMOTE_TEMP:-/tmp/ainfra-ansible-remote} \
ansible-playbook --syntax-check \
  -i "$template_root/tests/fixtures/inventory.yaml" \
  -e "@$template_root/tests/syntax-vars.yaml" \
  "$template_root/ansible/site.yml"

XDG_CACHE_HOME=${XDG_CACHE_HOME:-/tmp/ainfra-xdg-cache} \
GOMODCACHE=${GOMODCACHE:-/tmp/ainfra-go-modcache} \
GOCACHE=${GOCACHE:-/tmp/ainfra-go-buildcache} \
  go run "$repo_root/cmd/ainfra" doctor template "$template_root" \
  --format json >/dev/null
