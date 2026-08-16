package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
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
	session, err := app.PrepareMCPServe(app.MCPServeRequest{ProjectPath: "deployment"},
		app.PlanHostOptions{WorkingDirectory: root, HomeDirectory: t.TempDir(),
			CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(),
			Environment: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if session.Project.Name != "example" || session.Project.Root != projectRoot {
		t.Fatalf("unexpected project: %+v", session.Project)
	}
}

func TestPrepareMCPServeRejectsConflictingEnvironmentProject(t *testing.T) {
	t.Parallel()
	_, err := app.PrepareMCPServe(app.MCPServeRequest{ProjectPath: "/one"},
		app.PlanHostOptions{Environment: map[string]string{"AINFRA_PROJECT": "/two"}})
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
	_, err := app.PrepareMCPServe(app.MCPServeRequest{ProjectPath: link},
		app.PlanHostOptions{WorkingDirectory: root, HomeDirectory: t.TempDir(),
			CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(),
			Environment: map[string]string{}})
	if err == nil {
		t.Fatal("symlink project selection succeeded")
	}
}
