package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/tofu"
)

// PlanRecord is the immutable reviewed-plan binding persisted in a run.
type PlanRecord struct {
	SchemaVersion int             `json:"schemaVersion"`
	RunID         string          `json:"runId"`
	Intent        string          `json:"intent"`
	Deployment    NamedBinding    `json:"deployment"`
	Template      TemplateBinding `json:"template"`
	Inputs        []InputBinding  `json:"inputs"`
	Engine        EngineBinding   `json:"engine"`
	Plan          PlanBinding     `json:"plan"`
}

type NamedBinding struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}
type TemplateBinding struct {
	Source string `json:"source"`
	Commit string `json:"commit,omitempty"`
	Digest string `json:"digest"`
}
type InputBinding struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}
type EngineBinding struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	ExecutableDigest string `json:"executableDigest"`
}
type PlanBinding struct {
	Path        string `json:"path"`
	Digest      string `json:"digest"`
	SummaryPath string `json:"summaryPath"`
}

// RunRecord is immutable initial lifecycle metadata. Outcomes are separate.
type RunRecord struct {
	SchemaVersion     int      `json:"schemaVersion"`
	RunID             string   `json:"runId"`
	Operation         string   `json:"operation"`
	State             string   `json:"state"`
	CreatedAt         string   `json:"createdAt"`
	PlanRecord        string   `json:"planRecord"`
	OutcomeReferences []string `json:"outcomeReferences,omitempty"`
}

// PublishPlan atomically creates the immutable plan, run, and summary records.
func PublishPlan(prepared Prepared, deploymentName, engineVersion string, createdAt time.Time, summary tofu.Summary) (PlanRecord, error) {
	return PublishPlanIntent(prepared, deploymentName, engineVersion, "apply", createdAt, summary)
}

// PublishPlanIntent atomically publishes a plan with its exact intent.
func PublishPlanIntent(prepared Prepared, deploymentName, engineVersion, intent string, createdAt time.Time, summary tofu.Summary) (PlanRecord, error) {
	planDigest, err := digestFile(prepared.Root + string(os.PathSeparator) + "plan.tfplan")
	if err != nil {
		return PlanRecord{}, fmt.Errorf("bind saved plan: %w", err)
	}
	inputs := make([]InputBinding, len(prepared.NativeInputs))
	for index, binding := range prepared.NativeInputs {
		inputs[index] = InputBinding{Path: binding.Path, Digest: binding.Digest}
	}
	record := PlanRecord{SchemaVersion: 1, RunID: prepared.ID, Intent: intent, Deployment: NamedBinding{Name: deploymentName, Digest: prepared.Deployment.Digest}, Template: TemplateBinding{Source: prepared.TemplateSource, Commit: gitCommit(prepared.TemplateResolved), Digest: prepared.TemplateDigest}, Inputs: inputs, Engine: EngineBinding{Name: "opentofu", Version: engineVersion, ExecutableDigest: prepared.ExecutableDigest}, Plan: PlanBinding{Path: "plan.tfplan", Digest: planDigest, SummaryPath: "plan.json"}}
	runRecord := RunRecord{SchemaVersion: 1, RunID: prepared.ID, Operation: "plan", State: "succeeded", CreatedAt: createdAt.UTC().Format(time.RFC3339), PlanRecord: "plan-record.json"}
	for _, output := range []struct {
		name  string
		value any
	}{{"plan.json", summary}, {"plan-record.json", record}, {"run.json", runRecord}} {
		if err := writeJSON(prepared.Root, output.name, output.value); err != nil {
			return PlanRecord{}, err
		}
	}
	return record, nil
}

func writeJSON(root, name string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')
	staging := "." + name + ".staging"
	file, err := security.CreatePrivateFile(root, staging)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(root + string(os.PathSeparator) + staging)
		}
	}()
	_, writeErr := file.Write(contents)
	err = errors.Join(writeErr, file.Sync(), file.Close())
	if err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	privateRoot, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("open run root: %w", err)
	}
	renameErr := privateRoot.Rename(staging, name)
	closeErr := privateRoot.Close()
	if err := errors.Join(renameErr, closeErr); err != nil {
		return fmt.Errorf("publish %s: %w", name, err)
	}
	cleanup = false
	return nil
}

func gitCommit(resolved string) string {
	if len(resolved) == 40 || len(resolved) == 64 {
		return resolved
	}
	return ""
}
