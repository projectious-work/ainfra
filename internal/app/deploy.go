package app

import (
	"context"
	"fmt"

	"github.com/projectious-work/ainfra/internal/output"
)

// DeployRequest selects one exact reviewed plan for the complete applicable
// deployment pipeline.
type DeployRequest struct{ ApplyRequest }

// DeployFailure retains completed-stage evidence when a later stage fails.
type DeployFailure struct {
	Result output.Execution
	Cause  error
}

func (failure *DeployFailure) Error() string { return failure.Cause.Error() }
func (failure *DeployFailure) Unwrap() error { return failure.Cause }

// Deploy applies, collects output, generates inventory, configures, and then
// independently verifies configuration convergence.
func Deploy(ctx context.Context, request DeployRequest, options PlanHostOptions) (output.Execution, error) {
	applied, err := Apply(ctx, request.ApplyRequest, options)
	if err != nil {
		return applied, err
	}
	result := output.Execution{Deployment: applied.Deployment, RunID: applied.RunID,
		Operation: "deploy", ExecutionOutcome: "succeeded",
		EngineReports: append([]output.EngineReport(nil), applied.EngineReports...),
		Evidence:      append([]output.Evidence(nil), applied.Evidence...),
		Recovery:      output.Recovery{AutomaticRetryAllowed: false, NextCommands: []string{}}}
	artifactRequest := ArtifactRequest{Target: request.Target, ProjectPath: request.ProjectPath,
		ConfigPath: request.ConfigPath, RunID: request.PlanID}
	_, contract, err := resolveArtifactApplicability(artifactRequest, options)
	if err != nil {
		return failedDeploy(result, err)
	}
	if contract.Inventory == "none" {
		result.Stages = notApplicableDeployStages("template declares inventory mode none")
		return result, nil
	}
	standard, err := CollectOutput(ctx, artifactRequest, options)
	if err != nil {
		result.Stages = append(result.Stages, failedDeployStage("output"))
		return failedDeploy(result, fmt.Errorf("collect output: %w", err))
	}
	result.Evidence = append(result.Evidence, *standard.Artifact)
	result.Stages = append(result.Stages, output.StageReport{Operation: "output", Applicability: "applicable", Status: "succeeded"})
	generated, err := GenerateInventory(ctx, artifactRequest, options)
	if err != nil {
		result.Stages = append(result.Stages, failedDeployStage("inventory"))
		return failedDeploy(result, fmt.Errorf("generate inventory: %w", err))
	}
	result.Evidence = append(result.Evidence, *generated.Artifact)
	result.Stages = append(result.Stages, output.StageReport{Operation: "inventory", Applicability: "applicable", Status: "succeeded"})
	configured, err := Configure(ctx, ConfigureRequest{ArtifactRequest: artifactRequest}, options)
	if err != nil {
		result.Stages = append(result.Stages, failedDeployStage("configure"))
		return failedDeploy(result, fmt.Errorf("configure: %w", err))
	}
	result.EngineReports = append(result.EngineReports, configured.EngineReports...)
	result.Evidence = append(result.Evidence, configured.Evidence...)
	result.Stages = append(result.Stages, output.StageReport{Operation: "configure", Applicability: "applicable", Status: "succeeded"})
	verified, err := Configure(ctx, ConfigureRequest{ArtifactRequest: artifactRequest, Check: true}, options)
	if err != nil {
		result.Stages = append(result.Stages, failedDeployStage("configure-check"))
		return failedDeploy(result, fmt.Errorf("verify configuration: %w", err))
	}
	result.EngineReports = append(result.EngineReports, verified.EngineReports...)
	result.Evidence = append(result.Evidence, verified.Evidence...)
	result.Stages = append(result.Stages, output.StageReport{Operation: "configure-check", Applicability: "applicable", Status: "succeeded"})
	return result, nil
}

func notApplicableDeployStages(reason string) []output.StageReport {
	stages := make([]output.StageReport, 0, 4)
	for _, operation := range []string{"output", "inventory", "configure", "configure-check"} {
		stages = append(stages, output.StageReport{Operation: operation,
			Applicability: "not-applicable", Status: "not-applicable", Reason: reason})
	}
	return stages
}

func failedDeployStage(operation string) output.StageReport {
	return output.StageReport{Operation: operation, Applicability: "applicable", Status: "failed"}
}

func failedDeploy(result output.Execution, err error) (output.Execution, error) {
	result.ExecutionOutcome = "failed"
	result.Recovery.InspectionRequired = true
	return result, &DeployFailure{Result: result, Cause: err}
}
