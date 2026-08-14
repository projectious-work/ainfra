package app

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/template"
	"github.com/projectious-work/ainfra/internal/tofu"
)

// ApplyRequest selects one exact reviewed apply plan.
type ApplyRequest struct {
	Target, ProjectPath, ConfigPath, PlanID string
}

// ApplyFailure retains a terminal result for a failed or interrupted child.
type ApplyFailure struct {
	Result output.Execution
	Cause  error
}

func (failure *ApplyFailure) Error() string { return failure.Cause.Error() }
func (failure *ApplyFailure) Unwrap() error { return failure.Cause }

// Apply reverifies and executes only the exact reviewed saved plan.
func Apply(ctx context.Context, request ApplyRequest, options PlanHostOptions) (output.Execution, error) {
	if request.PlanID == "" {
		return output.Execution{}, errors.New("--plan requires an exact reviewed plan ID")
	}
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return output.Execution{}, errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return output.Execution{}, fmt.Errorf("load deployment contract: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{ConfigPath: request.ConfigPath},
		deployment.Target.Root, options)
	if err != nil {
		return output.Execution{}, fmt.Errorf("resolve apply configuration: %w", err)
	}
	tofuPath := settings.Executables.Tofu
	if tofuPath == "" {
		tofuPath, err = exec.LookPath("tofu")
		if err != nil {
			return output.Execution{}, fmt.Errorf("discover OpenTofu executable: %w", err)
		}
	}
	executable, err := security.ResolveExecutable(tofuPath)
	if err != nil {
		return output.Execution{}, fmt.Errorf("validate OpenTofu executable: %w", err)
	}
	environment, err := security.BuildEnvironment(options.ParentEnvironment,
		[]string{"HOME", "PATH", "SSL_CERT_DIR", "SSL_CERT_FILE"},
		map[string]string{"TF_IN_AUTOMATION": "1"})
	if err != nil {
		return output.Execution{}, fmt.Errorf("build OpenTofu environment: %w", err)
	}
	adapter := tofu.Adapter{Executable: executable, Environment: environment}
	version, err := adapter.Version(ctx, deployment.Target.Root)
	if err != nil {
		return output.Execution{}, fmt.Errorf("read OpenTofu version: %w", err)
	}
	unlock, err := (reconcile.FileLocker{}).Lock(deployment.Target.Root, project.ManifestName)
	if err != nil {
		return output.Execution{}, fmt.Errorf("acquire deployment operation lock: %w", err)
	}
	defer func() { _ = unlock() }()
	deployment, err = project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return output.Execution{}, fmt.Errorf("reload deployment under operation lock: %w", err)
	}
	lock, err := lockfile.Read(filepath.Join(deployment.Target.Root, lockfile.Filename))
	if err != nil {
		return output.Execution{}, fmt.Errorf("read template lock: %w", err)
	}
	reviewed, err := runstate.LoadReviewed(runstate.ReviewOptions{ID: request.PlanID,
		RunsRoot: settings.Paths.Runs, CacheRoot: settings.Paths.Cache,
		Deployment: deployment, Lock: lock, Executable: executable, EngineVersion: version})
	if err != nil {
		return output.Execution{}, fmt.Errorf("reverify reviewed plan: %w", err)
	}
	cachePath := filepath.Join(settings.Paths.Cache, "templates", "sha256",
		reviewed.Record.Template.Digest[len("sha256:"):])
	contract, err := template.LoadMaterialized(cachePath)
	if err != nil {
		return output.Execution{}, fmt.Errorf("load reviewed template contract: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	if err := runstate.AppendExecutionEvent(reviewed, "started", now(), nil); err != nil {
		return output.Execution{}, fmt.Errorf("record apply start: %w", err)
	}
	directory := filepath.Join("workspace", filepath.FromSlash(contract.Tofu.Directory))
	planPath, err := filepath.Rel(filepath.Join(reviewed.Root, directory),
		filepath.Join(reviewed.Root, reviewed.Record.Plan.Path))
	if err != nil {
		return output.Execution{}, err
	}
	outcome, applyErr := adapter.Apply(ctx, reviewed.Root, directory, planPath)
	state, executionOutcome, status := "succeeded", "succeeded", "succeeded"
	if applyErr != nil {
		state, executionOutcome, status = "failed", "failed", "failed"
		if errors.Is(applyErr, context.Canceled) || outcome.Result.Cancelled || outcome.Result.Signalled {
			state, executionOutcome, status = "cancelled", "interrupted", "interrupted"
		}
	}
	exitCode := outcome.Result.ExitCode
	var reportedExit *int
	if outcome.Result.Started && !outcome.Result.Signalled {
		reportedExit = &exitCode
	}
	result := output.Execution{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, RunID: reviewed.Record.RunID, Operation: "apply",
		ExecutionOutcome: executionOutcome,
		EngineReports: []output.EngineReport{{Engine: "opentofu", Status: status,
			ExitCode: reportedExit, Protocol: output.Protocol{Name: "opentofu-json-ui", Version: "1.2"}}},
		Evidence: []output.Evidence{{Kind: "run-events", Path: "events.jsonl", Sensitive: false}},
		Recovery: output.Recovery{AutomaticRetryAllowed: false,
			InspectionRequired: executionOutcome == "interrupted", NextCommands: []string{}}}
	if eventErr := runstate.AppendExecutionEvent(reviewed, state, now(), reportedExit); eventErr != nil {
		result.ExecutionOutcome = "interrupted"
		result.EngineReports[0].Status = "interrupted"
		result.Recovery.InspectionRequired = true
		result.Recovery.NextCommands = []string{"ainfra doctor run " + deployment.Metadata.Name}
		return result, &ApplyFailure{Result: result,
			Cause: fmt.Errorf("record apply outcome after execution: %w", eventErr)}
	}
	if executionOutcome == "interrupted" {
		if eventErr := runstate.AppendExecutionEvent(reviewed, "inspection-required", now(), reportedExit); eventErr != nil {
			result.Recovery.NextCommands = []string{"ainfra doctor run " + deployment.Metadata.Name}
			return result, &ApplyFailure{Result: result,
				Cause: fmt.Errorf("record apply inspection requirement: %w", eventErr)}
		}
	}
	if applyErr != nil {
		if executionOutcome == "interrupted" {
			result.Recovery.NextCommands = []string{"ainfra doctor run " + deployment.Metadata.Name}
		}
		return result, &ApplyFailure{Result: result, Cause: applyErr}
	}
	return result, nil
}
