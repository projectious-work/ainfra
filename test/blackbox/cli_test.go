package blackbox_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var binary string

func TestMain(m *testing.M) {
	temporary, err := os.MkdirTemp("", "ainfra-blackbox-")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(temporary, "ainfra")
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		output, outputErr := exec.Command("go", "env", "GOMODCACHE").Output()
		if outputErr != nil {
			panic(outputErr)
		}
		moduleCache = strings.TrimSpace(string(output))
	}
	build := exec.Command("go", "build", "-o", binary, "../../cmd/ainfra")
	build.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + temporary,
		"GOCACHE=" + filepath.Join(temporary, "go-build"),
		"GOMODCACHE=" + moduleCache,
		"GOPROXY=off",
		"GOTOOLCHAIN=local",
	}
	if output, buildErr := build.CombinedOutput(); buildErr != nil {
		if _, writeErr := os.Stderr.Write(output); writeErr != nil {
			panic(writeErr)
		}
		os.Exit(1)
	}
	code := m.Run()
	if err := os.RemoveAll(temporary); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestVersionJSON(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "--format", "json", "version")
	command.Dir = t.TempDir()
	command.Env = []string{"TERM=dumb", "HOME=" + t.TempDir(), "XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("run: %v, stderr: %s", err, stderr.String())
	}
	var envelope struct {
		APIVersion string `json:"apiVersion"`
		Command    string `json:"command"`
		OK         bool   `json:"ok"`
		Result     struct {
			Version                   string `json:"version"`
			Commit                    string `json:"commit"`
			BuiltAt                   string `json:"builtAt"`
			GoVersion                 string `json:"goVersion"`
			Platform                  string `json:"platform"`
			SupportedContractVersions struct {
				DocumentAPIVersions          []string `json:"documentApiVersions"`
				ResultAPIVersions            []string `json:"resultApiVersions"`
				StandardOutputSchemaVersions []string `json:"standardOutputSchemaVersions"`
			} `json:"supportedContractVersions"`
		} `json:"result"`
		Diagnostics []any `json:"diagnostics"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if envelope.APIVersion != "ainfra.result/v1" || envelope.Command != "version" || !envelope.OK {
		t.Errorf("unexpected envelope: %+v", envelope)
	}
	wantPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if envelope.Result.Platform != wantPlatform {
		t.Errorf("platform = %q, want %q", envelope.Result.Platform, wantPlatform)
	}
	if envelope.Result.Version == "" || envelope.Result.Commit == "" || envelope.Result.GoVersion == "" {
		t.Errorf("incomplete version result: %+v", envelope.Result)
	}
	if _, err := time.Parse(time.RFC3339, envelope.Result.BuiltAt); err != nil {
		t.Errorf("builtAt is not RFC 3339: %q", envelope.Result.BuiltAt)
	}
	if envelope.Diagnostics == nil || len(envelope.Diagnostics) != 0 {
		t.Errorf("diagnostics = %#v", envelope.Diagnostics)
	}
	contracts := envelope.Result.SupportedContractVersions
	if len(contracts.DocumentAPIVersions) != 1 || contracts.DocumentAPIVersions[0] != "ainfra.projectious.work/v1" ||
		len(contracts.ResultAPIVersions) != 1 || contracts.ResultAPIVersions[0] != "ainfra.result/v1" ||
		len(contracts.StandardOutputSchemaVersions) != 1 || contracts.StandardOutputSchemaVersions[0] != "1" {
		t.Errorf("supported contracts = %+v", contracts)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestMCPStdioDefaultRegistry(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	projectRoot := t.TempDir()
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: mcp-test
spec:
  template:
    source: local:../template
`
	if err := os.WriteFile(filepath.Join(projectRoot, "ainfra.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(projectRoot, ".ainfra"), 0o700); err != nil {
		t.Fatal(err)
	}
	process := exec.Command(binary, "mcp", "serve", "--stdio")
	process.Dir = projectRoot
	home := t.TempDir()
	runID := "20260816T120000Z-0123456789abcdef"
	runRoot := filepath.Join(home, ".local", "state", "ainfra", "runs", runID)
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	retained := map[string]string{
		"run.json":         `{"schemaVersion":1,"runId":"` + runID + `","operation":"apply","state":"succeeded","createdAt":"2026-08-16T12:00:00Z","planRecord":"plan-record.json"}`,
		"plan-record.json": `{"schemaVersion":1,"runId":"` + runID + `","intent":"apply","deployment":{"name":"mcp-test","digest":"sha256:x"},"template":{"source":"local:x","digest":"sha256:x"},"inputs":[],"engine":{"name":"opentofu","version":"1.10.0","executableDigest":"sha256:x"},"plan":{"path":"plan.tfplan","digest":"sha256:x","summaryPath":"plan.json"}}`,
		"events.jsonl": "{\"schemaVersion\":1,\"operation\":\"apply\",\"state\":\"started\",\"occurredAt\":\"2026-08-16T12:00:01Z\"}\n" +
			"{\"schemaVersion\":1,\"operation\":\"apply\",\"state\":\"inspection-required\",\"occurredAt\":\"2026-08-16T12:00:02Z\"}\n",
		"output.json":    `{"schema_version":"1","hosts":{"node":{"groups":["all"],"connection":{"type":"local"}}}}`,
		"inventory.yaml": "all:\n  hosts:\n    node:\n      ansible_connection: local\n",
	}
	for name, contents := range retained {
		if err := os.WriteFile(filepath.Join(runRoot, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	localRunRoot := filepath.Join(projectRoot, ".ainfra", "runs", runID)
	if err := os.MkdirAll(localRunRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"run.json", "events.jsonl"} {
		if err := os.WriteFile(filepath.Join(localRunRoot, name), []byte(retained[name]), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	process.Env = []string{"TERM=dumb", "HOME=" + home,
		"XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
	var stderr bytes.Buffer
	process.Stderr = &stderr
	client := mcp.NewClient(&mcp.Implementation{Name: "ainfra-blackbox", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: process}, nil)
	if err != nil {
		t.Fatalf("connect: %v, stderr: %s", err, stderr.String())
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
		!strings.Contains(catalog.Contents[0].Text, `"apiVersion":"ainfra.contracts/v1"`) {
		t.Fatalf("unexpected contract catalog: %+v", catalog)
	}
	schema, err := session.ReadResource(ctx,
		&mcp.ReadResourceParams{URI: "ainfra://schemas/v1/machine-output.schema.json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Contents) != 1 || schema.Contents[0].MIMEType != "application/schema+json" ||
		!strings.Contains(schema.Contents[0].Text, `"$schema"`) {
		t.Fatalf("unexpected schema resource: %+v", schema)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.version"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok || structured["apiVersion"] != "ainfra.result/v1" ||
		structured["tool"] != "ainfra.version" {
		t.Fatalf("unexpected structured result: %#v", result.StructuredContent)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.project.inspect"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	project, projectOK := structured["result"].(map[string]any)
	if !ok || !projectOK || project["name"] != "mcp-test" || project["root"] != projectRoot {
		t.Fatalf("unexpected project result: %#v", result.StructuredContent)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.status"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	status, statusOK := structured["result"].(map[string]any)
	statusProject, statusProjectOK := status["deployment"].(map[string]any)
	runs, runsOK := status["runs"].([]any)
	if result.IsError || !ok || !statusOK || !statusProjectOK || !runsOK ||
		statusProject["name"] != "mcp-test" || statusProject["root"] != projectRoot || len(runs) != 1 {
		t.Fatalf("unexpected status result: %#v", result.StructuredContent)
	}
	run, runOK := runs[0].(map[string]any)
	recovery, recoveryOK := run["recovery"].(map[string]any)
	if !runOK || !recoveryOK || run["executionOutcome"] != "interrupted" ||
		recovery["inspectionRequired"] != true || recovery["automaticRetryAllowed"] != false {
		t.Fatalf("unexpected recovery result: %#v", runs[0])
	}
	result, err = session.CallTool(ctx,
		&mcp.CallToolParams{Name: "ainfra.doctor.deployment"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	doctor, doctorOK := structured["result"].(map[string]any)
	diagnostics, diagnosticsOK := structured["diagnostics"].([]any)
	if !result.IsError || !ok || !doctorOK || structured["ok"] != false ||
		doctor["scope"] != "deployment" || !diagnosticsOK || len(diagnostics) != 1 {
		t.Fatalf("unexpected deployment doctor result: %#v", result)
	}
	finding, findingOK := diagnostics[0].(map[string]any)
	if !findingOK || finding["code"] != "AINFRA-E2310" || finding["status"] != "fail" {
		t.Fatalf("unexpected deployment diagnostic: %#v", diagnostics[0])
	}
	unexpected, unexpectedErr := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ainfra.doctor.deployment", Arguments: map[string]any{"reconcile": true},
	})
	if unexpectedErr == nil && !unexpected.IsError {
		t.Fatalf("doctor reconciliation input succeeded: %#v", unexpected)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.run"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	runDoctor, runDoctorOK := structured["result"].(map[string]any)
	if result.IsError || !ok || !runDoctorOK || structured["ok"] != true ||
		runDoctor["scope"] != "run" {
		t.Fatalf("unexpected run doctor result: %#v", result)
	}
	unexpected, unexpectedErr = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ainfra.doctor.run", Arguments: map[string]any{"runId": runID},
	})
	if unexpectedErr == nil && !unexpected.IsError {
		t.Fatalf("run doctor override input succeeded: %#v", unexpected)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.template"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	templateDoctor, templateDoctorOK := structured["result"].(map[string]any)
	summary, summaryOK := templateDoctor["summary"].(map[string]any)
	if result.IsError || !ok || !templateDoctorOK || !summaryOK || structured["ok"] != true ||
		templateDoctor["scope"] != "template" || summary["skip"] != float64(1) {
		t.Fatalf("unexpected template doctor result: %#v", result)
	}
	unexpected, unexpectedErr = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ainfra.doctor.template", Arguments: map[string]any{"path": "/tmp/other"},
	})
	if unexpectedErr == nil && !unexpected.IsError {
		t.Fatalf("template doctor path override succeeded: %#v", unexpected)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.doctor.environment"})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	environmentDoctor, environmentDoctorOK := structured["result"].(map[string]any)
	if !ok || !environmentDoctorOK || environmentDoctor["scope"] != "environment" {
		t.Fatalf("unexpected environment doctor result: %#v", result)
	}
	unexpected, unexpectedErr = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ainfra.doctor.environment", Arguments: map[string]any{"config": "/tmp/other"},
	})
	if unexpectedErr == nil && !unexpected.IsError {
		t.Fatalf("environment doctor config override succeeded: %#v", unexpected)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.output.read",
		Arguments: map[string]any{"runId": runID}})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	retainedOutput, retainedOutputOK := structured["result"].(map[string]any)
	standardOutput, standardOutputOK := retainedOutput["output"].(map[string]any)
	hosts, hostsOK := standardOutput["hosts"].(map[string]any)
	if result.IsError || !ok || !retainedOutputOK || !standardOutputOK || !hostsOK ||
		retainedOutput["runId"] != runID || hosts["node"] == nil {
		t.Fatalf("unexpected retained output result: %#v", result)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.inventory.read",
		Arguments: map[string]any{"runId": runID}})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok = result.StructuredContent.(map[string]any)
	retainedInventory, retainedInventoryOK := structured["result"].(map[string]any)
	if result.IsError || !ok || !retainedInventoryOK ||
		retainedInventory["mediaType"] != "application/yaml" ||
		!strings.Contains(retainedInventory["content"].(string), "ansible_connection: local") {
		t.Fatalf("unexpected retained inventory result: %#v", result)
	}
	unexpected, unexpectedErr = session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ainfra.output.read", Arguments: map[string]any{"runId": runID, "path": "/tmp/other"},
	})
	if unexpectedErr == nil && !unexpected.IsError {
		t.Fatalf("retained output path override succeeded: %#v", unexpected)
	}
	if _, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.apply"}); err == nil {
		t.Fatal("undisclosed mutation tool call succeeded")
	}
	concurrent := make(chan error, 32)
	for range 32 {
		go func() {
			response, callErr := session.CallTool(ctx, &mcp.CallToolParams{Name: "ainfra.version"})
			if callErr == nil && response.IsError {
				callErr = errors.New("version returned an error result")
			}
			concurrent <- callErr
		}()
	}
	for range 32 {
		if err := <-concurrent; err != nil {
			t.Fatalf("concurrent MCP request: %v", err)
		}
	}
}

func TestMCPStdioRejectsMalformedAndOversizedFrames(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name    string
		payload string
	}{
		{name: "malformed", payload: "{not-json}\n"},
		{name: "oversized", payload: `{"jsonrpc":"2.0","id":1,"method":"` +
			strings.Repeat("frame-canary-", (4<<20)/len("frame-canary-")+1) + `"}` + "\n"},
	} {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			projectRoot := t.TempDir()
			manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: malformed-frame-test
spec:
  template:
    source: local:../template
`
			if err := os.WriteFile(filepath.Join(projectRoot, "ainfra.yaml"),
				[]byte(manifest), 0o600); err != nil {
				t.Fatal(err)
			}
			process := exec.CommandContext(ctx, binary, "mcp", "serve", "--stdio")
			process.Dir = projectRoot
			process.Env = []string{"TERM=dumb", "HOME=" + t.TempDir(),
				"XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
			process.Stdin = strings.NewReader(fixture.payload)
			var stdout, stderr bytes.Buffer
			process.Stdout, process.Stderr = &stdout, &stderr
			if err := process.Run(); err == nil {
				t.Fatal("invalid MCP frame succeeded")
			}
			if stdout.Len() != 0 {
				t.Fatalf("invalid MCP frame polluted stdout: %q", stdout.String())
			}
			if !strings.Contains(stderr.String(), "AINFRA-E5001") ||
				strings.Contains(stderr.String(), "frame-canary-") {
				t.Fatalf("unexpected sanitized stderr: %q", stderr.String())
			}
		})
	}
}

func TestMCPStdioRejectsMissingProjectBeforeProtocolOutput(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, binary, "mcp", "serve", "--stdio")
	process.Dir = t.TempDir()
	process.Env = []string{"TERM=dumb", "HOME=" + t.TempDir(),
		"XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	if err := process.Run(); err == nil {
		t.Fatal("MCP server without a project succeeded")
	}
	if stdout.Len() != 0 {
		t.Fatalf("startup diagnostics contaminated protocol stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "AINFRA-E5001") {
		t.Fatalf("missing startup diagnostic: %q", stderr.String())
	}
}

func TestOperationalLoggingFlagsAndCredentialRedaction(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	logPath := filepath.Join(root, "ainfra.log")
	environment := []string{"TERM=dumb", "HOME=" + t.TempDir(),
		"XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
	run := func(arguments ...string) (int, string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		invocation := exec.CommandContext(ctx, binary, arguments...)
		invocation.Dir, invocation.Env = root, environment
		var stdout, stderr bytes.Buffer
		invocation.Stdout, invocation.Stderr = &stdout, &stderr
		err := invocation.Run()
		if err == nil {
			return 0, stderr.String()
		}
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		return exitError.ExitCode(), stderr.String()
	}
	loggingFlags := []string{"--log-file", logPath, "--log-format", "json", "--format=json"}
	if code, stderr := run(append([]string{"version", "-v"}, loggingFlags...)...); code != 0 || stderr != "" {
		t.Fatalf("version exit=%d stderr=%q", code, stderr)
	}
	credential := "Bearer abcdefghijklmnop"
	if code, stderr := run(append([]string{credential}, loggingFlags...)...); code != 2 || stderr != "" {
		t.Fatalf("invalid invocation exit=%d stderr=%q", code, stderr)
	}
	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(contents, []byte(`"message":"command started"`)) ||
		!bytes.Contains(contents, []byte(`"message":"command failed"`)) ||
		!bytes.Contains(contents, []byte("<redacted>")) ||
		bytes.Contains(contents, []byte("abcdefghijklmnop")) {
		t.Fatalf("unexpected operational log: %s", contents)
	}
	if info, err := os.Stat(logPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("operational log permissions=%v err=%v", info, err)
	}
}

func TestTemplateLockLocalJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	templateRoot := filepath.Join(root, "template-example")
	if err := os.CopyFS(templateRoot, os.DirFS("../../spec/examples/v1/template-example")); err != nil {
		t.Fatal(err)
	}
	deployment := filepath.Join(root, "deployment")
	if err := os.Mkdir(deployment, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: dev
spec:
  template:
    source: local:../template-example
`
	if err := os.WriteFile(filepath.Join(deployment, "ainfra.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	home, configHome, cacheHome := t.TempDir(), t.TempDir(), t.TempDir()
	environment := []string{
		"TERM=dumb", "HOME=" + home,
		"XDG_CONFIG_HOME=" + configHome, "XDG_CACHE_HOME=" + cacheHome,
	}
	command := exec.CommandContext(
		ctx, binary, "template", "lock", deployment, "--format=json",
	)
	command.Dir = root
	command.Env = environment
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("run: %v, stderr: %s", err, stderr.String())
	}
	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Result  struct {
			Source        string `json:"source"`
			ContentDigest string `json:"contentDigest"`
			Changed       bool   `json:"changed"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Command != "template.lock" || !envelope.OK ||
		envelope.Result.Source != "local:../template-example" ||
		!strings.HasPrefix(envelope.Result.ContentDigest, "sha256:") ||
		!envelope.Result.Changed {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	if _, err := os.Stat(filepath.Join(deployment, "ainfra.lock")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateRoot, "update.txt"), []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	updateContext, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer updateCancel()
	update := exec.CommandContext(
		updateContext, binary, "template", "update", deployment, "--format=json",
	)
	update.Dir, update.Env = root, environment
	updated, err := update.Output()
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !bytes.Contains(updated, []byte(`"command":"template.update"`)) ||
		!bytes.Contains(updated, []byte(`"changed":true`)) {
		t.Fatalf("unexpected update output: %s", updated)
	}
}

func TestReviewedPlanApplyAndStalePlanRefusal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	templateRoot := filepath.Join(root, "template")
	if err := os.CopyFS(templateRoot, os.DirFS("../../spec/examples/v1/template-example")); err != nil {
		t.Fatal(err)
	}
	deployment := filepath.Join(root, "deployment")
	if err := os.Mkdir(deployment, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: reviewed-apply
spec:
  template:
    source: local:../template
`
	if err := os.WriteFile(filepath.Join(deployment, "ainfra.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "applied-arguments")
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
  apply) sleep 1; printf '%s\n' "$@" > ` + shellLiteral(marker) + ` ;;
  output) printf '%s\n' '{"ainfra_inventory":{"sensitive":false,"value":{"schema_version":"1","hosts":{"localhost":{"groups":["local"],"connection":{"type":"local"}}}}}}' ;;
  *) exit 91 ;;
esac
`
	if err := os.WriteFile(tofuPath, []byte(fakeTofu), 0o700); err != nil {
		t.Fatal(err)
	}
	runnerPath := filepath.Join(root, "ansible-runner")
	fakeRunner := `#!/bin/sh
if [ "$1" = "--version" ]; then printf '%s\n' 'ansible-runner 2.4.1'; exit 0; fi
artifact=''
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--artifact-dir" ]; then artifact="$2"; shift 2; continue; fi
  shift
done
mkdir -p "$artifact/job_events"
printf '%s\n' '{"event":"playbook_on_stats","event_data":{"changed":{},"dark":{},"failures":{},"ok":{"localhost":2},"processed":{"localhost":1},"skipped":{}}}' > "$artifact/job_events/stats.json"
`
	if err := os.WriteFile(runnerPath, []byte(fakeRunner), 0o700); err != nil {
		t.Fatal(err)
	}
	cacheRoot, runsRoot := filepath.Join(root, "cache"), filepath.Join(root, "runs")
	configPath := filepath.Join(root, "config.yaml")
	configuration := `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
paths:
  cache: ` + cacheRoot + `
  runs: ` + runsRoot + `
executables:
  tofu: ` + tofuPath + `
  ansibleRunner: ` + runnerPath + "\n"
	if err := os.WriteFile(configPath, []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	environment := []string{"HOME=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, "cache"), "PATH=" + os.Getenv("PATH")}
	run := func(arguments ...string) ([]byte, int, string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, arguments...)
		command.Dir, command.Env = root, environment
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if err == nil {
			return stdout.Bytes(), 0, stderr.String()
		}
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %v: %v", arguments, err)
		}
		return stdout.Bytes(), exitError.ExitCode(), stderr.String()
	}
	if stdout, code, stderr := run("template", "lock", deployment, "--config", configPath,
		"--format=json"); code != 0 {
		t.Fatalf("template lock exit=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	planOutput, code, stderr := run("plan", deployment, "--config", configPath, "--format=json")
	if code != 0 {
		t.Fatalf("plan exit=%d stderr=%s", code, stderr)
	}
	var planned struct {
		Result struct {
			RunID string `json:"runId"`
		} `json:"result"`
	}
	if err := json.Unmarshal(planOutput, &planned); err != nil || planned.Result.RunID == "" {
		t.Fatalf("plan result=%s err=%v", planOutput, err)
	}
	type concurrentResult struct {
		output []byte
		code   int
		err    error
	}
	results := make(chan concurrentResult, 2)
	for range 2 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "apply", deployment, "--plan",
				planned.Result.RunID, "--config", configPath, "--format=json")
			command.Dir, command.Env = root, environment
			var stdout bytes.Buffer
			command.Stdout = &stdout
			err := command.Run()
			if err == nil {
				results <- concurrentResult{output: stdout.Bytes()}
				return
			}
			exitError, ok := err.(*exec.ExitError)
			if !ok {
				results <- concurrentResult{err: err}
				return
			}
			results <- concurrentResult{output: stdout.Bytes(), code: exitError.ExitCode()}
		}()
	}
	successes, stale := 0, 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		switch result.code {
		case 0:
			if !bytes.Contains(result.output, []byte(`"executionOutcome":"succeeded"`)) {
				t.Fatalf("unexpected apply output: %s", result.output)
			}
			successes++
		case 5:
			stale++
		default:
			t.Fatalf("concurrent apply exit=%d output=%s", result.code, result.output)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("concurrent results: success=%d stale=%d", successes, stale)
	}
	arguments, err := os.ReadFile(marker)
	if err != nil || !bytes.Contains(arguments, []byte("../../plan.tfplan")) {
		t.Fatalf("apply arguments=%q err=%v", arguments, err)
	}
	events, err := os.ReadFile(filepath.Join(runsRoot, planned.Result.RunID, "events.jsonl"))
	if err != nil || !bytes.Contains(events, []byte(`"state":"started"`)) ||
		!bytes.Contains(events, []byte(`"state":"succeeded"`)) {
		t.Fatalf("events=%q err=%v", events, err)
	}
	if _, code, _ := run("apply", deployment, "--plan", planned.Result.RunID,
		"--config", configPath, "--format=json"); code != 5 {
		t.Fatalf("replayed apply exit=%d, want 5", code)
	}
	outputResult, code, stderr := run("output", deployment, "--run", planned.Result.RunID,
		"--config", configPath, "--format=json")
	if code != 0 || !bytes.Contains(outputResult, []byte(`"command":"output"`)) {
		t.Fatalf("output exit=%d stdout=%s stderr=%s", code, outputResult, stderr)
	}
	inventoryResult, code, stderr := run("inventory", deployment, "--run", planned.Result.RunID,
		"--config", configPath, "--format=json")
	if code != 0 || !bytes.Contains(inventoryResult, []byte(`"command":"inventory"`)) {
		t.Fatalf("inventory exit=%d stdout=%s stderr=%s", code, inventoryResult, stderr)
	}
	inventoryContents, err := os.ReadFile(filepath.Join(runsRoot, planned.Result.RunID, "inventory.yaml"))
	if err != nil || string(inventoryContents) != "local:\n  hosts:\n    localhost:\n      ansible_connection: local\n" {
		t.Fatalf("inventory=%q err=%v", inventoryContents, err)
	}
	configureResult, code, stderr := run("configure", deployment, "--run", planned.Result.RunID,
		"--config", configPath, "--format=json")
	if code != 0 || !bytes.Contains(configureResult, []byte(`"operation":"configure"`)) {
		t.Fatalf("configure exit=%d stdout=%s stderr=%s", code, configureResult, stderr)
	}
	checkResult, code, stderr := run("configure", deployment, "--run", planned.Result.RunID,
		"--check", "--config", configPath, "--format=json")
	if code != 0 || !bytes.Contains(checkResult, []byte(`"operation":"configure-check"`)) {
		t.Fatalf("configure check exit=%d stdout=%s stderr=%s", code, checkResult, stderr)
	}
	if _, code, _ := run("configure", deployment, "--run", planned.Result.RunID,
		"--config", configPath, "--format=json"); code != 1 {
		t.Fatalf("replayed configure exit=%d, want 1", code)
	}
	secondOutput, code, stderr := run("plan", deployment, "--config", configPath, "--format=json")
	if code != 0 || json.Unmarshal(secondOutput, &planned) != nil {
		t.Fatalf("second plan exit=%d stdout=%s stderr=%s", code, secondOutput, stderr)
	}
	deployResult, code, stderr := run("deploy", deployment, "--plan", planned.Result.RunID,
		"--config", configPath, "--format=json")
	if code != 0 || !bytes.Contains(deployResult, []byte(`"operation":"deploy"`)) ||
		!bytes.Contains(deployResult, []byte(`"stages":[{"operation":"output","applicability":"applicable","status":"succeeded"}`)) ||
		!bytes.Contains(deployResult, []byte(`ansible-runner/configure/artifacts`)) ||
		!bytes.Contains(deployResult, []byte(`ansible-runner/configure-check/artifacts`)) {
		t.Fatalf("deploy exit=%d stdout=%s stderr=%s", code, deployResult, stderr)
	}
	thirdOutput, code, stderr := run("plan", deployment, "--config", configPath, "--format=json")
	if code != 0 || json.Unmarshal(thirdOutput, &planned) != nil {
		t.Fatalf("third plan exit=%d stdout=%s stderr=%s", code, thirdOutput, stderr)
	}
	planPath := filepath.Join(runsRoot, planned.Result.RunID, "plan.tfplan")
	if err := os.WriteFile(planPath, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, code, _ := run("apply", deployment, "--plan", planned.Result.RunID,
		"--config", configPath, "--format=json"); code != 5 {
		t.Fatalf("tampered apply exit=%d, want 5", code)
	}
}

func shellLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestUnknownCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "unknown")
	command.Dir = t.TempDir()
	command.Env = []string{"HOME=" + t.TempDir(), "XDG_CONFIG_HOME=" + t.TempDir(), "XDG_CACHE_HOME=" + t.TempDir()}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() != 2 {
		t.Fatalf("error = %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() == 0 {
		t.Errorf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestDoctorEnvironmentJSONIsOneCleanResult(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	home := t.TempDir()
	command := exec.CommandContext(
		ctx, binary, "doctor", "environment", "--format=json",
	)
	command.Dir = t.TempDir()
	command.Env = []string{
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, "cache"),
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok || exitError.ExitCode() != 3 {
			t.Fatalf("run: %v, stderr: %s", err, stderr.String())
		}
	}
	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Result  struct {
			Scope                  string `json:"scope"`
			EffectiveConfiguration struct {
				Values map[string]any `json:"values"`
			} `json:"effectiveConfiguration"`
		} `json:"result"`
	}
	decoder := json.NewDecoder(&stdout)
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if decoder.More() {
		t.Fatal("doctor wrote more than one JSON value")
	}
	if envelope.Command != "doctor.environment" ||
		envelope.Result.Scope != "environment" ||
		len(envelope.Result.EffectiveConfiguration.Values) != 12 || stderr.Len() != 0 {
		t.Fatalf("unexpected result: %+v, stderr=%q", envelope, stderr.String())
	}
}

func TestDoctorDeploymentJSONAndContractFailure(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ainfra.yaml"), []byte(`apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: blackbox
spec:
  template:
    source: local:../template
`), 0o600); err != nil {
		t.Fatal(err)
	}
	run := func(target string) (int, []byte, string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		invocation := exec.CommandContext(
			ctx, binary, "doctor", "deployment", target, "--format=json",
		)
		invocation.Dir = t.TempDir()
		invocation.Env = []string{
			"HOME=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
		}
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		invocation.Stdout = &stdout
		invocation.Stderr = &stderr
		err := invocation.Run()
		if err == nil {
			return 0, stdout.Bytes(), stderr.String()
		}
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run: %v", err)
		}
		return exitError.ExitCode(), stdout.Bytes(), stderr.String()
	}
	code, stdout, stderr := run(root)
	if code != 0 || stderr != "" {
		t.Fatalf("valid deployment exit=%d stderr=%q", code, stderr)
	}
	var success struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Result  struct {
			Scope    string `json:"scope"`
			Findings []any  `json:"findings"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout, &success); err != nil {
		t.Fatal(err)
	}
	if success.Command != "doctor.deployment" || !success.OK ||
		success.Result.Scope != "deployment" || len(success.Result.Findings) != 9 {
		t.Fatalf("success=%+v", success)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	repair := exec.CommandContext(
		ctx, binary, "doctor", "deployment", root, "--reconcile",
		"--non-interactive", "--yes", "--format=json",
	)
	repair.Dir = t.TempDir()
	repair.Env = []string{
		"HOME=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
	}
	var repairOutput bytes.Buffer
	var repairEvidence bytes.Buffer
	repair.Stdout, repair.Stderr = &repairOutput, &repairEvidence
	if err := repair.Run(); err != nil {
		t.Fatalf("reconcile: %v stderr=%q", err, repairEvidence.String())
	}
	var repaired struct {
		OK     bool `json:"ok"`
		Result struct {
			Findings []struct {
				Check          string `json:"check"`
				Reconciliation string `json:"reconciliation"`
			} `json:"findings"`
		} `json:"result"`
	}
	if err := json.Unmarshal(repairOutput.Bytes(), &repaired); err != nil {
		t.Fatal(err)
	}
	applied := false
	for _, finding := range repaired.Result.Findings {
		if finding.Check == "deployment.runtime-permissions" &&
			finding.Reconciliation == "applied" {
			applied = true
		}
	}
	information, statErr := os.Stat(filepath.Join(root, ".ainfra"))
	if statErr != nil {
		t.Fatalf("inspect repaired runtime directory: %v", statErr)
	}
	if !repaired.OK || !applied || information.Mode().Perm() != 0o700 ||
		repairEvidence.Len() == 0 {
		t.Fatalf(
			"repaired=%+v mode=%v statErr=%v stderr=%q",
			repaired, information.Mode(), statErr, repairEvidence.String(),
		)
	}

	code, stdout, stderr = run(filepath.Join(t.TempDir(), "missing"))
	if code != 2 || stderr != "" {
		t.Fatalf("invalid deployment exit=%d stderr=%q", code, stderr)
	}
	var failure struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
	}
	if err := json.Unmarshal(stdout, &failure); err != nil {
		t.Fatal(err)
	}
	if failure.Command != "doctor.deployment" || failure.OK {
		t.Fatalf("failure=%+v", failure)
	}
}

func TestDoctorTemplateJSON(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../spec/examples/v1/template-example")
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	invocation := exec.CommandContext(
		ctx, binary, "doctor", "template", root, "--format=json",
	)
	invocation.Dir = t.TempDir()
	invocation.Env = []string{
		"HOME=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, "config"),
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	invocation.Stdout = &stdout
	invocation.Stderr = &stderr
	if err := invocation.Run(); err != nil {
		t.Fatalf("run: %v stderr=%q", err, stderr.String())
	}
	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Result  struct {
			Scope    string `json:"scope"`
			Findings []any  `json:"findings"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Command != "doctor.template" || !envelope.OK ||
		envelope.Result.Scope != "template" || len(envelope.Result.Findings) != 4 ||
		stderr.Len() != 0 {
		t.Fatalf("envelope=%+v stderr=%q", envelope, stderr.String())
	}
}

func TestHelpAndInvalidInvocationJSON(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		wantCode  int
		command   string
	}{
		{name: "root help", arguments: []string{"--format=json", "help"}, command: "help"},
		{name: "root flag help", arguments: []string{"--help", "--format=json"}, command: "help"},
		{name: "group help", arguments: []string{"help", "doctor", "--format=json"}, command: "help"},
		{name: "group flag help", arguments: []string{"doctor", "--help", "--format=json"}, command: "help"},
		{name: "leaf help", arguments: []string{"version", "--help", "--format=json"}, command: "help"},
		{name: "leaf command help", arguments: []string{"help", "version", "--format=json"}, command: "help"},
		{name: "unknown", arguments: []string{"unknown", "--format=json"}, wantCode: 2, command: "invocation"},
		{name: "unknown help topic", arguments: []string{"help", "missing", "--format=json"}, wantCode: 2, command: "invocation"},
		{name: "unknown option", arguments: []string{"version", "--bogus", "--format=json"}, wantCode: 2, command: "invocation"},
		{name: "malformed global", arguments: []string{"version", "--format", "--format=json"}, wantCode: 2, command: "invocation"},
		{name: "extra argument", arguments: []string{"version", "extra", "--format=json"}, wantCode: 2, command: "invocation"},
		{name: "delimiter", arguments: []string{"--format=json", "version", "--"}, command: "version"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			invocation := exec.CommandContext(ctx, binary, test.arguments...)
			working := t.TempDir()
			state := t.TempDir()
			invocation.Dir = working
			invocation.Env = []string{"HOME=" + state, "XDG_CONFIG_HOME=" + state, "XDG_CACHE_HOME=" + state}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			invocation.Stdout = &stdout
			invocation.Stderr = &stderr
			err := invocation.Run()
			if test.wantCode == 0 && err != nil {
				t.Fatalf("run: %v", err)
			}
			if test.wantCode != 0 {
				exitError, ok := err.(*exec.ExitError)
				if !ok || exitError.ExitCode() != test.wantCode {
					t.Fatalf("error=%v, want exit %d", err, test.wantCode)
				}
			}
			var envelope struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
				t.Fatalf("decode stdout: %v", err)
			}
			if envelope.Command != test.command || stderr.Len() != 0 {
				t.Errorf("command=%q stderr=%q", envelope.Command, stderr.String())
			}
		})
	}
}
