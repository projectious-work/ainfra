package blackbox_test

import (
	"bytes"
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
	build.Env = append([]string{}, os.Environ()...)
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
	command := exec.Command(binary, "--format", "json", "version")
	command.Env = []string{"TERM=dumb"}
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
			Version   string `json:"version"`
			Commit    string `json:"commit"`
			BuiltAt   string `json:"builtAt"`
			GoVersion string `json:"goVersion"`
			Platform  string `json:"platform"`
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
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	command := exec.Command(binary, "unknown")
	command.Env = []string{}
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
