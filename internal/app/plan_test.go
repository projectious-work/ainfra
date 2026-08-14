package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	childexec "github.com/projectious-work/ainfra/internal/exec"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/project"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
	"github.com/projectious-work/ainfra/internal/template"
	"github.com/projectious-work/ainfra/internal/tofu"
)

func TestCreateApplyPlanPublishesBoundSanitizedRecords(t *testing.T) {
	t.Parallel()
	deploymentRoot := t.TempDir()
	manifest := filepath.Join(deploymentRoot, project.ManifestName)
	writePlanFixture(t, manifest, "deployment")
	writePlanFixture(t, filepath.Join(deploymentRoot, "backend.hcl"), "backend")
	writePlanFixture(t, filepath.Join(deploymentRoot, "variables.tfvars"), "variables")
	templateRoot := t.TempDir()
	writePlanFixture(t, filepath.Join(templateRoot, "ainfra-template.yaml"), "template")
	if err := os.Mkdir(filepath.Join(templateRoot, "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	cacheRoot := t.TempDir()
	materialized, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	executablePath := filepath.Join(t.TempDir(), "tofu")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(executablePath)
	if err != nil {
		t.Fatal(err)
	}
	deployment := project.Deployment{Target: project.Target{Root: deploymentRoot, ManifestPath: manifest}, Metadata: project.Metadata{Name: "development"}, Inputs: project.Inputs{TofuBackendConfigFiles: []string{"backend.hcl"}, TofuVariableFiles: []string{"variables.tfvars"}}}
	document := lockfile.New(lockfile.Template{Source: "local:template", Resolved: "local:template", Version: "1.0.0", Digest: materialized.Digest, ResolvedAt: time.Now()})
	var calls []string
	adapter := tofu.Adapter{Executable: executable, Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
		calls = append(calls, strings.Join(request.Args, " "))
		if request.Args[0] == "plan" {
			for _, argument := range request.Args {
				if strings.HasPrefix(argument, "-out=") {
					writePlanFixture(t, filepath.Join(request.WorkingRoot, request.WorkingDir, strings.TrimPrefix(argument, "-out=")), "saved-plan")
				}
			}
		}
		if request.Args[0] == "show" {
			_, _ = request.IO.Stdout.Write([]byte(`{"resource_changes":[{"change":{"actions":["create"]}}],"secret":"do-not-persist"}`))
		}
		return childexec.Result{Started: true}, nil
	}}
	runsRoot := filepath.Join(t.TempDir(), "runs")
	record, err := app.CreateApplyPlan(context.Background(), app.PlanOptions{Prepare: runstate.Options{ID: "20260813T190000Z-0123456789abcdef", RunsRoot: runsRoot, CacheRoot: cacheRoot, Deployment: deployment, Lock: document, Executable: executable}, Template: template.Contract{Tofu: template.Engine{Directory: "tofu", Version: ">=1.10.0 <2.0.0"}}, Adapter: adapter, EngineVersion: "1.10.0", CreatedAt: time.Date(2026, 8, 13, 19, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || !strings.Contains(calls[0], "-backend-config=../../inputs/backend.hcl") || !strings.Contains(calls[1], "-var-file=../../inputs/variables.tfvars") {
		t.Fatalf("calls: %#v", calls)
	}
	if record.Plan.Digest == "" || record.Intent != "apply" || record.Engine.ExecutableDigest != executable.Digest() {
		t.Fatalf("record: %#v", record)
	}
	for _, name := range []string{"plan.json", "plan-record.json", "run.json"} {
		contents, readErr := os.ReadFile(filepath.Join(runsRoot, record.RunID, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(contents), "do-not-persist") {
			t.Fatalf("%s retained secret", name)
		}
		info, statErr := os.Stat(filepath.Join(runsRoot, record.RunID, name))
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode: %v", name, info.Mode().Perm())
		}
	}
	reviewed, err := runstate.LoadReviewed(runstate.ReviewOptions{ID: record.RunID,
		RunsRoot: runsRoot, CacheRoot: cacheRoot, Deployment: deployment,
		Lock: document, Executable: executable, EngineVersion: "1.10.0"})
	if err != nil {
		t.Fatal(err)
	}
	cancelledAdapter := tofu.Adapter{Executable: executable,
		Run: func(context.Context, childexec.Request) (childexec.Result, error) {
			return childexec.Result{Started: true, Cancelled: true}, nil
		}}
	result, err := app.ExecuteReviewedApply(context.Background(), app.ApplyExecutionOptions{
		Reviewed: reviewed, Deployment: deployment,
		Template: template.Contract{Tofu: template.Engine{Directory: "tofu"}},
		Adapter:  cancelledAdapter, Now: func() time.Time {
			return time.Date(2026, 8, 14, 10, 0, 0, 0, time.UTC)
		},
	})
	var failure *app.ApplyFailure
	if !errors.As(err, &failure) || result.ExecutionOutcome != "interrupted" ||
		!result.Recovery.InspectionRequired {
		t.Fatalf("cancelled result=%+v err=%v", result, err)
	}
	events, err := os.ReadFile(filepath.Join(reviewed.Root, "events.jsonl"))
	if err != nil || !strings.Contains(string(events), `"state":"cancelled"`) ||
		!strings.Contains(string(events), `"state":"inspection-required"`) {
		t.Fatalf("events=%q err=%v", events, err)
	}
}

func TestExecuteReviewedApplyRoutesEvidenceFailureToInspection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "workspace", "tofu"), 0o700); err != nil {
		t.Fatal(err)
	}
	writePlanFixture(t, filepath.Join(root, "plan.tfplan"), "saved-plan")
	executablePath := filepath.Join(t.TempDir(), "tofu")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(executablePath)
	if err != nil {
		t.Fatal(err)
	}
	adapter := tofu.Adapter{Executable: executable,
		Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
			if err := os.Remove(filepath.Join(request.WorkingRoot, "events.jsonl")); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(request.WorkingRoot, "events.jsonl"), 0o700); err != nil {
				t.Fatal(err)
			}
			return childexec.Result{Started: true}, nil
		}}
	result, err := app.ExecuteReviewedApply(context.Background(), app.ApplyExecutionOptions{
		Reviewed: runstate.Reviewed{Root: root,
			Record: runstate.PlanRecord{RunID: "reviewed-plan-0123456789",
				Plan: runstate.PlanBinding{Path: "plan.tfplan"}}},
		Deployment: project.Deployment{Metadata: project.Metadata{Name: "development"}},
		Template:   template.Contract{Tofu: template.Engine{Directory: "tofu"}}, Adapter: adapter,
	})
	var failure *app.ApplyFailure
	if !errors.As(err, &failure) || result.ExecutionOutcome != "interrupted" ||
		!result.Recovery.InspectionRequired {
		t.Fatalf("evidence failure result=%+v err=%v", result, err)
	}
}

func writePlanFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
