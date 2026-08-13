package command_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/command"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/initialize"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/reconcile"
)

func run(arguments ...string) (command.ExitCode, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := command.Run(arguments, command.Options{
		Build: app.Build{
			Version: "1.0.0-alpha.1",
			Commit:  "0123456789abcdef",
			BuiltAt: "2026-08-07T00:00:00Z",
		},
		IO: command.IO{Stdout: &stdout, Stderr: &stderr},
	})
	return code, stdout.String(), stderr.String()
}

func runWithDoctor(arguments ...string) (command.ExitCode, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := command.Run(arguments, command.Options{
		IO: command.IO{Stdout: &stdout, Stderr: &stderr},
		DoctorEnvironment: func(
			request app.DoctorEnvironmentRequest,
		) (app.DoctorEnvironmentResponse, error) {
			format := "text"
			if request.Format != nil {
				format = *request.Format
			}
			return app.DoctorEnvironmentResponse{
				Format: format, OutputStyle: "auto", Color: "auto",
				Result: output.Doctor{
					Scope: "environment", Summary: output.DoctorSummary{Pass: 1},
					Findings: []diagnostic.Diagnostic{},
					EffectiveConfiguration: &output.EffectiveConfiguration{
						Values:                  map[string]output.EffectiveConfigurationValue{},
						Files:                   []output.ConfigurationFile{},
						RejectedProjectSettings: []output.RejectedProjectSetting{},
					},
				},
			}, nil
		},
	})
	return code, stdout.String(), stderr.String()
}

func TestVersionJSON(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := run("version", "--format", "json")
	if code != command.ExitSuccess {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(stdout), &value); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if stderr != "" {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestInitDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var target string
	code := command.Run(
		[]string{"init", "/tmp/example", "--format=json"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			Initialize: func(path string) (initialize.Result, error) {
				target = path
				return initialize.Result{
					Name: "example", Root: path,
					CreatedPaths: []string{path + "/ainfra.yaml"},
				}, nil
			},
		},
	)
	if code != command.ExitSuccess || target != "/tmp/example" ||
		!strings.Contains(stdout.String(), `"command":"init"`) {
		t.Fatalf("exit=%d target=%q stdout=%q", code, target, stdout.String())
	}
}

func TestTemplateLockDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var request app.TemplateLockRequest
	code := command.Run(
		[]string{"template", "lock", "/tmp/deployment", "--format=json"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			TemplateLock: func(value app.TemplateLockRequest) (output.Template, error) {
				request = value
				return output.Template{
					Source: "local:../template", ResolvedRevision: "local:../template",
					ContentDigest: "sha256:" + strings.Repeat("a", 64), Changed: true,
				}, nil
			},
		},
	)
	if code != command.ExitSuccess || request.Target != "/tmp/deployment" ||
		!strings.Contains(stdout.String(), `"command":"template.lock"`) {
		t.Fatalf("exit=%d request=%#v stdout=%q", code, request, stdout.String())
	}
}

func TestTemplateUpdateDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	called := false
	code := command.Run(
		[]string{"template", "update", "--format=json"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			TemplateUpdate: func(app.TemplateLockRequest) (output.Template, error) {
				called = true
				return output.Template{
					Source:           "git::https://example.com/templates.git",
					ResolvedRevision: strings.Repeat("a", 40),
					ContentDigest:    "sha256:" + strings.Repeat("b", 64), Changed: true,
				}, nil
			},
		},
	)
	if code != command.ExitSuccess || !called ||
		!strings.Contains(stdout.String(), `"command":"template.update"`) {
		t.Fatalf("exit=%d called=%t stdout=%q", code, called, stdout.String())
	}
}

func TestDoctorEnvironmentDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := runWithDoctor(
		"doctor", "environment", "--config=/tmp/config.yaml",
		"--project", "/tmp/deployment", "--format=json",
	)
	if code != command.ExitSuccess || stderr != "" {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			Scope string `json:"scope"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Command != "doctor.environment" || envelope.Result.Scope != "environment" {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
}

func TestDoctorDeploymentDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var received app.DoctorDeploymentRequest
	code := command.Run(
		[]string{"doctor", "deployment", "/deployment", "--format=json"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &stderr},
			DoctorDeployment: func(
				request app.DoctorDeploymentRequest,
			) (app.DoctorEnvironmentResponse, error) {
				received = request
				return app.DoctorEnvironmentResponse{
					Format: "json", OutputStyle: "auto", Color: "auto",
					Result: output.Doctor{
						Scope: "deployment", Summary: output.DoctorSummary{Pass: 1},
						Findings: []diagnostic.Diagnostic{},
					},
				}, nil
			},
		},
	)
	if code != command.ExitSuccess || stderr.Len() != 0 || received.Target != "/deployment" {
		t.Fatalf("exit=%d request=%#v stderr=%q", code, received, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"command":"doctor.deployment"`) {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestDoctorTemplateDispatchesCanonicalResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var received app.DoctorTemplateRequest
	code := command.Run(
		[]string{"doctor", "template", "/template", "--format=json"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			DoctorTemplate: func(
				request app.DoctorTemplateRequest,
			) (app.DoctorEnvironmentResponse, error) {
				received = request
				return app.DoctorEnvironmentResponse{
					Format: "json", OutputStyle: "auto", Color: "auto",
					Result: output.Doctor{
						Scope: "template", Summary: output.DoctorSummary{Pass: 1},
						Findings: []diagnostic.Diagnostic{},
					},
				}, nil
			},
		},
	)
	if code != command.ExitSuccess || received.Target != "/template" ||
		!strings.Contains(stdout.String(), `"command":"doctor.template"`) {
		t.Fatalf("exit=%d request=%#v stdout=%q", code, received, stdout.String())
	}
}

func TestBareDoctorIsExactAliasForDoctorAll(t *testing.T) {
	t.Parallel()
	invoke := func(arguments ...string) (command.ExitCode, string) {
		var stdout bytes.Buffer
		code := command.Run(arguments, command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			DoctorAll: func(
				request app.DoctorAllRequest,
			) (app.DoctorEnvironmentResponse, error) {
				return app.DoctorEnvironmentResponse{
					Format: "json", OutputStyle: "auto", Color: "auto",
					Result: output.Doctor{
						Scope: "all", Summary: output.DoctorSummary{Skip: 1},
						Findings: []diagnostic.Diagnostic{{
							Code: "AINFRA-E2601", Severity: diagnostic.SeverityInfo,
							Check: "template.resolved-source", Scope: "template",
							Status: "skip", Component: request.Target,
							Reconciliation: "not_available",
						}},
					},
				}, nil
			},
		})
		return code, stdout.String()
	}
	bareCode, bare := invoke("doctor", "/deployment", "--format=json")
	allCode, all := invoke("doctor", "all", "/deployment", "--format=json")
	if bareCode != command.ExitSuccess || allCode != command.ExitSuccess || bare != all {
		t.Fatalf("bare=(%d,%q) all=(%d,%q)", bareCode, bare, allCode, all)
	}
}

