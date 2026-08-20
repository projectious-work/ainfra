package doctor_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/doctor"
)

func TestTemplateRegistryReportsValidatedFacts(t *testing.T) {
	t.Parallel()
	report := doctor.TemplateRegistry(doctor.TemplateInput{
		Name: "example-template", Version: "1.0.0", Root: "/template",
		HasAnsible: true, Inventory: "ainfra_inventory",
	}).Run(context.Background(), doctor.ScopeTemplate, doctor.Input{}, doctor.Capabilities{})
	if report.Summary.Pass != 4 || report.Summary.Fail != 3 ||
		report.Summary.Skip != 1 || len(report.Findings) != 8 {
		t.Fatalf("report = %#v", report)
	}
}

func TestTemplateRegistryDoesNotRequireInventoryFixtureWhenInventoryIsNone(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, directory := range []string{"tofu", "tests", "docs"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"tofu/variables.tf": "variable \"name\" { type = string }\n",
		"tests/README.md":   "# Clean-room validation\n",
		"tests/validate.sh": "#!/bin/sh\nset -eu\n",
		"docs/variables.md": "## OpenTofu variables\n\n| Name | Type or shape | Required | Default | Valid values and constraints | Sensitive | Description |\n\n## Ansible variables\n\n## Cross-variable rules\n\n## Examples\n",
		"README.md":         "Prerequisites Architecture Variables Network Cost Failure Teardown Compatibility Validation\n",
	}
	for path, contents := range files {
		mode := os.FileMode(0o600)
		if path == "tests/validate.sh" {
			mode = 0o700
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(contents), mode); err != nil {
			t.Fatal(err)
		}
	}
	report := doctor.TemplateRegistry(doctor.TemplateInput{
		Name: "infra-only", Version: "1.0.0", Root: root, Inventory: "none",
	}).Run(context.Background(), doctor.ScopeTemplate, doctor.Input{}, doctor.Capabilities{})
	if report.Summary.Fail != 0 || report.Summary.Pass != 7 || report.Summary.Skip != 1 {
		t.Fatalf("report = %#v", report)
	}
}
