package doctor_test

import (
	"context"
	"testing"

	"github.com/projectious-work/ainfra/internal/doctor"
)

func TestDeploymentRegistryReportsValidatedFacts(t *testing.T) {
	t.Parallel()
	report := doctor.DeploymentRegistry(doctor.DeploymentInput{
		Name: "example", Root: "/deployment",
		ManifestPath: "/deployment/ainfra.yaml", NativeFiles: 3, RuntimeSafe: true,
	}).Run(context.Background(), doctor.ScopeDeployment, doctor.Input{}, doctor.Capabilities{})
	if report.Summary.Pass != 4 || len(report.Findings) != 4 {
		t.Fatalf("report = %#v", report)
	}
	for _, finding := range report.Findings {
		if finding.Scope != "deployment" || finding.Status != "pass" ||
			finding.Path == "" || finding.Message == "" {
			t.Fatalf("incomplete finding: %#v", finding)
		}
	}
}