func TestDoctorReconciliationRequiresPairedAutomationApproval(t *testing.T) {
	t.Parallel()
	invoke := func(arguments ...string) (command.ExitCode, int) {
		calls := 0
		code := command.Run(arguments, command.Options{
			IO: command.IO{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
			DoctorDeployment: func(
				request app.DoctorDeploymentRequest,
			) (app.DoctorEnvironmentResponse, error) {
				calls++
				response := app.DoctorEnvironmentResponse{
					Format: "text", OutputStyle: "plain", Color: "never",
					Result: output.Doctor{
						Scope: "deployment", Summary: output.DoctorSummary{Warning: 1},
						Findings: []diagnostic.Diagnostic{},
					},
				}
				if request.ApplyReconciliation {
					response.Result.Summary = output.DoctorSummary{Pass: 1}
					return response, nil
				}
				response.ReconciliationPlan = []reconcile.Action{{
					CheckID: "deployment.runtime-permissions", Path: "/deployment/.ainfra",
					Kind: reconcile.ActionCreateRuntimeDirectory, Mode: 0o700,
					RollbackLimitation: "remove manually",
				}}
				return response, nil
			},
		})
		return code, calls
	}
	denied, deniedCalls := invoke(
		"doctor", "deployment", "/deployment", "--reconcile", "--non-interactive",
	)
	if denied != command.ExitInvalidInput || deniedCalls != 1 {
		t.Fatalf("denied exit=%d calls=%d", denied, deniedCalls)
	}
	approved, approvedCalls := invoke(
		"doctor", "deployment", "/deployment", "--reconcile",
		"--non-interactive", "--yes",
	)
	if approved != command.ExitSuccess || approvedCalls != 2 {
		t.Fatalf("approved exit=%d calls=%d", approved, approvedCalls)
	}
}

func TestStaticHelpDoesNotConstructDoctor(t *testing.T) {
	t.Parallel()
	called := false
	var stdout bytes.Buffer
	code := command.Run(
		[]string{"doctor", "environment", "--help"},
		command.Options{
			IO: command.IO{Stdout: &stdout, Stderr: &bytes.Buffer{}},
			DoctorEnvironment: func(
				app.DoctorEnvironmentRequest,
			) (app.DoctorEnvironmentResponse, error) {
				called = true
				return app.DoctorEnvironmentResponse{}, nil
			},
		},
	)
	if code != command.ExitSuccess || called {
		t.Fatalf("exit=%d called=%v", code, called)
	}
}

func TestVersionAcceptsGlobalOptionsBeforeCommand(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := run("--format", "json", "version")
	if code != command.ExitSuccess {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if envelope["command"] != "version" {
		t.Errorf("command = %v", envelope["command"])
	}
}

func TestInvalidInvocation(t *testing.T) {
	t.Parallel()
	tests := [][]string{
		{"unknown"},
		{"version", "unexpected"},
		{"version", "--format", "xml"},
	}
	for _, arguments := range tests {
		code, stdout, stderr := run(arguments...)
		if code != command.ExitInvalidInput {
			t.Errorf("Run(%q) exit = %d", arguments, code)
		}
		if stdout != "" {
			t.Errorf("Run(%q) stdout = %q", arguments, stdout)
		}
		if stderr == "" {
			t.Errorf("Run(%q) missing diagnostic", arguments)
		}
	}
}

func TestInvalidJSONInvocation(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := run("version", "--format", "json", "unexpected")
	if code != command.ExitInvalidInput {
		t.Fatalf("exit = %d", code)
	}
	if stderr != "" {
		t.Errorf("stderr = %q", stderr)
	}
	var envelope struct {
		Command     string `json:"command"`
		OK          bool   `json:"ok"`
		Result      any    `json:"result"`
		Diagnostics []any  `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if envelope.Command != "invocation" || envelope.OK || envelope.Result != nil || len(envelope.Diagnostics) != 1 {
		t.Errorf("unexpected failure envelope: %+v", envelope)
	}
}

func TestHelpJSONNormalizesTopics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		arguments []string
		topic     string
	}{
		{[]string{"--format=json", "help"}, "ainfra"},
		{[]string{"help", "doctor", "--format", "json"}, "doctor"},
		{[]string{"doctor", "environment", "--help", "--format=json"}, "doctor.environment"},
		{[]string{"version", "--help", "--format=json"}, "version"},
	}
	for _, test := range tests {
		code, stdout, stderr := run(test.arguments...)
		if code != command.ExitSuccess || stderr != "" {
			t.Errorf("Run(%q) exit=%d stderr=%q", test.arguments, code, stderr)
			continue
		}
		var envelope struct {
			Command string `json:"command"`
			Result  struct {
				Topic       string `json:"topic"`
				Usage       string `json:"usage"`
				Summary     string `json:"summary"`
				Subcommands []any  `json:"subcommands"`
				Arguments   []any  `json:"arguments"`
				Options     []any  `json:"options"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
			t.Errorf("Run(%q) decode: %v", test.arguments, err)
			continue
		}
		if envelope.Command != "help" || envelope.Result.Topic != test.topic ||
			envelope.Result.Usage == "" || envelope.Result.Summary == "" ||
			envelope.Result.Subcommands == nil || envelope.Result.Arguments == nil ||
			envelope.Result.Options == nil {
			t.Errorf("Run(%q) envelope=%+v", test.arguments, envelope)
		}
	}
}

func TestInvocationOutputSelection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		arguments []string
		json      bool
	}{
		{[]string{"--format", "json", "unknown"}, true},
		{[]string{"unknown", "--format=json"}, true},
		{[]string{"help", "missing", "--format=json"}, true},
		{[]string{"--format", "xml", "unknown"}, false},
		{[]string{"--", "unknown", "--format=json"}, false},
	}
	for _, test := range tests {
		code, stdout, stderr := run(test.arguments...)
		if code != command.ExitInvalidInput {
			t.Errorf("Run(%q) exit=%d", test.arguments, code)
		}
		if test.json {
			if stdout == "" || stderr != "" {
				t.Errorf("Run(%q) stdout=%q stderr=%q", test.arguments, stdout, stderr)
			}
		} else if stdout != "" || stderr == "" {
			t.Errorf("Run(%q) stdout=%q stderr=%q", test.arguments, stdout, stderr)
		}
	}
}

