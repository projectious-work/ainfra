package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	childexec "github.com/projectious-work/ainfra/internal/exec"
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
	return applyForDeployment(ctx, deployment, settings, request.PlanID, options)
}

func applyForDeployment(ctx context.Context, deployment project.Deployment,
	settings config.Settings, planID string, options PlanHostOptions,
) (output.Execution, error) {
	return applyForDeploymentWithHook(ctx, deployment, settings, planID, options, nil)
}

func applyForDeploymentWithHook(ctx context.Context, deployment project.Deployment,
	settings config.Settings, planID string, options PlanHostOptions,
	beforeExecute func() error,
) (output.Execution, error) {
	tofuPath := settings.Executables.Tofu
	var err error
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
	deployment, err = project.Load(project.ResolveOptions{ProjectPath: deployment.Target.Root})
	if err != nil {
		return output.Execution{}, fmt.Errorf("reload deployment under operation lock: %w", err)
	}
	lock, err := lockfile.Read(filepath.Join(deployment.Target.Root, lockfile.Filename))
	if err != nil {
		return output.Execution{}, fmt.Errorf("read template lock: %w", err)
	}
	reviewed, err := runstate.LoadReviewed(runstate.ReviewOptions{ID: planID,
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
	if beforeExecute != nil {
		if err := beforeExecute(); err != nil {
			return output.Execution{}, fmt.Errorf("authorize reviewed apply execution: %w", err)
		}
	}
	return ExecuteReviewedApply(ctx, ApplyExecutionOptions{Reviewed: reviewed,
		Deployment: deployment, Template: contract, Adapter: adapter, Now: options.Now})
}

// ApplyExecutionOptions supplies already reverified apply dependencies.
type ApplyExecutionOptions struct {
	Reviewed   runstate.Reviewed
	Deployment project.Deployment
	Template   template.Contract
	Adapter    tofu.Adapter
	Now        func() time.Time
}

// ExecuteReviewedApply records and executes one already reverified saved plan.
func ExecuteReviewedApply(ctx context.Context, options ApplyExecutionOptions) (output.Execution, error) {
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	reviewed, deployment := options.Reviewed, options.Deployment
	evidenceIO, closeEvidence, err := openTofuEvidence(reviewed.Root)
	if err != nil {
		return output.Execution{}, fmt.Errorf("create private OpenTofu evidence: %w", err)
	}
	if err := runstate.AppendExecutionEvent(reviewed, "started", now(), nil); err != nil {
		_ = closeEvidence()
		return output.Execution{}, fmt.Errorf("record apply start: %w", err)
	}
	directory := filepath.Join("workspace", filepath.FromSlash(options.Template.Tofu.Directory))
	planPath, err := filepath.Rel(filepath.Join(reviewed.Root, directory),
		filepath.Join(reviewed.Root, reviewed.Record.Plan.Path))
	if err != nil {
		return output.Execution{}, err
	}
	adapter := options.Adapter
	adapter.EvidenceIO = evidenceIO
	outcome, applyErr := adapter.Apply(ctx, reviewed.Root, directory, planPath)
	applyErr = errors.Join(applyErr, closeEvidence())
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
		Evidence: []output.Evidence{{Kind: "run-events", Path: "events.jsonl", Sensitive: false},
			{Kind: "raw-engine-stream", Engine: "opentofu", Path: "opentofu.stdout", Sensitive: true},
			{Kind: "raw-engine-stream", Engine: "opentofu", Path: "opentofu.stderr", Sensitive: true}},
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

func openTofuEvidence(root string) (childexec.IOPolicy, func() error, error) {
	stdout, err := security.CreatePrivateFile(root, "opentofu.stdout")
	if err != nil {
		return childexec.IOPolicy{}, nil, err
	}
	stderr, err := security.CreatePrivateFile(root, "opentofu.stderr")
	if err != nil {
		_ = stdout.Close()
		_ = os.Remove(filepath.Join(root, "opentofu.stdout"))
		return childexec.IOPolicy{}, nil, err
	}
	closeEvidence := func() error {
		return errors.Join(stdout.Sync(), stderr.Sync(), stdout.Close(), stderr.Close())
	}
	return childexec.IOPolicy{
		RawStdout: &rawEvidenceWriter{file: stdout, remaining: 64 << 20},
		RawStderr: &rawEvidenceWriter{file: stderr, remaining: 64 << 20},
	}, closeEvidence, nil
}

type rawEvidenceWriter struct {
	file      *os.File
	remaining int64
}

func (writer *rawEvidenceWriter) Write(contents []byte) (int, error) {
	if int64(len(contents)) > writer.remaining {
		return 0, errors.New("raw OpenTofu evidence exceeds size limit")
	}
	written, err := writer.file.Write(contents)
	writer.remaining -= int64(written)
	return written, err
}
