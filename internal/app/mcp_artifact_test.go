package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/inventory"
)

const retainedRunID = "20260816T120000Z-0123456789abcdef0123456789abcdef"

func TestMCPRetainedArtifactReadersValidateBoundSanitizedArtifacts(t *testing.T) {
	t.Parallel()
	session, runRoot := retainedArtifactSession(t, "example")
	result, err := session.ReadOutput(retainedRunID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Deployment.Name != "example" || result.RunID != retainedRunID ||
		result.Output.Hosts["node"].Connection.Type != "local" {
		t.Fatalf("unexpected retained output: %+v", result)
	}
	inventoryResult, err := session.ReadInventory(retainedRunID)
	if err != nil {
		t.Fatal(err)
	}
	if inventoryResult.MediaType != "application/yaml" ||
		!strings.Contains(inventoryResult.Content, "ansible_connection: local") {
		t.Fatalf("unexpected retained inventory: %+v", inventoryResult)
	}
	if _, err := os.Stat(filepath.Join(runRoot, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("artifact read created runtime state: %v", err)
	}
}

func TestMCPRetainedArtifactReadersFailClosed(t *testing.T) {
	t.Parallel()
	t.Run("traversal", func(t *testing.T) {
		session, _ := retainedArtifactSession(t, "example")
		if _, err := session.ReadOutput("../" + retainedRunID); err == nil {
			t.Fatal("traversal run ID succeeded")
		}
	})
	t.Run("cross project binding", func(t *testing.T) {
		session, _ := retainedArtifactSession(t, "other")
		if _, err := session.ReadOutput(retainedRunID); err == nil {
			t.Fatal("cross-project retained output succeeded")
		}
	})
	t.Run("public artifact", func(t *testing.T) {
		session, root := retainedArtifactSession(t, "example")
		if err := os.Chmod(filepath.Join(root, "output.json"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := session.ReadOutput(retainedRunID); err == nil {
			t.Fatal("public retained output succeeded")
		}
	})
	t.Run("tampered inventory", func(t *testing.T) {
		session, root := retainedArtifactSession(t, "example")
		write(t, filepath.Join(root, "inventory.yaml"), "all: {}\n")
		if _, err := session.ReadInventory(retainedRunID); err == nil {
			t.Fatal("mismatched retained inventory succeeded")
		}
	})
	t.Run("secret-shaped output", func(t *testing.T) {
		session, root := retainedArtifactSession(t, "example")
		write(t, filepath.Join(root, "output.json"), `{"schema_version":"1","hosts":{"node":{"groups":["all"],"connection":{"type":"ssh","address":"ssh://user:password@example.test","user":"root","port":22}}}}`)
		if _, err := session.ReadOutput(retainedRunID); err == nil {
			t.Fatal("secret-shaped retained output succeeded")
		}
	})
}

func retainedArtifactSession(t *testing.T, planDeployment string) (app.MCPServeSession, string) {
	t.Helper()
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
	runsRoot := filepath.Join(root, "runs")
	runRoot := filepath.Join(runsRoot, retainedRunID)
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(runRoot, "run.json"), `{"schemaVersion":1,"runId":"`+retainedRunID+`","operation":"apply","state":"succeeded","createdAt":"2026-08-16T12:00:00Z","planRecord":"plan-record.json"}`)
	write(t, filepath.Join(runRoot, "plan-record.json"), `{"schemaVersion":1,"runId":"`+retainedRunID+`","intent":"apply","deployment":{"name":"`+planDeployment+`","digest":"sha256:test"},"template":{"source":"local:../template","digest":"sha256:test"},"inputs":[],"engine":{"name":"opentofu","version":"1","executableDigest":"sha256:test"},"plan":{"path":"plan.tfplan","digest":"sha256:test","summaryPath":"plan.json"}}`)
	standardJSON := `{"schema_version":"1","hosts":{"node":{"groups":["all"],"connection":{"type":"local"}}}}`
	write(t, filepath.Join(runRoot, "output.json"), standardJSON)
	standard, err := inventory.Parse([]byte(standardJSON))
	if err != nil {
		t.Fatal(err)
	}
	inventoryYAML, err := inventory.YAML(standard)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(runRoot, "inventory.yaml"), string(inventoryYAML))
	planOptions := app.PlanHostOptions{WorkingDirectory: root, HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: runsRoot, Environment: map[string]string{}}
	session, err := app.PrepareMCPServe(context.Background(),
		app.MCPServeRequest{ProjectPath: projectRoot}, mcpServeOptions(t, planOptions))
	if err != nil {
		t.Fatal(err)
	}
	return session, runRoot
}
