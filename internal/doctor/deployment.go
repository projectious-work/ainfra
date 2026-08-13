package doctor

import (
	"context"
	"fmt"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

// DeploymentInput contains only already-validated deployment contract facts.
type DeploymentInput struct {
	Name          string
	Root          string
	ManifestPath  string
	NativeFiles   int
	RuntimeSafe   bool
	TemplateFacts []DeploymentTemplateFact
}

// DeploymentTemplateFact is one precomputed, side-effect-free lock/cache fact.
type DeploymentTemplateFact struct {
	ID         string
	Code       string
	Path       string
	Status     string
	Message    string
	Evidence   string
	NextAction string
}

// DeploymentRegistry returns checks for a loaded local deployment contract.
func DeploymentRegistry(input DeploymentInput) Registry {
	definitions := []Definition{
		deploymentFactCheck(
			"deployment.manifest", "AINFRA-E2301", input.ManifestPath,
			"deployment manifest is structurally valid",
		),
		deploymentRuntimeCheck(input.Root, input.RuntimeSafe),
		deploymentFactCheck(
			"deployment.native-inputs", "AINFRA-E2302", input.Root,
			fmt.Sprintf("%d native input pointers are contained regular files", input.NativeFiles),
		),
		deploymentFactCheck(
			"deployment.identity", "AINFRA-E2303", input.Root,
			fmt.Sprintf("deployment %q has one canonical root", input.Name),
		),
	}
	for _, fact := range input.TemplateFacts {
		definitions = append(definitions, deploymentTemplateCheck(fact))
	}
	registry, err := NewRegistry(definitions...)
	if err != nil {
		panic(err)
	}
	return registry
}

func deploymentTemplateCheck(fact DeploymentTemplateFact) Definition {
	return Definition{
		ID: fact.ID, Scope: ScopeDeployment,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			severity := diagnostic.SeverityInfo
			if fact.Status == "fail" {
				severity = diagnostic.SeverityError
			}
			return diagnostic.Diagnostic{
				Code: fact.Code, Severity: severity, Status: fact.Status,
				Component: "template", Path: fact.Path, Message: fact.Message,
				Evidence: fact.Evidence, NextAction: fact.NextAction,
			}
		},
	}
}

func deploymentRuntimeCheck(path string, safe bool) Definition {
	return Definition{
		ID: "deployment.runtime-permissions", Scope: ScopeDeployment,
		Prerequisites: []string{"deployment.manifest"},
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			finding := diagnostic.Diagnostic{
				Code: "AINFRA-E2305", Component: "deployment", Path: path,
			}
			if safe {
				finding.Severity, finding.Status = diagnostic.SeverityInfo, "pass"
				finding.Message = "ainfra runtime directory has restrictive permissions"
				return finding
			}
			finding.Severity, finding.Status = diagnostic.SeverityWarning, "warning"
			finding.Reconciliation = "available"
			finding.Message = "ainfra runtime directory is absent or has unsafe permissions"
			finding.NextAction = "Rerun doctor with --reconcile and confirm the local repair plan."
			return finding
		},
	}
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
