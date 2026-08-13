package doctor_test

import (
	"context"
	"testing"

	"github.com/projectious-work/ainfra/internal/doctor"
)

func TestTemplateRegistryReportsValidatedFacts(t *testing.T) {
	t.Parallel()
	report := doctor.TemplateRegistry(doctor.TemplateInput{
		Name: "example-template", Version: "1.0.0", Root: "/template",
		HasAnsible: true, Inventory: "ainfra_inventory",
	}).Run(context.Background(), doctor.ScopeTemplate, doctor.Input{}, doctor.Capabilities{})
	if report.Summary.Pass != 4 || len(report.Findings) != 4 {
		t.Fatalf("report = %#v", report)
	}
}
