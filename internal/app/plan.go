package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/projectious-work/ainfra/internal/project"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/template"
	"github.com/projectious-work/ainfra/internal/tofu"
)

// PlanOptions supplies already trusted policy outputs for saved planning.
type PlanOptions struct {
	Prepare       runstate.Options
	Template      template.Contract
	Adapter       tofu.Adapter
	EngineVersion string
	CreatedAt     time.Time
}

// CreateApplyPlan prepares immutable inputs, invokes OpenTofu, sanitizes its
// structural summary, and publishes the reviewed plan record. It never applies.
func CreateApplyPlan(ctx context.Context, options PlanOptions) (runstate.PlanRecord, error) {
	prepared, err := runstate.Prepare(options.Prepare)
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	engineDirectory := filepath.Join("workspace", filepath.FromSlash(options.Template.Tofu.Directory))
	backend, variables, err := snapshotArguments(prepared, options.Prepare.Deployment, engineDirectory)
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	if _, err := options.Adapter.Init(ctx, prepared.Root, engineDirectory, backend); err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("initialize OpenTofu: %w", err)
	}
	planPath, err := filepath.Rel(filepath.Join(prepared.Root, engineDirectory), filepath.Join(prepared.Root, "plan.tfplan"))
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	if _, err := options.Adapter.Plan(ctx, prepared.Root, engineDirectory, planPath, variables, false); err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("create saved OpenTofu plan: %w", err)
	}
	summary, _, err := options.Adapter.ShowSummary(ctx, prepared.Root, engineDirectory, planPath)
	if err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("summarize saved OpenTofu plan: %w", err)
	}
	return runstate.PublishPlan(prepared, options.Prepare.Deployment.Metadata.Name, options.EngineVersion, options.CreatedAt, summary)
}

func snapshotArguments(prepared runstate.Prepared, deployment project.Deployment, engineDirectory string) ([]string, []string, error) {
	lookup := make(map[string]string, len(prepared.NativeInputs))
	for _, binding := range prepared.NativeInputs {
		lookup[binding.Path] = binding.SnapshotPath
	}
	convert := func(values []string) ([]string, error) {
		arguments := make([]string, len(values))
		for index, value := range values {
			snapshot, ok := lookup[filepath.ToSlash(value)]
			if !ok {
				return nil, fmt.Errorf("native input snapshot missing for %q", value)
			}
			relative, err := filepath.Rel(filepath.Join(prepared.Root, engineDirectory), filepath.Join(prepared.Root, filepath.FromSlash(snapshot)))
			if err != nil {
				return nil, err
			}
			arguments[index] = relative
		}
		return arguments, nil
	}
	backend, err := convert(deployment.Inputs.TofuBackendConfigFiles)
	if err != nil {
		return nil, nil, err
	}
	variables, err := convert(deployment.Inputs.TofuVariableFiles)
	return backend, variables, err
}
