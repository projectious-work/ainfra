package contracts_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPhase9TemplateContractAndSecurityPosture(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "templates", "hetzner-kubernetes-baseline")
	read := func(path string) string {
		t.Helper()
		contents, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return string(contents)
	}

	manifest := read("ainfra-template.yaml")
	terraform := ""
	for _, name := range []string{
		"versions.tf", "variables.tf", "locals.tf", "network.tf",
		"firewall.tf", "servers.tf", "outputs.tf",
	} {
		terraform += read(filepath.Join("tofu", name))
	}
	ansible := read("ansible/ansible.cfg") + read("ansible/site.yml") +
		read("ansible/roles/k3s/tasks/main.yml") +
		read("ansible/roles/cloudflared/tasks/main.yml")

	for _, required := range []string{
		"apiVersion: ainfra.projectious.work/v1",
		"name: hetzner-kubernetes-baseline",
		"inventory: ainfra_inventory",
	} {
		if !strings.Contains(manifest, required) {
			t.Errorf("manifest missing %q", required)
		}
	}
	for _, required := range []string{
		`version = "1.64.0"`, `default     = false`,
		`temporary = "true"`, `address = hcloud_server_network`,
		`groups = ["k3s_cluster", "k3s_servers"]`, `backend "local" {}`,
	} {
		if !strings.Contains(terraform, required) {
			t.Errorf("OpenTofu contract missing %q", required)
		}
	}
	for _, prohibited := range []string{
		"random_password", "tls_private_key", "0.0.0.0/0",
	} {
		if strings.Contains(terraform, prohibited) {
			t.Errorf("OpenTofu contains prohibited construct %q", prohibited)
		}
	}
	for _, required := range []string{
		"host_key_checking = True", "v1.36.1+k3s1", "2026.7.2",
		"cloudflare_tunnel_token", "no_log: true",
	} {
		if !strings.Contains(ansible, required) {
			t.Errorf("Ansible contract missing %q", required)
		}
	}
}

func TestPhase9NativeVariablesAreDocumented(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "templates", "hetzner-kubernetes-baseline")
	variables, err := os.ReadFile(filepath.Join(root, "tofu", "variables.tf"))
	if err != nil {
		t.Fatal(err)
	}
	docs, err := os.ReadFile(filepath.Join(root, "docs", "variables.md"))
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`variable "([a-z0-9_]+)"`)
	for _, match := range pattern.FindAllStringSubmatch(string(variables), -1) {
		if !strings.Contains(string(docs), "`"+match[1]+"`") {
			t.Errorf("OpenTofu variable %q is undocumented", match[1])
		}
	}
}

func TestPhase9TofuEngineIsSelfContained(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "templates", "hetzner-kubernetes-baseline")
	servers, err := os.ReadFile(filepath.Join(root, "tofu", "servers.tf"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(servers), "../cloud-init") {
		t.Fatal("OpenTofu engine references cloud-init outside its snapshot boundary")
	}
	if _, err := os.Stat(filepath.Join(root, "tofu", "cloud-config.yaml.tftpl")); err != nil {
		t.Fatalf("OpenTofu engine cloud-init template is unavailable: %v", err)
	}
}
