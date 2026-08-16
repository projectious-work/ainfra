package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/projectious-work/ainfra/internal/inventory"
	"github.com/projectious-work/ainfra/internal/output"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
)

const retainedArtifactLimit = 16 << 20

// MCPRetainedOutput is sanitized standardized output retained by a bound run.
type MCPRetainedOutput struct {
	Deployment output.Deployment        `json:"deployment"`
	RunID      string                   `json:"runId"`
	Output     inventory.StandardOutput `json:"output"`
}

// MCPRetainedInventory is verified deterministic inventory retained by a
// bound run.
type MCPRetainedInventory struct {
	Deployment output.Deployment `json:"deployment"`
	RunID      string            `json:"runId"`
	MediaType  string            `json:"mediaType"`
	Content    string            `json:"content"`
}

// ReadOutput returns an already-published, sanitized output contract for a
// run bound to the fixed startup project.
func (session MCPServeSession) ReadOutput(runID string) (MCPRetainedOutput, error) {
	root, err := session.boundRunRoot(runID)
	if err != nil {
		return MCPRetainedOutput{}, err
	}
	standard, err := readStandardOutput(root)
	if err != nil {
		return MCPRetainedOutput{}, err
	}
	return MCPRetainedOutput{Deployment: session.Project, RunID: runID, Output: standard}, nil
}

// ReadInventory returns an already-published inventory only when it exactly
// matches the deterministic inventory derived from sanitized retained output.
func (session MCPServeSession) ReadInventory(runID string) (MCPRetainedInventory, error) {
	root, err := session.boundRunRoot(runID)
	if err != nil {
		return MCPRetainedInventory{}, err
	}
	standard, err := readStandardOutput(root)
	if err != nil {
		return MCPRetainedInventory{}, err
	}
	expected, err := inventory.YAML(standard)
	if err != nil {
		return MCPRetainedInventory{}, fmt.Errorf("render retained inventory: %w", err)
	}
	contents, err := readPrivateArtifact(root, "inventory.yaml", retainedArtifactLimit)
	if err != nil {
		return MCPRetainedInventory{}, fmt.Errorf("read retained inventory: %w", err)
	}
	if !bytes.Equal(contents, expected) {
		return MCPRetainedInventory{}, errors.New("retained inventory does not match standardized output")
	}
	return MCPRetainedInventory{Deployment: session.Project, RunID: runID,
		MediaType: "application/yaml", Content: string(contents)}, nil
}

func (session MCPServeSession) boundRunRoot(runID string) (string, error) {
	if runID == "" {
		return "", errors.New("runId is required")
	}
	root, err := security.ResolveContained(session.runsRoot, runID)
	if err != nil {
		return "", errors.New("retained run ID is invalid")
	}
	information, err := os.Lstat(root)
	if err != nil || !information.IsDir() || information.Mode().Perm()&0o077 != 0 {
		return "", errors.New("retained run is not a private directory")
	}
	var runRecord runstate.RunRecord
	if err := readStrictArtifact(root, "run.json", &runRecord); err != nil {
		return "", fmt.Errorf("read retained run record: %w", err)
	}
	var planRecord runstate.PlanRecord
	if err := readStrictArtifact(root, "plan-record.json", &planRecord); err != nil {
		return "", fmt.Errorf("read retained plan record: %w", err)
	}
	if runRecord.SchemaVersion != 1 || planRecord.SchemaVersion != 1 ||
		runRecord.RunID != runID || planRecord.RunID != runID ||
		runRecord.PlanRecord != "plan-record.json" || planRecord.Intent != "apply" {
		return "", errors.New("retained run binding is invalid")
	}
	if planRecord.Deployment.Name != session.Project.Name {
		return "", errors.New("retained run belongs to another deployment")
	}
	return root, nil
}

func readStandardOutput(root string) (inventory.StandardOutput, error) {
	contents, err := readPrivateArtifact(root, "output.json", retainedArtifactLimit)
	if err != nil {
		return inventory.StandardOutput{}, fmt.Errorf("read retained output: %w", err)
	}
	standard, err := inventory.Parse(contents)
	if err != nil {
		return inventory.StandardOutput{}, fmt.Errorf("validate retained output: %w", err)
	}
	return standard, nil
}

func readStrictArtifact(root, name string, target any) error {
	contents, err := readPrivateArtifact(root, name, 1<<20)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("record must contain exactly one JSON value")
	}
	return nil
}
