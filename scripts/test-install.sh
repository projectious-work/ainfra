#!/usr/bin/env bash

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
version=9.8.7
base="ainfra_${version}_linux_amd64"

test_root=$(mktemp -d "${TMPDIR:-/tmp}/ainfra-installer-test.XXXXXX")
cleanup() {
  rm -rf -- "$test_root"
}
trap cleanup EXIT

fixture="$test_root/fixture"
mkdir -p "$test_root/bin" "$test_root/install" "$fixture/$base"

cat >"$fixture/$base/ainfra" <<EOF
#!/bin/sh
printf '%s\n' 'ainfra $version'
EOF
chmod +x "$fixture/$base/ainfra"
tar -czf "$fixture/$base.tar.gz" -C "$fixture" "$base"
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$fixture" && sha256sum "$base.tar.gz" >checksums.sha256)
else
  (cd "$fixture" && shasum -a 256 "$base.tar.gz" >checksums.sha256)
fi

cat >"$test_root/bin/uname" <<'EOF'
#!/bin/sh
case "$1" in
  -s) printf '%s\n' Linux ;;
  -m) printf '%s\n' x86_64 ;;
  *) exit 2 ;;
esac
EOF

cat >"$test_root/bin/curl" <<'EOF'
#!/bin/sh
output=
url=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) output=$2; shift 2 ;;
    --proto|--proto-redir|-w) shift 2 ;;
    -*) shift ;;
    *) url=$1; shift ;;
  esac
done
name=${url##*/}
case "$url" in
  *api.test*) printf '%s\n' '[{"tag_name":"v9.8.7"}]'; exit 0 ;;
esac
cp "$AINFRA_TEST_FIXTURE/$name" "$output"
EOF

chmod +x "$test_root/bin/uname" "$test_root/bin/curl"
PATH="$test_root/bin:$PATH" \
  VERSION="$version" \
  INSTALL_DIR="$test_root/install" \
  VERIFY_SIGNATURE=0 \
  AINFRA_TEST_FIXTURE="$fixture" \
  "$repo_root/scripts/install.sh"

"$test_root/install/ainfra" version | grep -F "$version" >/dev/null

PATH="$test_root/bin:$PATH" \
  INSTALL_DIR="$test_root/install-latest" \
  VERIFY_SIGNATURE=0 \
  AINFRA_RELEASE_API=https://api.test/releases \
  AINFRA_TEST_FIXTURE="$fixture" \
  "$repo_root/scripts/install.sh" >/dev/null
"$test_root/install-latest/ainfra" version | grep -F "$version" >/dev/null

cp "$fixture/checksums.sha256" "$test_root/checksums.original"
cp "$fixture/$base.tar.gz" "$test_root/"
sed 's/^[0-9a-f]/0/' "$fixture/checksums.sha256" \
  >"$test_root/checksums.sha256"
if PATH="$test_root/bin:$PATH" \
  VERSION="$version" \
  INSTALL_DIR="$test_root/rejected" \
  VERIFY_SIGNATURE=0 \
  AINFRA_TEST_FIXTURE="$test_root" \
  "$repo_root/scripts/install.sh" >/dev/null 2>&1; then
  printf '%s\n' 'installer accepted an invalid checksum' >&2
  exit 1
fi

printf '%s\n' 'installer tests passed'
