// Package reconcile plans and applies narrowly registered local doctor fixes.
package reconcile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectious-work/ainfra/internal/security"
)

// ActionKind identifies one closed, locally owned filesystem repair.
type ActionKind string

const (
	// ActionCreateRuntimeDirectory creates the deployment runtime directory.
	ActionCreateRuntimeDirectory ActionKind = "create_runtime_directory"
	// ActionRestrictRuntimeDirectory removes group and other permissions.
	ActionRestrictRuntimeDirectory ActionKind = "restrict_runtime_directory"
)

// Action is one reviewable and deterministic proposed change.
type Action struct {
	CheckID            string
	Path               string
	Kind               ActionKind
	Mode               os.FileMode
	RollbackLimitation string
}

// Plan is the complete ordered set of changes reviewed by the caller.
type Plan struct {
	DeploymentRoot string
	Actions        []Action
}

// Result records the outcome of one attempted reconciliation action.
type Result struct {
	Action Action
	Status string
	Error  string
}

// Locker serializes deployment-local writes with lifecycle operations.
type Locker interface {
	Lock(deploymentRoot, manifestName string) (func() error, error)
}

// Planner computes repairs without changing the filesystem.
type Planner struct{}

// RuntimeDirectory plans only the ainfra-owned runtime directory repair.
func (Planner) RuntimeDirectory(deploymentRoot string) (Plan, error) {
	root, err := filepath.Abs(deploymentRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve deployment root: %w", err)
	}
	runtimePath := filepath.Join(root, ".ainfra")
	plan := Plan{DeploymentRoot: root, Actions: []Action{}}
	information, err := os.Lstat(runtimePath)
	if errors.Is(err, os.ErrNotExist) {
		plan.Actions = append(plan.Actions, Action{
			CheckID: "deployment.runtime-permissions", Path: runtimePath,
			Kind: ActionCreateRuntimeDirectory, Mode: 0o700,
			RollbackLimitation: "The empty directory can be removed manually.",
		})
		return plan, nil
	}
	if err != nil {
		return Plan{}, fmt.Errorf("inspect runtime directory: %w", err)
	}
	if information.Mode()&os.ModeSymlink != 0 || !information.IsDir() {
		return Plan{}, &security.Refusal{
			Policy: "reconcile runtime directory", Path: runtimePath,
			Reason: "path is not a non-symlink directory",
		}
	}
	if information.Mode().Perm() != 0o700 {
		plan.Actions = append(plan.Actions, Action{
			CheckID: "deployment.runtime-permissions", Path: runtimePath,
			Kind: ActionRestrictRuntimeDirectory, Mode: 0o700,
			RollbackLimitation: "Previous permissions are not restored automatically.",
		})
	}
	return plan, nil
}

// Apply rechecks the complete plan under the deployment lock before writing.
func (planner Planner) Apply(plan Plan, locker Locker) ([]Result, error) {
	if locker == nil {
		return nil, errors.New("deployment locker is required")
	}
	manifestPath, err := security.ResolveContained(plan.DeploymentRoot, "ainfra.yaml")
	if err != nil {
		return nil, err
	}
	if err := security.RequireRegular(manifestPath); err != nil {
		return nil, err
	}
	unlock, err := locker.Lock(plan.DeploymentRoot, filepath.Base(manifestPath))
	if err != nil {
		return nil, fmt.Errorf("acquire deployment operation lock: %w", err)
	}
	defer func() { _ = unlock() }()
	current, err := planner.RuntimeDirectory(plan.DeploymentRoot)
	if err != nil {
		return nil, err
	}
	if !sameActions(plan.Actions, current.Actions) {
		return nil, errors.New("reconciliation preconditions changed; review a new plan")
	}
	results := make([]Result, 0, len(plan.Actions))
	for _, action := range plan.Actions {
		result := Result{Action: action, Status: "applied"}
		if err := applyAction(action); err != nil {
			result.Status, result.Error = "failed", err.Error()
			results = append(results, result)
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func applyAction(action Action) error {
	switch action.Kind {
	case ActionCreateRuntimeDirectory:
		return os.Mkdir(action.Path, action.Mode)
	case ActionRestrictRuntimeDirectory:
		return os.Chmod(action.Path, action.Mode)
	default:
		return fmt.Errorf("unsupported reconciliation action %q", action.Kind)
	}
}

func sameActions(left, right []Action) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
