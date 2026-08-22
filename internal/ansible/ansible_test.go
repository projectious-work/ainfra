package ansible

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/security"
)

func TestVerifyConverged(t *testing.T) {
	stats := Stats{Changed: map[string]int{}, Dark: map[string]int{}, Failures: map[string]int{}, Processed: map[string]int{"a": 1, "b": 1}}
	if err := VerifyConverged(stats, []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	stats.Changed["a"] = 1
	if err := VerifyConverged(stats, []string{"a", "b"}); err == nil {
		t.Fatal("accepted changed host")
	}
}

func TestVersionAcceptsCurrentAndLegacyRunnerOutput(t *testing.T) {
	t.Parallel()
	for _, output := range []string{"2.4.3\n", "ansible-runner 2.4.3\n"} {
		output := output
		t.Run(strings.TrimSpace(output), func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			executablePath := filepath.Join(root, "ansible-runner")
			if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			executable, err := security.ResolveExecutable(executablePath)
			if err != nil {
				t.Fatal(err)
			}
			adapter := Adapter{Executable: executable, Run: func(_ context.Context,
				request childexec.Request,
			) (childexec.Result, error) {
				_, _ = request.IO.Stdout.Write([]byte(output))
				return childexec.Result{Started: true}, nil
			}}
			version, err := adapter.Version(context.Background(), root)
			if err != nil || version != "2.4.3" {
				t.Fatalf("version=%q err=%v", version, err)
			}
		})
	}
}

func TestVerifyConvergedRequiresExactHosts(t *testing.T) {
	stats := Stats{Changed: map[string]int{}, Dark: map[string]int{}, Failures: map[string]int{}, Processed: map[string]int{"a": 1}}
	if err := VerifyConverged(stats, []string{"a", "b"}); err == nil {
		t.Fatal("accepted incomplete host set")
	}
}

func TestConfigureUsesControlledArgumentsAndNestedStats(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"private", "project", "artifacts", "inputs"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"inventory.yaml", "project/site.yml", "inputs/vars.yml"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executablePath := filepath.Join(root, "ansible-runner")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(executablePath)
	if err != nil {
		t.Fatal(err)
	}
	adapter := Adapter{Executable: executable, Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
		wantPrefix := []string{"run", "private", "--project-dir", "project", "--inventory", "inventory.yaml", "--artifact-dir", "artifacts", "--playbook", "site.yml", "--cmdline"}
		if !reflect.DeepEqual(request.Args[:len(wantPrefix)], wantPrefix) {
			t.Fatalf("args=%#v", request.Args)
		}
		if !strings.Contains(request.Args[len(request.Args)-1], "@"+filepath.Join(root, "inputs", "vars.yml")) ||
			!strings.Contains(request.Args[len(request.Args)-1], "--check") {
			t.Fatalf("cmdline=%q", request.Args[len(request.Args)-1])
		}
		events := filepath.Join(root, "artifacts", "identifier", "job_events")
		if err := os.MkdirAll(events, 0o700); err != nil {
			t.Fatal(err)
		}
		contents := `{"event":"playbook_on_stats","event_data":{"changed":{},"dark":{},"failures":{},"ok":{"host":1},"processed":{"host":1},"skipped":{}}}`
		if err := os.WriteFile(filepath.Join(events, "1.json"), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		return childexec.Result{Started: true}, nil
	}}
	outcome, err := adapter.Configure(context.Background(), root, "private", "project", "inventory.yaml", "site.yml", "artifacts", []string{"inputs/vars.yml"}, true)
	if err != nil || outcome.Stats.Processed["host"] != 1 {
		t.Fatalf("outcome=%+v err=%v", outcome, err)
	}
}

func TestReadStatsRejectsMultipleSummaries(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	contents := []byte(`{"event":"playbook_on_stats","event_data":{"changed":{},"dark":{},"failures":{},"ok":{},"processed":{},"skipped":{}}}`)
	for _, name := range []string{"one.json", "two.json"} {
		if err := os.WriteFile(filepath.Join(root, "artifacts", name), contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ReadStats(root, "artifacts"); err == nil {
		t.Fatal("accepted multiple terminal summaries")
	}
}

func TestReadStatsRejectsCorruptAndSymlinkedEvidence(t *testing.T) {
	root := t.TempDir()
	artifacts := filepath.Join(root, "artifacts")
	if err := os.Mkdir(artifacts, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artifacts, "stats.json"), []byte(`{"event":"playbook_on_stats","event_data":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadStats(root, "artifacts"); err == nil {
		t.Fatal("accepted corrupt evidence")
	}
	if err := os.Remove(filepath.Join(artifacts, "stats.json")); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "outside.json")
	if err := os.WriteFile(target, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(artifacts, "linked.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadStats(root, "artifacts"); err == nil {
		t.Fatal("accepted symlinked evidence")
	}
}

func TestAdapterPropagatesCancellation(t *testing.T) {
	root := t.TempDir()
	executablePath := filepath.Join(root, "ansible-runner")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(executablePath)
	if err != nil {
		t.Fatal(err)
	}
	adapter := Adapter{Executable: executable, Run: func(context.Context, childexec.Request) (childexec.Result, error) {
		return childexec.Result{Started: true, Cancelled: true}, nil
	}}
	if _, err := adapter.execute(context.Background(), root, ".", []string{"run"}, childexec.IOPolicy{}); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("cancellation error=%v", err)
	}
}
