package command_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/command"
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
		OK          bool  `json:"ok"`
		Result      any   `json:"result"`
		Diagnostics []any `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if envelope.OK || envelope.Result != nil || len(envelope.Diagnostics) != 1 {
		t.Errorf("unexpected failure envelope: %+v", envelope)
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
