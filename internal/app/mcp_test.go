package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/doctor"
)

func TestPrepareMCPServeFixesCanonicalProjectAtStartup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	projectRoot := filepath.Join(root, "deployment")
	if err := os.Mkdir(projectRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(projectRoot, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
`)
	planOptions := app.PlanHostOptions{WorkingDirectory: root, HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(),
		Environment: map[string]string{}}
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	authorizationRequest, grant := authorizationFixture(now)
	grant.ProjectRoot = projectRoot
	serveOptions := mcpServeOptions(t, planOptions)
	serveOptions.Authorization = authorizationProvider{grant: grant}
	serveOptions.Now = func() time.Time { return now }
	session, err := app.PrepareMCPServe(context.Background(),
		app.MCPServeRequest{ProjectPath: "deployment"}, serveOptions)
	if err != nil {
		t.Fatal(err)
	}
	if session.Project.Name != "example" || session.Project.Root != projectRoot {
		t.Fatalf("unexpected project: %+v", session.Project)
	}
	status, err := session.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Deployment != session.Project || len(status.Runs) != 0 {
		t.Fatalf("unexpected status: %+v", status)
	}
	doctorResult, err := session.DoctorDeployment()
	if err != nil {
		t.Fatal(err)
	}
	if doctorResult.Scope != "deployment" || len(doctorResult.Findings) == 0 {
		t.Fatalf("unexpected deployment doctor: %+v", doctorResult)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("read-only doctor created runtime directory: %v", err)
	}
	runDoctor := session.DoctorRun()
	if runDoctor.Scope != "run" || runDoctor.Summary.Skip != 1 ||
		len(runDoctor.Findings) != 1 || runDoctor.Findings[0].Status != "skip" {
		t.Fatalf("unexpected run doctor: %+v", runDoctor)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("run doctor created runtime directory: %v", err)
	}
	templateDoctor, err := session.DoctorTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if templateDoctor.Scope != "template" || templateDoctor.Summary.Skip != 1 ||
		len(templateDoctor.Findings) != 1 || templateDoctor.Findings[0].Status != "skip" {
		t.Fatalf("unexpected template doctor: %+v", templateDoctor)
	}
	environmentDoctor := session.DoctorEnvironment()
	if environmentDoctor.Scope != "environment" || len(environmentDoctor.Findings) == 0 ||
		environmentDoctor.EffectiveConfiguration == nil {
		t.Fatalf("unexpected environment doctor: %+v", environmentDoctor)
	}
	reconciliation, err := session.PlanReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if reconciliation.Deployment != session.Project || len(reconciliation.Actions) != 1 ||
		reconciliation.Actions[0].Path != ".ainfra" ||
		reconciliation.Actions[0].Kind != "create_runtime_directory" ||
		reconciliation.Actions[0].Mode != "0700" {
		t.Fatalf("unexpected reconciliation plan: %+v", reconciliation)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("reconciliation planning created runtime directory: %v", err)
	}
	authorizationRequest.ProjectRoot = "/client-controlled-root"
	authorization, err := session.AuthorizeMutation(context.Background(), authorizationRequest)
	if err != nil || authorization.AuthorizationID != grant.AuthorizationID {
		t.Fatalf("fixed-project authorization: %+v, %v", authorization, err)
	}
	missing := authorizationRequest
	missing.Approval = strings.Repeat("x", (64<<10)+1)
	if _, err := session.AuthorizeMutation(context.Background(), missing); err == nil {
		t.Fatal("oversized approval material succeeded")
	}
}

func TestPrepareMCPServeValidatesCapabilityAllowlist(t *testing.T) {
	t.Parallel()
	for _, capabilities := range [][]string{
		{"unknown"},
		{app.MCPPlanningCapability, app.MCPPlanningCapability},
		{app.MCPDeploymentCapability, app.MCPDeploymentCapability},
		{app.MCPDestructionCapability},
	} {
		_, err := app.PrepareMCPServe(context.Background(),
			app.MCPServeRequest{Capabilities: capabilities}, mcpServeOptions(t,
				app.PlanHostOptions{Environment: map[string]string{}}))
		if err == nil {
			t.Fatalf("invalid capabilities succeeded: %v", capabilities)
		}
	}
	projectRoot := t.TempDir()
	write(t, filepath.Join(projectRoot, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: capability-test
spec:
  template:
    source: local:../template
`)
	plan := app.PlanHostOptions{WorkingDirectory: projectRoot, HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(), Environment: map[string]string{}}
	session, err := app.PrepareMCPServe(context.Background(), app.MCPServeRequest{
		ProjectPath: projectRoot, Capabilities: []string{app.MCPDeploymentCapability},
	}, mcpServeOptions(t, plan))
	if err != nil || !session.CapabilityEnabled(app.MCPDeploymentCapability) {
		t.Fatalf("deployment capability was not enabled: %+v, %v", session, err)
	}
}

func TestPrepareMCPServeRejectsConflictingEnvironmentProject(t *testing.T) {
	t.Parallel()
	planOptions := app.PlanHostOptions{Environment: map[string]string{"AINFRA_PROJECT": "/two"}}
	_, err := app.PrepareMCPServe(context.Background(),
		app.MCPServeRequest{ProjectPath: "/one"}, mcpServeOptions(t, planOptions))
	if err == nil {
		t.Fatal("conflicting project selections succeeded")
	}
}

func TestPrepareMCPServeRejectsSymlinkProject(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	projectRoot := filepath.Join(root, "deployment")
	if err := os.Mkdir(projectRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(projectRoot, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
`)
	link := filepath.Join(root, "project-link")
	if err := os.Symlink(projectRoot, link); err != nil {
		t.Fatal(err)
	}
	planOptions := app.PlanHostOptions{WorkingDirectory: root, HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(),
		Environment: map[string]string{}}
	_, err := app.PrepareMCPServe(context.Background(),
		app.MCPServeRequest{ProjectPath: link}, mcpServeOptions(t, planOptions))
	if err == nil {
		t.Fatal("symlink project selection succeeded")
	}
}

func mcpServeOptions(t *testing.T, plan app.PlanHostOptions) app.MCPServeOptions {
	t.Helper()
	return app.MCPServeOptions{Plan: plan, Doctor: app.DoctorEnvironmentOptions{
		GOOS: "linux", GOARCH: "arm64", WorkingDirectory: plan.WorkingDirectory,
		HomeDirectory: plan.HomeDirectory, CacheDirectory: plan.CacheDirectory,
		RunDirectory: plan.RunDirectory, Environment: plan.Environment,
		InspectExecutable: func(_ context.Context, name, _ string) (doctor.ExecutableFact, error) {
			return doctor.ExecutableFact{Path: "/tools/" + name, Version: name + " 1.0"}, nil
		},
	}}
}
