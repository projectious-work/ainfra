package app

import (
	"errors"
	"fmt"

	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
)

// MCPServeRequest contains server-start policy selected by the operator.
// Protocol-specific types belong to the MCP adapter and must not enter app.
type MCPServeRequest struct {
	ProjectPath  string
	Capabilities []string
}

// MCPServeSession contains the immutable, validated project identity exposed
// to protocol adapters. Request handlers cannot replace this root.
type MCPServeSession struct {
	Project output.Deployment
}

// PrepareMCPServe resolves one project and validates its normal configuration
// before a protocol transport starts accepting requests.
func PrepareMCPServe(request MCPServeRequest, options PlanHostOptions) (MCPServeSession, error) {
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.ProjectPath != "" && environmentPath != "" &&
		request.ProjectPath != environmentPath {
		return MCPServeSession{}, errors.New("--project conflicts with AINFRA_PROJECT")
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: options.WorkingDirectory,
		ProjectPath:      request.ProjectPath,
		EnvironmentPath:  environmentPath,
	})
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("load MCP project: %w", err)
	}
	if _, err := resolvePlanConfiguration(PlanRequest{}, deployment.Target.Root, options); err != nil {
		return MCPServeSession{}, fmt.Errorf("resolve MCP configuration: %w", err)
	}
	return MCPServeSession{Project: output.Deployment{
		Name: deployment.Metadata.Name,
		Root: deployment.Target.Root,
	}}, nil
}
