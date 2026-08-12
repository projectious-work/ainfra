package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/projectious-work/ainfra/internal/doctor"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
	"github.com/projectious-work/ainfra/internal/security"
)

// DoctorError classifies a refused doctor request without coupling app to CLI
// exit-code numbers.
type DoctorError struct {
	Kind    string
	Message string
}

// Error returns the display-safe refusal message.
func (failure *DoctorError) Error() string { return failure.Message }

// DoctorDeploymentRequest identifies one local deployment diagnosis.
type DoctorDeploymentRequest struct {
	Target              string
	ConfigPath          string
	ProjectPath         string
	Format              *string
	OutputStyle         *string
	Color               *string
	Reconcile           bool
	ApplyReconciliation bool
}

// DoctorDeployment diagnoses a local deployment without resolving its template
// source or invoking an infrastructure engine.
func DoctorDeployment(
	request DoctorDeploymentRequest,
	options DoctorEnvironmentOptions,
) (DoctorEnvironmentResponse, error) {
	if request.Target != "" && request.ProjectPath != "" {
		return DoctorEnvironmentResponse{}, errors.New(
			"deployment TARGET conflicts with --project",
		)
	}
	if request.Target != "" && options.Environment["AINFRA_PROJECT"] != "" {
		return DoctorEnvironmentResponse{}, errors.New(
			"deployment TARGET conflicts with AINFRA_PROJECT",
		)
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: options.WorkingDirectory,
		ExplicitPath:     request.Target, ProjectPath: request.ProjectPath,
		EnvironmentPath: options.Environment["AINFRA_PROJECT"],
	})
	if err != nil {
		kind := "input"
		var refusal *security.Refusal
		if errors.As(err, &refusal) {
			kind = "security"
		}
		return DoctorEnvironmentResponse{}, &DoctorError{
			Kind: kind, Message: fmt.Sprintf("load deployment contract: %s", err),
		}
	}
	configuration, err := resolveDoctorConfiguration(
		DoctorEnvironmentRequest{
			ConfigPath: request.ConfigPath, ProjectPath: deployment.Target.Root,
			Format: request.Format, OutputStyle: request.OutputStyle, Color: request.Color,
		},
		options, deployment.Target.Root,
	)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	nativeFiles := len(deployment.Inputs.TofuVariableFiles) +
		len(deployment.Inputs.TofuBackendConfigFiles) +
		len(deployment.Inputs.AnsibleVariableFiles)
	if deployment.SSH.KnownHosts != "" {
		nativeFiles++
	}
	planner := reconcile.Planner{}
	if options.ReconcilePlanner != nil {
		planner = *options.ReconcilePlanner
	}
	runtimePlan, err := planner.RuntimeDirectory(deployment.Target.Root)
	if err != nil {
		return DoctorEnvironmentResponse{}, classifyDoctorLoad("runtime directory", err)
	}
	reconciliationStatus := ""
	reconciliationEvidence := ""
	if request.ApplyReconciliation && len(runtimePlan.Actions) > 0 {
		results, applyErr := planner.Apply(
			runtimePlan, reconcile.FileLocker{},
		)
		if applyErr != nil {
			return DoctorEnvironmentResponse{}, fmt.Errorf("apply reconciliation: %w", applyErr)
		}
		failed := false
		for _, result := range results {
			if result.Status != "applied" {
				failed = true
				reconciliationEvidence = "local repair failed; evidence retained"
			}
		}
		reconciliationStatus = "applied"
		if failed {
			reconciliationStatus = "failed"
		}
		runtimePlan, err = planner.RuntimeDirectory(deployment.Target.Root)
		if err != nil {
			return DoctorEnvironmentResponse{}, err
		}
	}
	report := doctor.DeploymentRegistry(doctor.DeploymentInput{
		Name: deployment.Metadata.Name, Root: deployment.Target.Root,
		ManifestPath: deployment.Target.ManifestPath, NativeFiles: nativeFiles,
		RuntimeSafe: len(runtimePlan.Actions) == 0,
	}).Run(context.Background(), doctor.ScopeDeployment, doctor.Input{}, doctor.Capabilities{})
	if reconciliationStatus != "" {
		for index := range report.Findings {
			if report.Findings[index].Check == "deployment.runtime-permissions" {
				status := reconciliationStatus
				if status == "failed" && report.Findings[index].Status != "pass" {
					status = "still_failing"
				}
				report.Findings[index].Reconciliation = status
				if reconciliationEvidence != "" {
					report.Findings[index].Evidence = reconciliationEvidence
				}
			}
		}
	}
	plan := []reconcile.Action{}
	if request.Reconcile {
		plan = runtimePlan.Actions
	}
	return DoctorEnvironmentResponse{
		Result: output.Doctor{
			Scope: "deployment",
			Summary: output.DoctorSummary{
				Pass: report.Summary.Pass, Skip: report.Summary.Skip,
				Warning: report.Summary.Warning, Fail: report.Summary.Fail,
			},
			Findings: report.Findings,
		},
		Format:             configuration.Settings.UI.Format,
		OutputStyle:        configuration.Settings.UI.OutputStyle,
		Color:              configuration.Settings.UI.Color,
		ReconciliationPlan: plan,
	}, nil
}
