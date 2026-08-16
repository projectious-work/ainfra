package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/source"
)

func TestMCPCreatePlanUsesFixedResolvedSessionAndPropagatesCancellation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	projectRoot := filepath.Join(root, "deployment")
	if err := os.Mkdir(projectRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	writeMCPPlanFile(t, filepath.Join(projectRoot, project.ManifestName), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: mcp-plan
spec:
  template:
    source: local:../template
`)
	templateRoot := filepath.Join(root, "template")
	if err := os.CopyFS(templateRoot, os.DirFS("../../spec/examples/v1/template-example")); err != nil {
		t.Fatal(err)
	}
	cacheRoot := filepath.Join(root, "cache")
	materialized, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	document := lockfile.New(lockfile.Template{Source: "local:../template",
		Resolved: "local:../template", Version: "1.0.0", Digest: materialized.Digest,
		ResolvedAt: time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)})
	if err := lockfile.Write(filepath.Join(projectRoot, lockfile.Filename), document); err != nil {
		t.Fatal(err)
	}
	tofuPath := filepath.Join(root, "tofu")
	fakeTofu := `#!/bin/sh
case "$1" in
  version) printf '%s\n' '{"terraform_version":"1.10.0"}' ;;
  init) mkdir -p .terraform ;;
  plan)
    for argument in "$@"; do
      case "$argument" in -out=*) output="${argument#-out=}" ;; esac
    done
    printf '%s\n' 'saved-plan' > "$output"
    ;;
  show) printf '%s\n' '{"resource_changes":[{"change":{"actions":["create"]}}]}' ;;
  *) exit 91 ;;
esac
`
	writeMCPPlanFile(t, tofuPath, fakeTofu)
	if err := os.Chmod(tofuPath, 0o700); err != nil {
		t.Fatal(err)
	}
	deployment, err := project.Load(project.ResolveOptions{ProjectPath: projectRoot})
	if err != nil {
		t.Fatal(err)
	}
	runsRoot := filepath.Join(root, "runs")
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	session := MCPServeSession{Project: output.Deployment{Name: "mcp-plan", Root: projectRoot},
		project: deployment, settings: config.Settings{Paths: config.Paths{Cache: cacheRoot, Runs: runsRoot},
			Executables: config.Executables{Tofu: tofuPath}},
		planOptions: PlanHostOptions{ParentEnvironment: []string{"HOME=" + t.TempDir(),
			"PATH=" + os.Getenv("PATH")}, Now: func() time.Time { return now },
			Random: func(value []byte) (int, error) {
				return copy(value, strings.Repeat("a", len(value))), nil
			}},
	}
	result, err := session.CreatePlan(context.Background(), "apply")
	if err != nil {
		t.Fatal(err)
	}
	if result.Deployment != session.Project || result.Intent != "apply" ||
		result.PlanDigest == "" || result.RunID != "20260816T120000Z-"+strings.Repeat("61", 16) {
		t.Fatalf("unexpected MCP plan: %+v", result)
	}
	for _, name := range []string{"run.json", "plan-record.json", "plan.json"} {
		if _, err := os.Stat(filepath.Join(runsRoot, result.RunID, name)); err != nil {
			t.Fatalf("missing retained %s: %v", name, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := session.CreatePlan(ctx, "destroy"); err == nil {
		t.Fatal("cancelled MCP destroy planning succeeded")
	}
}

func writeMCPPlanFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
