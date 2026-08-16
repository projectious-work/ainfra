package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
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
	deploymentContract := session.InspectDeployment()
	if deploymentContract.Name != "mcp-plan" ||
		deploymentContract.Template.Source != "local:../template" {
		t.Fatalf("unexpected deployment contract: %+v", deploymentContract)
	}
	lockPath := filepath.Join(projectRoot, lockfile.Filename)
	if err := os.Remove(lockPath); err != nil {
		t.Fatal(err)
	}
	templatePlan, err := session.PlanTemplateLock(context.Background(), "lock")
	if err != nil {
		t.Fatal(err)
	}
	if !templatePlan.Changed || templatePlan.ContentDigest != materialized.Digest {
		t.Fatalf("unexpected template lock plan: %+v", templatePlan)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("template lock plan published a lock: %v", err)
	}
	if err := lockfile.Write(lockPath, document); err != nil {
		t.Fatal(err)
	}
	templatePlan, err = session.PlanTemplateLock(context.Background(), "update")
	if err != nil || templatePlan.Changed {
		t.Fatalf("unchanged template update plan: %+v, %v", templatePlan, err)
	}
	templateContract, err := session.InspectTemplate()
	if err != nil {
		t.Fatal(err)
	}
	encodedContract, err := json.Marshal(templateContract)
	if err != nil {
		t.Fatal(err)
	}
	if templateContract.Name == "" || templateContract.Digest != materialized.Digest ||
		bytes.Contains(encodedContract, []byte(root)) || bytes.Contains(encodedContract, []byte(cacheRoot)) {
		t.Fatalf("unsafe or incomplete template contract: %s", encodedContract)
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
	randomByte = 'c'
	session.now = func() time.Time { return now }
	destroyPlan, err := session.CreatePlan(context.Background(), "destroy")
	if err != nil {
		t.Fatal(err)
	}
	grant.PlanID, grant.PlanDigest = destroyPlan.RunID, destroyPlan.PlanDigest
	grant.Operation, grant.Intent = "apply", "apply"
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	if _, err := session.DestroyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: destroyPlan.RunID, Caller: "agent-1", Approval: "apply-approval"}); err == nil {
		t.Fatal("apply-intent approval authorized destruction")
	}
	grant.Operation, grant.Intent = "destroy", "destroy"
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	destroyed, err := session.DestroyAuthorized(context.Background(), MCPApplyRequest{
		PlanID: destroyPlan.RunID, Caller: "agent-1", Approval: "destroy-approval"})
	if err != nil {
		t.Fatal(err)
	}
	if destroyed.Execution.Operation != "destroy" ||
		destroyed.Execution.ExecutionOutcome != "succeeded" ||
		destroyed.Authorization.AuthorizationID != grant.AuthorizationID {
		t.Fatalf("unexpected authorized destroy: %+v", destroyed)
	}
	contents, err = os.ReadFile(filepath.Join(runsRoot, destroyPlan.RunID,
		"authorization.json"))
	if err != nil || strings.Contains(string(contents), "destroy-approval") ||
		!strings.Contains(string(contents), `"operation": "destroy"`) {
		t.Fatalf("invalid destroy authorization evidence: %s, %v", contents, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := session.CreatePlan(ctx, "destroy"); err == nil {
		t.Fatal("cancelled MCP destroy planning succeeded")
	}
}

func TestMCPTemplatePlanPropagatesCancellationToGitAcquisition(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeMCPPlanFile(t, filepath.Join(projectRoot, project.ManifestName), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: mcp-template-plan
spec:
  template:
    source: git::https://example.invalid/template.git
    ref: main
`)
	deployment, err := project.Load(project.ResolveOptions{ProjectPath: projectRoot})
	if err != nil {
		t.Fatal(err)
	}
	acquired := false
	session := MCPServeSession{project: deployment, cacheRoot: t.TempDir(),
		templateOptions: TemplateLockOptions{AcquireGit: func(ctx context.Context,
			_ source.Reference, _, _ string,
		) (source.GitAcquisition, error) {
			acquired = true
			return source.GitAcquisition{}, ctx.Err()
		}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := session.PlanTemplateLock(ctx, "lock"); !acquired || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled template plan: acquired=%v err=%v", acquired, err)
	}
}

func TestMCPReconciliationRequiresExactIndependentApproval(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeMCPPlanFile(t, filepath.Join(root, project.ManifestName), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: mcp-reconcile
spec:
  template:
    source: local:../template
`)
	deployment, err := project.Load(project.ResolveOptions{ProjectPath: root})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	session := MCPServeSession{Project: output.Deployment{Name: "mcp-reconcile", Root: root},
		project: deployment, cacheRoot: t.TempDir(), now: func() time.Time { return now },
		authorizationUses: &sync.Map{}}
	plan, err := session.PlanReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	encodedPlan, err := json.Marshal(plan)
	if err != nil || bytes.Contains(encodedPlan, []byte(session.cacheRoot)) {
		t.Fatalf("reconciliation plan leaked host cache path: %s, %v", encodedPlan, err)
	}
	repeated, err := session.PlanReconciliation()
	if err != nil || plan.PlanID != repeated.PlanID || plan.PlanDigest != repeated.PlanDigest ||
		len(plan.Actions) != 1 {
		t.Fatalf("unstable reconciliation binding: %+v %+v %v", plan, repeated, err)
	}
	if _, err := session.ReconcileAuthorized(context.Background(), MCPExecutionRequest{
		PlanID: plan.PlanID, Caller: "agent-1", Approval: "opaque"}); err == nil {
		t.Fatal("reconciliation without independent provider succeeded")
	}
	if _, err := os.Stat(filepath.Join(root, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("refused reconciliation mutated project: %v", err)
	}
	grant := MCPAuthorizationGrant{AuthorizationID: "approval-reconcile-1", Issuer: "operator-1",
		Caller: "agent-1", ProjectRoot: root, Operation: "reconcile", PlanID: plan.PlanID,
		PlanDigest: plan.PlanDigest, Intent: "reconcile", ApprovedAt: now.Add(-time.Minute),
		ExpiresAt: now.Add(time.Minute)}
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	session.reconcilePlanner = reconcile.Planner{ApplyAction: func(reconcile.Action) error {
		return errors.New("sensitive /host/path")
	}}
	failed, err := session.ReconcileAuthorized(context.Background(), MCPExecutionRequest{
		PlanID: plan.PlanID, Caller: "agent-1", Approval: "opaque"})
	if err == nil || strings.Contains(err.Error(), "/host/path") || len(failed.Results) != 1 ||
		failed.Results[0].Error != "reconciliation action failed" {
		t.Fatalf("unsanitized reconciliation failure: %+v, %v", failed, err)
	}
	session.reconcilePlanner = reconcile.Planner{}
	if _, err := session.ReconcileAuthorized(context.Background(), MCPExecutionRequest{
		PlanID: plan.PlanID, Caller: "agent-1", Approval: "opaque"}); err == nil ||
		!strings.Contains(err.Error(), "already consumed") {
		t.Fatalf("replayed reconciliation approval succeeded: %v", err)
	}
	grant.AuthorizationID = "approval-reconcile-2"
	session.authorization = mcpPlanAuthorizationProvider{grant: grant}
	result, err := session.ReconcileAuthorized(context.Background(), MCPExecutionRequest{
		PlanID: plan.PlanID, Caller: "agent-1", Approval: "opaque"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PlanDigest != plan.PlanDigest ||
		result.Authorization.AuthorizationID != grant.AuthorizationID || len(result.Results) != 1 ||
		result.Results[0].Status != "applied" || result.Results[0].Action.Path != ".ainfra" {
		t.Fatalf("unexpected reconciliation result: %+v", result)
	}
	information, err := os.Stat(filepath.Join(root, ".ainfra"))
	if err != nil || !information.IsDir() || information.Mode().Perm() != 0o700 {
		t.Fatalf("reconciliation result mode=%v err=%v", information, err)
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
