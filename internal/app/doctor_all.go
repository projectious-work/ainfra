package app

import (
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/doctor"
	"github.com/projectious-work/ainfra/internal/output"
)

// DoctorAllRequest identifies the deployment used by every applicable check.
type DoctorAllRequest = DoctorDeploymentRequest

// DoctorAll composes all currently applicable read-only diagnostic scopes.
func DoctorAll(
	request DoctorAllRequest,
	options DoctorEnvironmentOptions,
) (DoctorEnvironmentResponse, error) {
	projectPath := request.ProjectPath
	if projectPath == "" {
		projectPath = request.Target
	}
	environment, err := DoctorEnvironment(DoctorEnvironmentRequest{
		ConfigPath: request.ConfigPath, ProjectPath: projectPath,
		Format: request.Format, OutputStyle: request.OutputStyle, Color: request.Color,
	}, options)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	deployment, err := DoctorDeployment(request, options)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	run, err := DoctorRun(request, options)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	findings := append([]diagnostic.Diagnostic{}, environment.Result.Findings...)
	findings = append(findings, deployment.Result.Findings...)
	findings = append(findings, diagnostic.Diagnostic{
		Code: "AINFRA-E2601", Severity: diagnostic.SeverityInfo,
		Check: "template.resolved-source", Scope: "template", Status: "skip",
		Component: "template", Reconciliation: "not_available",
		Message:    "no resolved template is available for this deployment",
		NextAction: "Resolve and lock the template before diagnosing its local contract.",
	})
	findings = append(findings, run.Result.Findings...)
	doctor.SortFindings(findings)
	summary := summarizeDoctorFindings(findings)
	return DoctorEnvironmentResponse{
		Result: output.Doctor{
			Scope: "all", Summary: summary, Findings: findings,
			EffectiveConfiguration: environment.Result.EffectiveConfiguration,
		},
		Format: environment.Format, OutputStyle: environment.OutputStyle,
		Color: environment.Color,
	}, nil
}

func summarizeDoctorFindings(findings []diagnostic.Diagnostic) output.DoctorSummary {
	summary := output.DoctorSummary{}
	for _, finding := range findings {
		switch finding.Status {
		case "pass":
			summary.Pass++
		case "skip":
			summary.Skip++
		case "warning":
			summary.Warning++
		case "fail":
			summary.Fail++
		}
	}
	return summary
}
