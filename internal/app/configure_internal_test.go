package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/inventory"
	"github.com/projectious-work/ainfra/internal/project"
)

func TestBuildAnsibleEnvironmentEnforcesBoundSSHTrust(t *testing.T) {
	standard := inventory.StandardOutput{SchemaVersion: "1", Hosts: map[string]inventory.Host{
		"host": {Groups: []string{"group"}, Connection: inventory.Connection{Type: "ssh", Address: "host.example", User: "ops", Port: 22}},
	}}
	root := t.TempDir()
	deployment := project.Deployment{}
	if _, err := buildAnsibleEnvironment(nil, root, deployment, standard); err == nil {
		t.Fatal("accepted SSH without known_hosts")
	}
	deployment.SSH.KnownHosts = "known_hosts"
	if err := os.Mkdir(filepath.Join(root, "inputs"), 0o700); err != nil {
		t.Fatal(err)
	}
	knownHosts := filepath.Join(root, "inputs", "known_hosts")
	if err := os.WriteFile(knownHosts, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := buildAnsibleEnvironment(nil, root, deployment, standard); err == nil {
		t.Fatal("accepted empty known_hosts")
	}
	if err := os.WriteFile(knownHosts, []byte("host ssh-ed25519 AAAAfixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environment, err := buildAnsibleEnvironment([]string{"HOME=/private", "ANSIBLE_HOST_KEY_CHECKING=False", "UNSAFE=value"}, root, deployment, standard)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(environment.Values(), "\n")
	if !strings.Contains(joined, "ANSIBLE_HOST_KEY_CHECKING=True") ||
		!strings.Contains(joined, "StrictHostKeyChecking=yes") ||
		!strings.Contains(joined, knownHosts) || strings.Contains(joined, "UNSAFE=") {
		t.Fatalf("environment=%q", joined)
	}
}

func TestBuildAnsibleEnvironmentLocalDoesNotRequireSSHInputs(t *testing.T) {
	standard := inventory.StandardOutput{SchemaVersion: "1", Hosts: map[string]inventory.Host{
		"localhost": {Groups: []string{"local"}, Connection: inventory.Connection{Type: "local"}},
	}}
	environment, err := buildAnsibleEnvironment(nil, t.TempDir(), project.Deployment{}, standard)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(environment.Values(), "\n"), "ANSIBLE_SSH_ARGS") {
		t.Fatal("local inventory received SSH policy")
	}
}
