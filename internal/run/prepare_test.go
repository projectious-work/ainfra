package run_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
	"github.com/projectious-work/ainfra/internal/tofu"
)

func TestPrepareCreatesPrivateVerifiedWorkspaceAndBindings(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t)
	prepared, err := run.Prepare(fixture.options)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.ID != fixture.options.ID || prepared.TemplateDigest != fixture.lock.Template.Digest || prepared.ExecutableDigest == "" {
		t.Fatalf("unexpected prepared binding: %#v", prepared)
	}
	wantPaths := []string{"backend.hcl", "one.tfvars", "two.tfvars"}
	gotPaths := make([]string, len(prepared.NativeInputs))
	for index, binding := range prepared.NativeInputs {
		gotPaths[index] = binding.Path
		if binding.Digest == "" {
			t.Fatalf("missing input digest: %#v", binding)
		}
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("input order = %#v", gotPaths)
	}
	if digest, err := source.TreeDigest(prepared.Workspace); err != nil || digest != fixture.lock.Template.Digest {
		t.Fatalf("workspace digest = %q err=%v", digest, err)
	}
	info, err := os.Stat(prepared.Root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("run directory mode=%v", info.Mode().Perm())
	}
}

func TestPrepareRefusesReuseAndPoisonedCacheWithoutPartialRun(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t)
	if _, err := run.Prepare(fixture.options); err != nil {
		t.Fatal(err)
	}
	if _, err := run.Prepare(fixture.options); err == nil {
		t.Fatal("existing run ID unexpectedly reused")
	}
	fixture = newFixture(t)
	cacheFile := filepath.Join(fixture.materialized.Path, "ainfra-template.yaml")
	if err := os.WriteFile(cacheFile, []byte("poisoned"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := run.Prepare(fixture.options); err == nil {
		t.Fatal("poisoned cache unexpectedly accepted")
	}
	if _, err := os.Stat(filepath.Join(fixture.options.RunsRoot, fixture.options.ID)); !os.IsNotExist(err) {
		t.Fatalf("partial run was retained: %v", err)
	}
}

func TestLoadReviewedReverifiesBindingsAndRecordsEvents(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t)
	prepared, err := run.Prepare(fixture.options)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(prepared.Root, "plan.tfplan"), "saved plan")
	if _, err := run.PublishPlan(prepared, fixture.options.Deployment.Metadata.Name,
		"1.10.0", time.Now(), structSummary()); err != nil {
		t.Fatal(err)
	}
	reviewed, err := run.LoadReviewed(run.ReviewOptions{ID: fixture.options.ID,
		RunsRoot: fixture.options.RunsRoot, CacheRoot: fixture.options.CacheRoot,
		Deployment: fixture.options.Deployment, Lock: fixture.lock,
		Executable: fixture.options.Executable, EngineVersion: "1.10.0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := run.AppendExecutionEvent(reviewed, "started", time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	exitCode := 0
	if err := run.AppendExecutionEvent(reviewed, "succeeded", time.Now(), &exitCode); err != nil {
		t.Fatal(err)
	}
	events, err := os.ReadFile(filepath.Join(reviewed.Root, "events.jsonl"))
	if err != nil || len(events) == 0 {
		t.Fatalf("events=%q err=%v", events, err)
	}
	writeFile(t, fixture.options.Deployment.Target.ManifestPath, "changed")
	if _, err := run.LoadReviewed(run.ReviewOptions{ID: fixture.options.ID,
		RunsRoot: fixture.options.RunsRoot, CacheRoot: fixture.options.CacheRoot,
		Deployment: fixture.options.Deployment, Lock: fixture.lock,
		Executable: fixture.options.Executable, EngineVersion: "1.10.0"}); err == nil {
		t.Fatal("changed deployment unexpectedly retained plan authorization")
	}
}

func TestLoadReviewedIntentSeparatesApplyAndDestroyAuthorization(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t)
	prepared, err := run.Prepare(fixture.options)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(prepared.Root, "plan.tfplan"), "saved destroy plan")
	if _, err := run.PublishPlanIntent(prepared, fixture.options.Deployment.Metadata.Name,
		"1.10.0", "destroy", time.Now(), tofu.Summary{Delete: 1}); err != nil {
		t.Fatal(err)
	}
	options := run.ReviewOptions{ID: fixture.options.ID,
		RunsRoot: fixture.options.RunsRoot, CacheRoot: fixture.options.CacheRoot,
		Deployment: fixture.options.Deployment, Lock: fixture.lock,
		Executable: fixture.options.Executable, EngineVersion: "1.10.0"}
	if _, err := run.LoadReviewed(options); err == nil {
		t.Fatal("destroy plan unexpectedly authorized apply")
	}
	if _, err := run.LoadReviewedIntent(options, "destroy"); err != nil {
		t.Fatalf("destroy plan was not authorized: %v", err)
	}
}

func structSummary() tofu.Summary { return tofu.Summary{Create: 1} }

type fixture struct {
	options      run.Options
	lock         lockfile.Document
	materialized source.Materialized
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	manifest := filepath.Join(root, project.ManifestName)
	writeFile(t, manifest, "deployment")
	for _, name := range []string{"backend.hcl", "one.tfvars", "two.tfvars"} {
		writeFile(t, filepath.Join(root, name), name)
	}
	templateRoot := t.TempDir()
	writeFile(t, filepath.Join(templateRoot, "ainfra-template.yaml"), "template")
	cacheRoot := t.TempDir()
	materialized, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	document := lockfile.New(lockfile.Template{Source: "local:template", Resolved: "local:template", Version: "1.0.0", Digest: materialized.Digest, ResolvedAt: time.Now()})
	executablePath := filepath.Join(t.TempDir(), "tofu")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(executablePath)
	if err != nil {
		t.Fatal(err)
	}
	deployment := project.Deployment{Target: project.Target{Root: root, ManifestPath: manifest}, Metadata: project.Metadata{Name: "development"}, Inputs: project.Inputs{TofuBackendConfigFiles: []string{"backend.hcl"}, TofuVariableFiles: []string{"one.tfvars", "two.tfvars"}}}
	return fixture{lock: document, materialized: materialized, options: run.Options{ID: "20260813T180000Z-0123456789abcdef", RunsRoot: filepath.Join(t.TempDir(), "runs"), CacheRoot: cacheRoot, Deployment: deployment, Lock: document, Executable: executable}}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
