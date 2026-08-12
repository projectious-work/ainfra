package doctor

import (
	"context"
	"fmt"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

// DeploymentInput contains only already-validated deployment contract facts.
type DeploymentInput struct {
	Name         string
	Root         string
	ManifestPath string
	NativeFiles  int
}

// DeploymentRegistry returns checks for a loaded local deployment contract.
func DeploymentRegistry(input DeploymentInput) Registry {
	definitions := []Definition{
		deploymentFactCheck(
			"deployment.manifest", "AINFRA-E2301", input.ManifestPath,
			"deployment manifest is structurally valid",
		),
		deploymentFactCheck(
			"deployment.native-inputs", "AINFRA-E2302", input.Root,
			fmt.Sprintf("%d native input pointers are contained regular files", input.NativeFiles),
		),
		deploymentFactCheck(
			"deployment.identity", "AINFRA-E2303", input.Root,
			fmt.Sprintf("deployment %q has one canonical root", input.Name),
		),
	}
	registry, err := NewRegistry(definitions...)
	if err != nil {
		panic(err)
	}
	return registry
}

func deploymentFactCheck(id, code, path, message string) Definition {
	return Definition{
		ID: id, Scope: ScopeDeployment,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			return diagnostic.Diagnostic{
				Code: code, Severity: diagnostic.SeverityInfo, Status: "pass",
				Component: "deployment", Path: path, Message: message,
			}
		},
	}
}
