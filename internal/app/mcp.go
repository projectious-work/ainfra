package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
)

// MCPServeRequest contains server-start policy selected by the operator.
// Protocol-specific types belong to the MCP adapter and must not enter app.
type MCPServeRequest struct {
	ProjectPath  string
	Capabilities []string
}

// MCPServeOptions supplies the closed host facts captured before serving.
type MCPServeOptions struct {
	Plan   PlanHostOptions
	Doctor DoctorEnvironmentOptions
}

// MCPServeSession contains the immutable, validated project identity exposed
// to protocol adapters. Request handlers cannot replace this root.
type MCPServeSession struct {
	Project           output.Deployment
	project           project.Deployment
	runsRoot          string
	cacheRoot         string
	environmentDoctor output.Doctor
}

// Status reads sanitized retained lifecycle state for the fixed startup
// project. It does not rediscover a root or reload configuration.
func (session MCPServeSession) Status() (output.Status, error) {
	return statusForDeployment(session.project, session.runsRoot)
}

// DoctorDeployment runs the existing deployment checks against the fixed
// startup snapshot. Reconciliation is neither planned nor applied.
func (session MCPServeSession) DoctorDeployment() (output.Doctor, error) {
	result, _, err := diagnoseDeployment(session.project, session.cacheRoot,
		reconcile.Planner{}, false, false)
	return result, err
}

// DoctorRun validates the latest retained evidence layout beneath the fixed
// startup project without reading infrastructure state.
func (session MCPServeSession) DoctorRun() output.Doctor {
	return diagnoseLatestRun(session.project.Target.Root)
}

// DoctorTemplate diagnoses only the verified template bound to the startup
// project. Missing lock/cache evidence produces the same typed skip/failure
// findings as doctor all and never triggers source acquisition.
func (session MCPServeSession) DoctorTemplate() (output.Doctor, error) {
	deployment, _, err := diagnoseDeployment(session.project, session.cacheRoot,
		reconcile.Planner{}, false, false)
	if err != nil {
		return output.Doctor{}, err
	}
	findings := resolvedTemplateFindings(deployment.Findings)
	return output.Doctor{Scope: "template", Summary: summarizeDoctorFindings(findings),
		Findings: findings}, nil
}

// DoctorEnvironment returns the diagnostic snapshot captured before the MCP
// transport opened.
func (session MCPServeSession) DoctorEnvironment() output.Doctor {
	return session.environmentDoctor
}

// PrepareMCPServe resolves one project and validates its normal configuration
// before a protocol transport starts accepting requests.
func PrepareMCPServe(ctx context.Context, request MCPServeRequest,
	options MCPServeOptions,
) (MCPServeSession, error) {
	planOptions := options.Plan
	environmentPath := planOptions.Environment["AINFRA_PROJECT"]
	if request.ProjectPath != "" && environmentPath != "" &&
		request.ProjectPath != environmentPath {
		return MCPServeSession{}, errors.New("--project conflicts with AINFRA_PROJECT")
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: planOptions.WorkingDirectory,
		ProjectPath:      request.ProjectPath,
		EnvironmentPath:  environmentPath,
	})
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("load MCP project: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{}, deployment.Target.Root, planOptions)
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("resolve MCP configuration: %w", err)
	}
	doctorOptions := options.Doctor
	doctorOptions.Environment = cloneEnvironment(doctorOptions.Environment)
	delete(doctorOptions.Environment, "AINFRA_PROJECT")
	environment, err := DoctorEnvironmentContext(ctx, DoctorEnvironmentRequest{
		ProjectPath: deployment.Target.Root, NonInteractive: true,
	}, doctorOptions)
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("diagnose MCP environment: %w", err)
	}
	return MCPServeSession{Project: output.Deployment{
		Name: deployment.Metadata.Name,
		Root: deployment.Target.Root,
	}, project: deployment, runsRoot: settings.Paths.Runs,
		cacheRoot: settings.Paths.Cache, environmentDoctor: environment.Result}, nil
}

func cloneEnvironment(source map[string]string) map[string]string {
	cloned := make(map[string]string, len(source))
	for name, value := range source {
		cloned[name] = value
	}
	return cloned
}
