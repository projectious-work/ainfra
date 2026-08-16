package mcpserver_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/mcpserver"
)

func TestDefaultRegistryIsTypedAndReadOnly(t *testing.T) {
	t.Parallel()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	projectRoot := t.TempDir()
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
`
	if err := os.WriteFile(filepath.Join(projectRoot, "ainfra.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidRunsRoot := filepath.Join(t.TempDir(), "runs")
	if err := os.WriteFile(invalidRunsRoot, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	sessionPolicy, err := app.PrepareMCPServe(app.MCPServeRequest{ProjectPath: projectRoot},
		app.PlanHostOptions{WorkingDirectory: t.TempDir(), HomeDirectory: t.TempDir(),
			CacheDirectory: t.TempDir(), RunDirectory: invalidRunsRoot,
			Environment: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	server := mcpserver.New(sessionPolicy, mcpserver.Options{Build: app.Build{Version: "1.2.3",
		Commit: "0123456789abcdef", BuiltAt: "2026-08-15T00:00:00Z"}, Stderr: &bytes.Buffer{}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverDone := make(chan error, 1)
	go func() { serverDone <- server.Run(ctx, serverTransport) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("close session: %v", err)
		}
	})
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 3 {
		t.Fatalf("unexpected default tools: %+v", tools.Tools)
	}
	for _, tool := range tools.Tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("default tool is not read-only: %+v", tool)
		}
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.version"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok := result.StructuredContent.(map[string]any)
	if !ok || value["apiVersion"] != "ainfra.result/v1" || value["tool"] != "ainfra.version" {
		t.Fatalf("unexpected structured result: %#v", result.StructuredContent)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.project.inspect"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	project, projectOK := value["result"].(map[string]any)
	if !ok || !projectOK || project["name"] != "example" || project["root"] != projectRoot {
		t.Fatalf("unexpected project result: %#v", result.StructuredContent)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.status"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	diagnostics, diagnosticsOK := value["diagnostics"].([]any)
	if !result.IsError || !ok || value["apiVersion"] != "ainfra.result/v1" ||
		value["tool"] != "ainfra.status" || value["ok"] != false ||
		!diagnosticsOK || len(diagnostics) != 1 {
		t.Fatalf("unexpected typed status failure: %#v", result)
	}
}

func TestProtocolVersionIsCurrentPhaseSelection(t *testing.T) {
	t.Parallel()
	if got := mcpserver.ProtocolVersion(); got != "2026-07-28" {
		t.Fatalf("protocol version = %q", got)
	}
}
