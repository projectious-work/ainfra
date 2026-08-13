// Package tofu constructs and executes controlled OpenTofu invocations.
package tofu

import (
	"context"
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

// Init initializes the native module with declared backend files in order.
func (adapter Adapter) Init(ctx context.Context, root, directory string, backendFiles []string) (Outcome, error) {
	arguments := []string{"init", "-input=false", "-no-color"}
	for _, path := range backendFiles {
		if err := validateInputPath(root, path); err != nil {
			return Outcome{}, fmt.Errorf("backend configuration: %w", err)
		}
		arguments = append(arguments, "-backend-config="+filepath.Clean(path))
	}
	return adapter.execute(ctx, root, directory, arguments)
}

// Plan creates an apply or destroy saved plan with variable files in order.
func (adapter Adapter) Plan(ctx context.Context, root, directory, output string, variableFiles []string, destroy bool) (Outcome, error) {
	if err := validateOutputPath(root, output); err != nil {
		return Outcome{}, fmt.Errorf("saved plan output: %w", err)
	}
	arguments := []string{"plan", "-input=false", "-no-color", "-out=" + filepath.Clean(output)}
	if destroy {
		arguments = append(arguments, "-destroy")
	}
	for _, path := range variableFiles {
		if err := validateInputPath(root, path); err != nil {
			return Outcome{}, fmt.Errorf("variable file: %w", err)
		}
		arguments = append(arguments, "-var-file="+filepath.Clean(path))
	}
	return adapter.execute(ctx, root, directory, arguments)
}

func (adapter Adapter) execute(ctx context.Context, root, directory string, arguments []string) (Outcome, error) {
	run := adapter.Run
	if run == nil {
		runner := childexec.Runner{}
		run = runner.Run
	}
	invocation := Invocation{Executable: adapter.Executable.Path(), Arguments: append([]string(nil), arguments...), WorkingRoot: root, WorkingDir: directory}
	result, err := run(ctx, childexec.Request{Executable: adapter.Executable, Args: arguments, WorkingRoot: root, WorkingDir: directory, Environment: adapter.Environment})
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

func validateInputPath(root, path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("contained relative path required")
	}
	resolved, err := security.ResolveContained(root, path)
	if err != nil {
		return err
	}
	return security.RequireRegular(resolved)
}

func validateOutputPath(root, path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("contained relative path required")
	}
	parent, err := security.ResolveContained(root, filepath.Dir(path))
	if err != nil {
		return err
	}
	info, err := filepath.Abs(parent)
	if err != nil || info == "" {
		return errors.New("invalid output parent")
	}
	return nil
}
