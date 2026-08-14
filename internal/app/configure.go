package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/projectious-work/ainfra/internal/ansible"
	"github.com/projectious-work/ainfra/internal/inventory"
	"github.com/projectious-work/ainfra/internal/output"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
)

// ConfigureRequest selects a retained applied run and optional verification.
type ConfigureRequest struct {
	ArtifactRequest
	Check bool
}

// ConfigureFailure retains attributed terminal evidence for a failed runner.
type ConfigureFailure struct {
	Result output.Execution
	Cause  error
}

func (failure *ConfigureFailure) Error() string { return failure.Cause.Error() }
func (failure *ConfigureFailure) Unwrap() error { return failure.Cause }

// Configure runs the declared native playbook against generated inventory.
func Configure(ctx context.Context, request ConfigureRequest, options PlanHostOptions) (output.Execution, error) {
	_, contract, err := resolveArtifactApplicability(request.ArtifactRequest, options)
	if err != nil {
		return output.Execution{}, err
	}
	if contract.Ansible == nil || contract.Inventory == "none" {
		return output.Execution{}, errors.New("ansible configuration is not applicable to this infrastructure-only template")
	}
	resolved, unlock, err := resolveLockedArtifactContext(ctx, request.ArtifactRequest, options)
	if err != nil {
		return output.Execution{}, err
	}
	defer func() { _ = unlock() }()
	outputContents, err := readPrivateArtifact(resolved.reviewed.Root, "output.json", 16<<20)
	if err != nil {
		return output.Execution{}, fmt.Errorf("read validated output: %w", err)
	}
	standard, err := inventory.Parse(outputContents)
	if err != nil {
		return output.Execution{}, fmt.Errorf("revalidate retained output: %w", err)
	}
	if _, err := readPrivateArtifact(resolved.reviewed.Root, "inventory.yaml", 16<<20); err != nil {
		return output.Execution{}, fmt.Errorf("read generated inventory: %w", err)
	}
	ssh := false
	expectedHosts := make([]string, 0, len(standard.Hosts))
	for name, host := range standard.Hosts {
		expectedHosts = append(expectedHosts, name)
		ssh = ssh || host.Connection.Type == "ssh"
	}
	sort.Strings(expectedHosts)
	if ssh && resolved.deployment.SSH.KnownHosts == "" {
		return output.Execution{}, errors.New("SSH inventory requires an independently populated known_hosts file")
	}
	runnerPath := resolved.settings.Executables.AnsibleRunner
	if runnerPath == "" {
		runnerPath, err = exec.LookPath("ansible-runner")
	}
	if err != nil {
		return output.Execution{}, fmt.Errorf("discover Ansible Runner executable: %w", err)
	}
	executable, err := security.ResolveExecutable(runnerPath)
	if err != nil {
		return output.Execution{}, fmt.Errorf("validate Ansible Runner executable: %w", err)
	}
	explicit := map[string]string{"ANSIBLE_HOST_KEY_CHECKING": "True"}
	if ssh {
		knownHosts := filepath.Join(resolved.reviewed.Root, "inputs", filepath.FromSlash(resolved.deployment.SSH.KnownHosts))
		info, statErr := os.Stat(knownHosts)
		if statErr != nil || info.Size() == 0 {
			return output.Execution{}, errors.New("SSH inventory requires a populated bound known_hosts file")
		}
		explicit["ANSIBLE_SSH_ARGS"] = "-o StrictHostKeyChecking=yes -o UserKnownHostsFile=" + strconv.Quote(knownHosts)
	}
	environment, err := security.BuildEnvironment(options.ParentEnvironment,
		[]string{"HOME", "PATH", "LANG", "LC_ALL", "SSL_CERT_DIR", "SSL_CERT_FILE"}, explicit)
	if err != nil {
		return output.Execution{}, fmt.Errorf("build Ansible environment: %w", err)
	}
	adapter := ansible.Adapter{Executable: executable, Environment: environment}
	version, err := adapter.Version(ctx, resolved.reviewed.Root)
	if err != nil {
		return output.Execution{}, fmt.Errorf("read Ansible Runner version: %w", err)
	}
	operation := "configure"
	if request.Check {
		operation = "configure-check"
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	privateData := filepath.Join("ansible-runner", operation)
	artifactDir := filepath.Join(privateData, "artifacts")
	for _, relative := range []string{privateData, artifactDir} {
		if err := os.MkdirAll(filepath.Join(resolved.reviewed.Root, relative), 0o700); err != nil {
			return output.Execution{}, fmt.Errorf("create Ansible Runner directory: %w", err)
		}
	}
	if err := runstate.BeginOperation(resolved.reviewed, operation, now()); err != nil {
		return output.Execution{}, err
	}
	projectDir := filepath.Join("workspace", filepath.FromSlash(resolved.contract.Ansible.Directory))
	variableFiles := make([]string, len(resolved.deployment.Inputs.AnsibleVariableFiles))
	for index, path := range resolved.deployment.Inputs.AnsibleVariableFiles {
		variableFiles[index] = filepath.Join("inputs", filepath.FromSlash(path))
	}
	outcome, runErr := adapter.Configure(ctx, resolved.reviewed.Root, privateData,
		projectDir, "inventory.yaml", resolved.contract.Ansible.Playbook, artifactDir,
		variableFiles, request.Check)
	if runErr == nil && request.Check {
		runErr = ansible.VerifyConverged(outcome.Stats, expectedHosts)
	}
	state, status, executionOutcome := "succeeded", "succeeded", "succeeded"
	if runErr != nil {
		state, status, executionOutcome = "failed", "failed", "failed"
	}
	if errors.Is(runErr, context.Canceled) || outcome.Result.Cancelled || outcome.Result.Signalled {
		state, status, executionOutcome = "cancelled", "interrupted", "interrupted"
	}
	exitCode := outcome.Result.ExitCode
	var reportedExit *int
	if outcome.Result.Started && !outcome.Result.Signalled {
		reportedExit = &exitCode
	}
	result := output.Execution{Deployment: output.Deployment{Name: resolved.deployment.Metadata.Name,
		Root: resolved.deployment.Target.Root}, RunID: request.RunID, Operation: operation,
		ExecutionOutcome: executionOutcome,
		EngineReports: []output.EngineReport{{Engine: "ansible-runner", Status: status,
			ExitCode: reportedExit, Protocol: output.Protocol{Name: "ansible-runner-events", Version: version}}},
		Evidence: []output.Evidence{{Kind: "runner-artifacts", Engine: "ansible-runner",
			Path: filepath.ToSlash(artifactDir), Sensitive: true}},
		Recovery: output.Recovery{AutomaticRetryAllowed: false,
			InspectionRequired: executionOutcome == "interrupted", NextCommands: []string{}}}
	if eventErr := runstate.AppendOperationEvent(resolved.reviewed, operation, state, now(), reportedExit); eventErr != nil {
		result.ExecutionOutcome = "interrupted"
		result.EngineReports[0].Status = "interrupted"
		result.Recovery.InspectionRequired = true
		return result, &ConfigureFailure{Result: result, Cause: fmt.Errorf("record Ansible outcome: %w", eventErr)}
	}
	if runErr != nil {
		return result, &ConfigureFailure{Result: result, Cause: runErr}
	}
	return result, nil
}