func TestPlanHelpMatchesCanonicalMetadata(t *testing.T) {
	code, stdout, stderr := run("help", "plan", "--format=json")
	if code != command.ExitSuccess || stderr != "" {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	var envelope struct {
		Result struct {
			Usage     string                     `json:"usage"`
			Arguments []struct{ Name string }    `json:"arguments"`
			Options   []struct{ Names []string } `json:"options"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result.Usage != "ainfra plan [DEPLOYMENT] [--destroy] [options]" ||
		len(envelope.Result.Arguments) != 1 || envelope.Result.Arguments[0].Name != "DEPLOYMENT" ||
		len(envelope.Result.Options) == 0 || envelope.Result.Options[0].Names[0] != "--destroy" {
		t.Errorf("non-canonical plan help: %+v", envelope.Result)
	}
}

func TestPlanHelpGolden(t *testing.T) {
	code, stdout, stderr := run("help", "plan", "--output-style=plain", "--color=never")
	if code != command.ExitSuccess || stderr != "" {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	want, err := os.ReadFile("testdata/help-plan.golden")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != string(want) {
		t.Errorf("help output:\n%s\nwant:\n%s", stdout, want)
	}
}

func TestDelimiterIsNotPositional(t *testing.T) {
	code, _, stderr := run("version", "--")
	if code != command.ExitSuccess || stderr != "" {
		t.Errorf("version --: exit=%d stderr=%q", code, stderr)
	}
	code, stdout, stderr := run("--format=json", "help", "--", "version")
	if code != command.ExitSuccess || stderr != "" || !strings.Contains(stdout, `"topic":"version"`) {
		t.Errorf("help -- version: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestExitCodeContract(t *testing.T) {
	t.Parallel()
	if command.ExitSuccess != 0 || command.ExitOperationFailed != 1 ||
		command.ExitInvalidInput != 2 || command.ExitDependency != 3 ||
		command.ExitSecurity != 4 || command.ExitStaleBinding != 5 ||
		command.ExitInterrupted != 6 {
		t.Fatal("exit-code contract changed")
	}
}
