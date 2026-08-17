package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/source"
)

func TestMCPAndCLIApplyShareLifecycleContract(t *testing.T) {
	t.Parallel()
	cli := newApplyEquivalenceFixture(t, "cli", 'a')
	mcpFixture := newApplyEquivalenceFixture(t, "mcp", 'b')

	cliResult, err := applyForDeployment(context.Background(), cli.session.project,
		cli.session.settings, cli.plan.RunID, cli.session.planOptions)
	if err != nil {
		t.Fatalf("CLI application path: %v", err)
	}
	grant := MCPAuthorizationGrant{AuthorizationID: "approval-equivalence",
		Issuer: "operator-1", Caller: "agent-1", ProjectRoot: mcpFixture.session.Project.Root,
		Operation: "apply", PlanID: mcpFixture.plan.RunID,
		PlanDigest: mcpFixture.plan.PlanDigest, Intent: "apply",
		ApprovedAt: mcpFixture.now.Add(-time.Minute), ExpiresAt: mcpFixture.now.Add(time.Minute)}
	mcpFixture.session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	mcpResult, err := mcpFixture.session.ApplyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: mcpFixture.plan.RunID, Caller: "agent-1", Approval: "opaque"})
	if err != nil {
		t.Fatalf("MCP application path: %v", err)
	}

	if cliResult.Operation != mcpResult.Execution.Operation ||
		cliResult.ExecutionOutcome != mcpResult.Execution.ExecutionOutcome ||
		!reflect.DeepEqual(cliResult.EngineReports, mcpResult.Execution.EngineReports) ||
		!reflect.DeepEqual(cliResult.Recovery, mcpResult.Execution.Recovery) {
		t.Fatalf("lifecycle result differs:\nCLI: %+v\nMCP: %+v",
			cliResult, mcpResult.Execution)
	}
	if len(mcpResult.Execution.Evidence) != len(cliResult.Evidence)+1 ||
		!reflect.DeepEqual(cliResult.Evidence, mcpResult.Execution.Evidence[:len(cliResult.Evidence)]) ||
		mcpResult.Execution.Evidence[len(cliResult.Evidence)].Kind != "authorization" {
		t.Fatalf("evidence differs beyond authorization:\nCLI: %+v\nMCP: %+v",
			cliResult.Evidence, mcpResult.Execution.Evidence)
	}
	cliArguments := readApplyEquivalenceFile(t, cli.marker)
	mcpArguments := readApplyEquivalenceFile(t, mcpFixture.marker)
	if cliArguments != mcpArguments || !strings.Contains(cliArguments, "../../plan.tfplan") {
		t.Fatalf("child invocation differs: CLI=%q MCP=%q", cliArguments, mcpArguments)
	}
	cliEvents := readApplyEquivalenceFile(t,
		filepath.Join(cli.session.runsRoot, cli.plan.RunID, "events.jsonl"))
	mcpEvents := readApplyEquivalenceFile(t,
		filepath.Join(mcpFixture.session.runsRoot, mcpFixture.plan.RunID, "events.jsonl"))
	for _, state := range []string{`"state":"started"`, `"state":"succeeded"`} {
		if !strings.Contains(cliEvents, state) || !strings.Contains(mcpEvents, state) {
			t.Fatalf("lifecycle state %s differs: CLI=%q MCP=%q", state, cliEvents, mcpEvents)
		}
	}
	if _, err := os.Stat(filepath.Join(mcpFixture.session.runsRoot,
		mcpFixture.plan.RunID, "authorization.json")); err != nil {
		t.Fatalf("MCP authorization evidence: %v", err)
	}
}

type applyEquivalenceFixture struct {
	session MCPServeSession
	plan    output.Plan
	marker  string
	now     time.Time
}

func newApplyEquivalenceFixture(t *testing.T, name string, randomByte byte) applyEquivalenceFixture {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(root, "deployment")
	if err := os.Mkdir(projectRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	writeMCPPlanFile(t, filepath.Join(projectRoot, project.ManifestName), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: equivalence
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
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	lock := lockfile.New(lockfile.Template{Source: "local:../template",
		Resolved: "local:../template", Version: "1.0.0", Digest: materialized.Digest,
		ResolvedAt: now})
	if err := lockfile.Write(filepath.Join(projectRoot, lockfile.Filename), lock); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "apply-arguments")
	tofuPath := filepath.Join(root, "tofu")
	writeMCPPlanFile(t, tofuPath, `#!/bin/sh
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
  apply) printf '%s\n' "$@" > `+marker+` ;;
  *) exit 91 ;;
esac
`)
	if err := os.Chmod(tofuPath, 0o700); err != nil {
		t.Fatal(err)
	}
	deployment, err := project.Load(project.ResolveOptions{ProjectPath: projectRoot})
	if err != nil {
		t.Fatal(err)
	}
	runsRoot := filepath.Join(root, "runs")
	session := MCPServeSession{Project: output.Deployment{Name: "equivalence", Root: projectRoot},
		project: deployment, runsRoot: runsRoot, cacheRoot: cacheRoot,
		settings: config.Settings{Paths: config.Paths{Cache: cacheRoot, Runs: runsRoot},
			Executables: config.Executables{Tofu: tofuPath}},
		planOptions: PlanHostOptions{ParentEnvironment: []string{"HOME=" + t.TempDir(),
			"PATH=" + os.Getenv("PATH")}, Now: func() time.Time { return now },
			Random: func(value []byte) (int, error) {
				for index := range value {
					value[index] = randomByte
				}
				return len(value), nil
			}},
		now: func() time.Time { return now }, authorizationUses: &sync.Map{},
	}
	plan, err := session.CreatePlan(context.Background(), "apply")
	if err != nil {
		t.Fatalf("create %s plan: %v", name, err)
	}
	return applyEquivalenceFixture{session: session, plan: plan, marker: marker, now: now}
}

func readApplyEquivalenceFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
