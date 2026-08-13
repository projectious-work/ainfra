package template_test

import (
	"os"
	"path/filepath"
	"testing"

	templatecontract "github.com/projectious-work/ainfra/internal/template"
)

func TestLoadInfrastructureOnlyTemplate(t *testing.T) {
	t.Parallel()
	root := templateRoot(t, "infra-template", false)
	contract, err := templatecontract.Load(root)
	if err != nil {
		t.Fatalf("load template: %v", err)
	}
	if contract.Name != "infra-template" || contract.Ansible != nil ||
		contract.Inventory != "none" {
		t.Fatalf("contract = %#v", contract)
	}
}

func TestLoadAnsibleTemplateRequiresNativeFiles(t *testing.T) {
	t.Parallel()
	root := templateRoot(t, "ansible-template", true)
	if _, err := templatecontract.Load(root); err != nil {
		t.Fatalf("load template: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "ansible", "requirements.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := templatecontract.Load(root); err == nil {
		t.Fatal("missing native requirements unexpectedly accepted")
	}
}

func TestLoadRejectsNameMismatchUnknownFieldAndRuntimeState(t *testing.T) {
	t.Parallel()
	t.Run("name mismatch", func(t *testing.T) {
		root := templateRoot(t, "actual-name", false)
		rewriteManifest(t, root, false, "different-name", "")
		if _, err := templatecontract.Load(root); err == nil {
			t.Fatal("name mismatch unexpectedly accepted")
		}
	})
	t.Run("unknown field", func(t *testing.T) {
		root := templateRoot(t, "unknown-field", false)
		rewriteManifest(t, root, false, "unknown-field", "unexpected: true\n")
		if _, err := templatecontract.Load(root); err == nil {
			t.Fatal("unknown field unexpectedly accepted")
		}
	})
	t.Run("runtime state", func(t *testing.T) {
		root := templateRoot(t, "runtime-state", false)
		if err := os.Mkdir(filepath.Join(root, ".terraform"), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := templatecontract.Load(root); err == nil {
			t.Fatal("runtime state unexpectedly accepted")
		}
	})
}

func templateRoot(t *testing.T, name string, ansible bool) string {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, name)
	for _, directory := range []string{
		root, filepath.Join(root, "tofu"), filepath.Join(root, "docs"),
		filepath.Join(root, "examples", "minimal"),
	} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{
		"docs/variables.md",
	} {
		writeTemplateFile(t, filepath.Join(root, file), "safe example\n")
	}
	readme := "# safe\n\n```sh\ntofu validate\n"
	if ansible {
		readme += "ansible-playbook site.yml\n"
	}
	readme += "```\n"
	writeTemplateFile(t, filepath.Join(root, "README.md"), readme)
	writeTemplateFile(t, filepath.Join(root, "examples", "minimal", "terraform.tfvars"), "opaque\n")
	if ansible {
		writeTemplateFile(t, filepath.Join(root, "examples", "minimal", "ansible-vars.yaml"), "opaque: true\n")
	}
	if ansible {
		if err := os.Mkdir(filepath.Join(root, "ansible"), 0o700); err != nil {
			t.Fatal(err)
		}
		writeTemplateFile(t, filepath.Join(root, "ansible", "site.yml"), "---\n")
		writeTemplateFile(t, filepath.Join(root, "ansible", "requirements.yml"), "---\ncollections: []\n")
	}
	rewriteManifest(t, root, ansible, name, "")
	writeExampleManifest(t, root, ansible)
	return root
}

func writeExampleManifest(t *testing.T, root string, ansible bool) {
	t.Helper()
	ansibleInputs := ""
	if ansible {
		ansibleInputs = `    ansible:
      variableFiles: [ansible-vars.yaml]
`
	}
	contents := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../..
  inputs:
    tofu:
      variableFiles: [terraform.tfvars]
` + ansibleInputs
	writeTemplateFile(t, filepath.Join(root, "examples", "minimal", "ainfra.yaml"), contents)
}

func rewriteManifest(t *testing.T, root string, ansible bool, name, extra string) {
	t.Helper()
	ansibleBlock := ""
	inventory := "none"
	if ansible {
		ansibleBlock = `    ansible:
      directory: ansible
      playbook: site.yml
      version: ">=2.18.0 <3.0.0"
`
		inventory = "ainfra_inventory"
	}
	contents := `apiVersion: ainfra.projectious.work/v1
kind: Template
metadata:
  name: ` + name + `
  version: 1.0.0
spec:
  engines:
    tofu:
      directory: tofu
      version: ">=1.10.0 <2.0.0"
` + ansibleBlock + `  outputs:
    inventory: ` + inventory + "\n" + extra
	writeTemplateFile(t, filepath.Join(root, "ainfra-template.yaml"), contents)
}

func writeTemplateFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
