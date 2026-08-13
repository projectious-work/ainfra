package app

import (
	"context"
	"fmt"

	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/doctor"
	"github.com/projectious-work/ainfra/internal/output"
	templatecontract "github.com/projectious-work/ainfra/internal/template"
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
	findings = append(findings, resolvedTemplateFindings(deployment.Result.Findings)...)
	findings = append(findings, run.Result.Findings...)
	doctor.SortFindings(findings)
	summary := summarizeDoctorFindings(findings)
	return DoctorEnvironmentResponse{
		Result: output.Doctor{
			Scope: "all", Summary: summary, Findings: findings,
			EffectiveConfiguration: environment.Result.EffectiveConfiguration,
		},
		Format: environment.Format, OutputStyle: environment.OutputStyle,
		Color:              environment.Color,
		ReconciliationPlan: deployment.ReconciliationPlan,
	}, nil
}

func resolvedTemplateFindings(deploymentFindings []diagnostic.Diagnostic) []diagnostic.Diagnostic {
	cachePath := ""
	for _, finding := range deploymentFindings {
		if finding.Check == "template.cache" && finding.Status == "pass" {
			cachePath = finding.Path
		}
	}
	if cachePath == "" {
		return []diagnostic.Diagnostic{{
			Code: "AINFRA-E2601", Severity: diagnostic.SeverityInfo,
			Check: "template.resolved-source", Scope: "template", Status: "skip",
			Component: "template", Reconciliation: "not_available",
			Message:    "no verified locked template is available for this deployment",
			NextAction: "Run 'ainfra template lock' and resolve deployment doctor failures.",
		}}
	}
	contract, err := templatecontract.LoadMaterialized(cachePath)
	if err != nil {
		return []diagnostic.Diagnostic{{
			Code: "AINFRA-E2602", Severity: diagnostic.SeverityError,
			Check: "template.resolved-source", Scope: "template", Status: "fail",
			Component: "template", Path: cachePath, Reconciliation: "not_available",
			Message:    fmt.Sprintf("verified cache does not contain a valid template contract: %s", err),
			NextAction: "Correct the template source and run 'ainfra template update'.",
		}}
	}
	report := doctor.TemplateRegistry(doctor.TemplateInput{
		Name: contract.Name, Version: contract.Version, Root: contract.Root,
		HasAnsible: contract.Ansible != nil, Inventory: contract.Inventory,
	}).Run(context.Background(), doctor.ScopeTemplate, doctor.Input{}, doctor.Capabilities{})
	return report.Findings
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
