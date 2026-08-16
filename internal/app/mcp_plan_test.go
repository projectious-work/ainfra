package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	runstate "github.com/projectious-work/ainfra/internal/run"
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
	apply) exit 0 ;;
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
	randomByte := byte('a')
	session := MCPServeSession{Project: output.Deployment{Name: "mcp-plan", Root: projectRoot},
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
		now: func() time.Time { return now },
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
	grant := MCPAuthorizationGrant{AuthorizationID: "approval-1", Issuer: "operator-1",
		Caller: "agent-1", ProjectRoot: projectRoot, Operation: "apply",
		PlanID: result.RunID, PlanDigest: result.PlanDigest, Intent: "apply",
		ApprovedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	applied, err := session.ApplyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: result.RunID, Caller: "agent-1", Approval: "opaque-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if applied.Execution.Operation != "apply" ||
		applied.Execution.ExecutionOutcome != "succeeded" ||
		applied.Authorization.AuthorizationID != grant.AuthorizationID {
		t.Fatalf("unexpected authorized apply: %+v", applied)
	}
	authorizationPath := filepath.Join(runsRoot, result.RunID, "authorization.json")
	contents, err := os.ReadFile(authorizationPath)
	if err != nil {
		t.Fatal(err)
	}
	var retained runstate.AuthorizationRecord
	if json.Unmarshal(contents, &retained) != nil || retained.PlanDigest != result.PlanDigest ||
		strings.Contains(string(contents), "opaque-secret") {
		t.Fatalf("invalid retained authorization: %s", contents)
	}
	if _, err := session.ApplyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: result.RunID, Caller: "agent-1", Approval: "opaque-secret"}); err == nil {
		t.Fatal("authorized plan replay succeeded")
	}
	randomByte = 'b'
	second, err := session.CreatePlan(context.Background(), "apply")
	if err != nil {
		t.Fatal(err)
	}
	grant.PlanID, grant.PlanDigest = second.RunID, second.PlanDigest
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	nowCalls := 0
	session.now = func() time.Time {
		nowCalls++
		if nowCalls == 1 {
			return now
		}
		return now.Add(2 * time.Minute)
	}
	if _, err := session.ApplyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: second.RunID, Caller: "agent-1", Approval: "opaque-secret"}); err == nil ||
		!strings.Contains(err.Error(), "expired before execution") {
		t.Fatalf("authorization expiry during preparation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runsRoot, second.RunID,
		"authorization.json")); !os.IsNotExist(err) {
		t.Fatalf("expired authorization evidence exists: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := session.CreatePlan(ctx, "destroy"); err == nil {
		t.Fatal("cancelled MCP destroy planning succeeded")
	}
}

type mcpPlanAuthorizationProvider struct{ grant MCPAuthorizationGrant }

func (provider mcpPlanAuthorizationProvider) Verify(context.Context,
	string,
) (MCPAuthorizationGrant, error) {
	return provider.grant, nil
}

func writeMCPPlanFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
