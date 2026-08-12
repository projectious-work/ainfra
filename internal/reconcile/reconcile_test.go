package reconcile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/reconcile"
)

func TestRuntimeDirectoryReconciliationIsPlannedAppliedAndIdempotent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeManifest(t, root)
	planner := reconcile.Planner{}
	plan, err := planner.RuntimeDirectory(root)
	if err != nil || len(plan.Actions) != 1 ||
		plan.Actions[0].Kind != reconcile.ActionCreateRuntimeDirectory {
		t.Fatalf("plan=%#v err=%v", plan, err)
	}
	results, err := planner.Apply(plan, reconcile.FileLocker{})
	if err != nil || len(results) != 1 || results[0].Status != "applied" {
		t.Fatalf("results=%#v err=%v", results, err)
	}
	information, err := os.Stat(filepath.Join(root, ".ainfra"))
	if err != nil || information.Mode().Perm() != 0o700 {
		t.Fatalf("runtime mode=%v err=%v", information.Mode(), err)
	}
	second, err := planner.RuntimeDirectory(root)
	if err != nil || len(second.Actions) != 0 {
		t.Fatalf("second plan=%#v err=%v", second, err)
	}
}

func TestRuntimeDirectoryReconciliationRejectsChangedPreconditions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeManifest(t, root)
	planner := reconcile.Planner{}
	plan, err := planner.RuntimeDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".ainfra"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := planner.Apply(plan, reconcile.FileLocker{}); err == nil {
		t.Fatal("stale reconciliation plan unexpectedly applied")
	}
}

func TestRuntimeDirectoryReconciliationRejectsSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeManifest(t, root)
	if err := os.Symlink(t.TempDir(), filepath.Join(root, ".ainfra")); err != nil {
		t.Fatal(err)
	}
	if _, err := (reconcile.Planner{}).RuntimeDirectory(root); err == nil {
		t.Fatal("symlink runtime directory unexpectedly accepted")
	}
}

func TestReconciliationPreservesFailureAndContinues(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeManifest(t, root)
	planner := reconcile.Planner{}
	plan, err := planner.RuntimeDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	plan.Actions = append(plan.Actions, reconcile.Action{
		CheckID: "deployment.second-safe-fix", Path: filepath.Join(root, ".ainfra-second"),
		Kind: reconcile.ActionCreateRuntimeDirectory, Mode: 0o700,
		RollbackLimitation: "remove manually",
	})
	invocations := 0
	failing := reconcile.Planner{
		RecheckPlan: func(string) (reconcile.Plan, error) { return plan, nil },
		ApplyAction: func(reconcile.Action) error {
			invocations++
			if invocations == 1 {
				return os.ErrPermission
			}
			return nil
		},
	}
	results, err := failing.Apply(plan, testLocker{})
	if err != nil {
		t.Fatal(err)
	}
	if invocations != 2 || len(results) != 2 || results[0].Status != "failed" ||
		results[0].Error == "" || results[1].Status != "applied" {
		t.Fatalf("invocations=%d results=%#v", invocations, results)
	}
}

type testLocker struct{}

func (testLocker) Lock(string, string) (func() error, error) {
	return func() error { return nil }, nil
}

func writeManifest(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "ainfra.yaml"), []byte("manifest\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
