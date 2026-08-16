package mcpserver_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/mcpserver"
	"github.com/projectious-work/ainfra/internal/output"
)

func TestDefaultRegistryIsTypedAndReadOnly(t *testing.T) {
	t.Parallel()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	sessionPolicy := app.MCPServeSession{Project: output.Deployment{Name: "example", Root: "/project"}}
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
	if len(tools.Tools) != 2 {
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
	if !ok || !projectOK || project["name"] != "example" || project["root"] != "/project" {
		t.Fatalf("unexpected project result: %#v", result.StructuredContent)
	}
}

func TestProtocolVersionIsCurrentPhaseSelection(t *testing.T) {
	t.Parallel()
	if got := mcpserver.ProtocolVersion(); got != "2026-07-28" {
		t.Fatalf("protocol version = %q", got)
	}
}
