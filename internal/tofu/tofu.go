// Package tofu constructs and executes controlled OpenTofu invocations.
package tofu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/security"
)

// Run executes one already-policy-bound OpenTofu request.
type Run func(context.Context, childexec.Request) (childexec.Result, error)

// Adapter binds OpenTofu to an immutable executable and explicit environment.
type Adapter struct {
	Executable  security.Executable
	Environment security.Environment
	Run         Run
}

// Invocation is a display-safe record of an OpenTofu boundary call.
type Invocation struct {
	Executable  string
	Arguments   []string
	WorkingRoot string
	WorkingDir  string
}

// Outcome records the invocation and observed process result.
type Outcome struct {
	Invocation Invocation
	Result     childexec.Result
}

// Summary is a secret-free structural count derived from OpenTofu plan JSON.
type Summary struct {
	Create  int `json:"create"`
	Update  int `json:"update"`
	Delete  int `json:"delete"`
	Replace int `json:"replace"`
	Read    int `json:"read"`
	NoOp    int `json:"noOp"`
}

// Init initializes the native module with declared backend files in order.
func (adapter Adapter) Init(ctx context.Context, root, directory string, backendFiles []string) (Outcome, error) {
	arguments := []string{"init", "-input=false", "-no-color"}
	for _, path := range backendFiles {
		if err := validateInputPath(root, directory, path); err != nil {
			return Outcome{}, fmt.Errorf("backend configuration: %w", err)
		}
		arguments = append(arguments, "-backend-config="+filepath.Clean(path))
	}
	return adapter.execute(ctx, root, directory, arguments)
}

// Plan creates an apply or destroy saved plan with variable files in order.
func (adapter Adapter) Plan(ctx context.Context, root, directory, output string, variableFiles []string, destroy bool) (Outcome, error) {
	if err := validateOutputPath(root, directory, output); err != nil {
		return Outcome{}, fmt.Errorf("saved plan output: %w", err)
	}
	arguments := []string{"plan", "-input=false", "-no-color", "-out=" + filepath.Clean(output)}
	if destroy {
		arguments = append(arguments, "-destroy")
	}
	for _, path := range variableFiles {
		if err := validateInputPath(root, directory, path); err != nil {
			return Outcome{}, fmt.Errorf("variable file: %w", err)
		}
		arguments = append(arguments, "-var-file="+filepath.Clean(path))
	}
	return adapter.execute(ctx, root, directory, arguments)
}

// ShowSummary derives action counts without retaining values from raw plan JSON.
func (adapter Adapter) ShowSummary(ctx context.Context, root, directory, planPath string) (Summary, Outcome, error) {
	if err := validateInputPath(root, directory, planPath); err != nil {
		return Summary{}, Outcome{}, fmt.Errorf("saved plan: %w", err)
	}
	var output bytes.Buffer
	outcome, err := adapter.executeWithIO(ctx, root, directory, []string{"show", "-json", filepath.Clean(planPath)}, childexec.IOPolicy{Stdout: &boundedWriter{destination: &output, remaining: 16 << 20}})
	if err != nil {
		return Summary{}, outcome, err
	}
	var document struct {
		ResourceChanges []struct {
			Change struct {
				Actions []string `json:"actions"`
			} `json:"change"`
		} `json:"resource_changes"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	if err := decoder.Decode(&document); err != nil {
		return Summary{}, outcome, fmt.Errorf("decode OpenTofu plan summary: %w", err)
	}
	var summary Summary
	for _, resource := range document.ResourceChanges {
		actions := resource.Change.Actions
		switch {
		case len(actions) == 1 && actions[0] == "create":
			summary.Create++
		case len(actions) == 1 && actions[0] == "update":
			summary.Update++
		case len(actions) == 1 && actions[0] == "delete":
			summary.Delete++
		case len(actions) == 1 && actions[0] == "read":
			summary.Read++
		case len(actions) == 1 && actions[0] == "no-op":
			summary.NoOp++
		case len(actions) == 2 && ((actions[0] == "delete" && actions[1] == "create") || (actions[0] == "create" && actions[1] == "delete")):
			summary.Replace++
		default:
			return Summary{}, outcome, errors.New("OpenTofu plan contains unsupported actions")
		}
	}
	return summary, outcome, nil
}

type boundedWriter struct {
	destination *bytes.Buffer
	remaining   int
}

func (writer *boundedWriter) Write(contents []byte) (int, error) {
	if len(contents) > writer.remaining {
		return 0, errors.New("OpenTofu plan summary exceeds size limit")
	}
	written, err := writer.destination.Write(contents)
	writer.remaining -= written
	return written, err
}

func (adapter Adapter) execute(ctx context.Context, root, directory string, arguments []string) (Outcome, error) {
	return adapter.executeWithIO(ctx, root, directory, arguments, childexec.IOPolicy{})
}

func (adapter Adapter) executeWithIO(ctx context.Context, root, directory string, arguments []string, policy childexec.IOPolicy) (Outcome, error) {
	run := adapter.Run
	if run == nil {
		runner := childexec.Runner{}
		run = runner.Run
	}
	invocation := Invocation{Executable: adapter.Executable.Path(), Arguments: append([]string(nil), arguments...), WorkingRoot: root, WorkingDir: directory}
	result, err := run(ctx, childexec.Request{Executable: adapter.Executable, Args: arguments, WorkingRoot: root, WorkingDir: directory, Environment: adapter.Environment, IO: policy})
	outcome := Outcome{Invocation: invocation, Result: result}
	if err != nil {
		return outcome, err
	}
	if result.Cancelled {
		return outcome, context.Canceled
	}
	if result.ExitCode != 0 {
		return outcome, fmt.Errorf("opentofu exited with status %d", result.ExitCode)
	}
	return outcome, nil
}

func validateInputPath(root, directory, path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("contained relative path required")
	}
	resolved, err := security.ResolveContained(root, filepath.Join(directory, path))
	if err != nil {
		return err
	}
	return security.RequireRegular(resolved)
}

func validateOutputPath(root, directory, path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("contained relative path required")
	}
	parent, err := security.ResolveContained(root, filepath.Join(directory, filepath.Dir(path)))
	if err != nil {
		return err
	}
	info, err := filepath.Abs(parent)
	if err != nil || info == "" {
		return errors.New("invalid output parent")
	}
	return nil
}
