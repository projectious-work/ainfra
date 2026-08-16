package mcpserver_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/doctor"
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
	planOptions := app.PlanHostOptions{WorkingDirectory: t.TempDir(), HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: invalidRunsRoot,
		Environment: map[string]string{}}
	var inspections atomic.Int32
	sessionPolicy, err := app.PrepareMCPServe(context.Background(),
		app.MCPServeRequest{ProjectPath: projectRoot}, app.MCPServeOptions{
			Plan: planOptions, Doctor: app.DoctorEnvironmentOptions{
				GOOS: "linux", GOARCH: "arm64", WorkingDirectory: planOptions.WorkingDirectory,
				HomeDirectory: planOptions.HomeDirectory, CacheDirectory: planOptions.CacheDirectory,
				RunDirectory: planOptions.RunDirectory, Environment: map[string]string{},
				InspectExecutable: func(_ context.Context, name, _ string) (doctor.ExecutableFact, error) {
					inspections.Add(1)
					return doctor.ExecutableFact{Path: "/tools/" + name, Version: name + " 1.0"}, nil
				},
			},
		})
	if err != nil {
		t.Fatal(err)
	}
	if inspections.Load() != 4 {
		t.Fatalf("startup executable inspections = %d", inspections.Load())
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
	if len(tools.Tools) != 9 {
		t.Fatalf("unexpected default tools: %+v", tools.Tools)
	}
	for _, tool := range tools.Tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("default tool is not read-only: %+v", tool)
		}
	}
	resources, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources.Resources) != 11 {
		t.Fatalf("unexpected resources: %+v", resources.Resources)
	}
	catalog, err := session.ReadResource(ctx,
		&mcp.ReadResourceParams{URI: "ainfra://contracts/v1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Contents) != 1 ||
		!bytes.Contains([]byte(catalog.Contents[0].Text), []byte(`"apiVersion":"ainfra.contracts/v1"`)) {
		t.Fatalf("unexpected contract catalog: %+v", catalog)
	}
	schema, err := session.ReadResource(ctx,
		&mcp.ReadResourceParams{URI: "ainfra://schemas/v1/ainfra.schema.json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Contents) != 1 || schema.Contents[0].MIMEType != "application/schema+json" ||
		!bytes.Contains([]byte(schema.Contents[0].Text), []byte(`"$schema"`)) {
		t.Fatalf("unexpected schema resource: %+v", schema)
	}
	if _, err := session.ReadResource(ctx,
		&mcp.ReadResourceParams{URI: "ainfra://schemas/v1/../go.mod"}); err == nil {
		t.Fatal("unknown resource URI succeeded")
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
	result, err = session.CallTool(ctx,
		&mcp.CallToolParams{Name: "ainfra.doctor.deployment"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	doctor, doctorOK := value["result"].(map[string]any)
	if !ok || !doctorOK || doctor["scope"] != "deployment" {
		t.Fatalf("unexpected deployment doctor result: %#v", result)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.run"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	runDoctor, runDoctorOK := value["result"].(map[string]any)
	if result.IsError || !ok || !runDoctorOK || runDoctor["scope"] != "run" {
		t.Fatalf("unexpected run doctor result: %#v", result)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.template"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	templateDoctor, templateDoctorOK := value["result"].(map[string]any)
	if result.IsError || !ok || !templateDoctorOK || templateDoctor["scope"] != "template" {
		t.Fatalf("unexpected template doctor result: %#v", result)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.environment"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.StructuredContent.(map[string]any)
	environmentDoctor, environmentDoctorOK := value["result"].(map[string]any)
	if result.IsError || !ok || !environmentDoctorOK || environmentDoctor["scope"] != "environment" {
		t.Fatalf("unexpected environment doctor result: %#v", result)
	}
	if inspections.Load() != 4 {
		t.Fatalf("environment snapshot was recomputed: %d inspections", inspections.Load())
	}
	for _, name := range []string{"ainfra.output.read", "ainfra.inventory.read"} {
		result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: name,
			Arguments: map[string]any{"runId": "20260816T120000Z-0123456789abcdef0123456789abcdef"}})
		if err != nil {
			t.Fatal(err)
		}
		value, ok = result.StructuredContent.(map[string]any)
		if !result.IsError || !ok || value["tool"] != name || value["ok"] != false {
			t.Fatalf("unexpected typed retained artifact failure: %#v", result)
		}
	}
}

func TestProtocolVersionIsCurrentPhaseSelection(t *testing.T) {
	t.Parallel()
	if got := mcpserver.ProtocolVersion(); got != "2026-07-28" {
		t.Fatalf("protocol version = %q", got)
	}
}

func TestPlanningCapabilityDisclosesOnlyNonApplyingPreview(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: planning-test
spec:
  template:
    source: local:../template
`
	if err := os.WriteFile(filepath.Join(projectRoot, "ainfra.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	planOptions := app.PlanHostOptions{WorkingDirectory: t.TempDir(), HomeDirectory: t.TempDir(),
		CacheDirectory: t.TempDir(), RunDirectory: t.TempDir(), Environment: map[string]string{}}
	policy, err := app.PrepareMCPServe(context.Background(), app.MCPServeRequest{
		ProjectPath: projectRoot, Capabilities: []string{app.MCPPlanningCapability},
	}, app.MCPServeOptions{Plan: planOptions, Doctor: app.DoctorEnvironmentOptions{
		GOOS: "linux", GOARCH: "arm64", WorkingDirectory: planOptions.WorkingDirectory,
		HomeDirectory: planOptions.HomeDirectory, CacheDirectory: planOptions.CacheDirectory,
		RunDirectory: planOptions.RunDirectory, Environment: map[string]string{},
		InspectExecutable: func(_ context.Context, name, _ string) (doctor.ExecutableFact, error) {
			return doctor.ExecutableFact{Path: "/tools/" + name, Version: name + " 1.0"}, nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := mcpserver.New(policy, mcpserver.Options{Build: app.Build{Version: "1"},
		Stderr: &bytes.Buffer{}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Run(ctx, serverTransport) }()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 11 {
		t.Fatalf("planning registry: %+v", tools.Tools)
	}
	var createPlan *mcp.Tool
	for _, tool := range tools.Tools {
		if tool.Name == "ainfra.plan.create" {
			createPlan = tool
			break
		}
	}
	if createPlan == nil || createPlan.Annotations == nil ||
		createPlan.Annotations.ReadOnlyHint ||
		createPlan.Annotations.DestructiveHint == nil ||
		*createPlan.Annotations.DestructiveHint || createPlan.Annotations.IdempotentHint {
		t.Fatalf("unexpected plan annotations: %+v", createPlan)
	}
	result, err := session.CallTool(ctx,
		&mcp.CallToolParams{Name: "ainfra.reconciliation.plan"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	plan, planOK := structured["result"].(map[string]any)
	actions, actionsOK := plan["actions"].([]any)
	if result.IsError || !ok || !planOK || !actionsOK || len(actions) != 1 {
		t.Fatalf("unexpected reconciliation plan: %#v", result)
	}
	action, actionOK := actions[0].(map[string]any)
	if !actionOK || action["path"] != ".ainfra" || action["mode"] != "0700" {
		t.Fatalf("unexpected reconciliation action: %#v", actions[0])
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".ainfra")); !os.IsNotExist(err) {
		t.Fatalf("planning tool applied reconciliation: %v", err)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.plan.create",
		Arguments: map[string]any{"intent": "invalid"}})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	if !result.IsError || !ok || structured["tool"] != "ainfra.plan.create" ||
		structured["ok"] != false {
		t.Fatalf("unexpected invalid plan result: %#v", result)
	}
}
