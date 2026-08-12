package blackbox_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	temporary, err := os.MkdirTemp("", "ainfra-blackbox-")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(temporary, "ainfra")
	build := exec.Command("go", "build", "-o", binary, "../../cmd/ainfra")
	build.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + temporary,
		"GOCACHE=" + filepath.Join(temporary, "go-build"),
		"GOMODCACHE=" + os.Getenv("GOMODCACHE"),
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
	if err := command.Run(); err != nil {
		t.Fatalf("run: %v, stderr: %s", err, stderr.String())
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
	if envelope.Command != "doctor.environment" || !envelope.OK ||
		envelope.Result.Scope != "environment" ||
		len(envelope.Result.EffectiveConfiguration.Values) != 12 || stderr.Len() != 0 {
		t.Fatalf("unexpected result: %+v, stderr=%q", envelope, stderr.String())
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
