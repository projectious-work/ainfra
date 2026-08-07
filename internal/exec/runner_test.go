package exec_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/security"
)

func TestRunnerPreservesArgumentsAndRedacts(t *testing.T) {
	executable, err := security.ResolveExecutable(mustExecutable(t))
	if err != nil {
		t.Fatalf("ResolveExecutable: %v", err)
	}
	environment, err := security.BuildEnvironment(nil, nil, map[string]string{
		"AINFRA_TEST_HELPER": "echo",
	})
	if err != nil {
		t.Fatalf("BuildEnvironment: %v", err)
	}
	var stdout bytes.Buffer
	result, err := (childexec.Runner{}).Run(context.Background(), childexec.Request{
		Executable:      executable,
		Args:            []string{"-test.run=TestHelperProcess", "--", "; touch /tmp/not-run", "secret"},
		WorkingRoot:     t.TempDir(),
		WorkingDir:      ".",
		Environment:     environment,
		IO:              childexec.IOPolicy{Stdout: &stdout},
		SensitiveValues: []string{"secret"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit = %d", result.ExitCode)
	}
	if !strings.Contains(stdout.String(), `arg="; touch /tmp/not-run"`) {
		t.Errorf("argument boundary lost: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "secret") || !strings.Contains(stdout.String(), "<redacted>") {
		t.Errorf("sensitive output not redacted: %q", stdout.String())
	}
}

func TestRunnerCancellation(t *testing.T) {
	executable, err := security.ResolveExecutable(mustExecutable(t))
	if err != nil {
		t.Fatal(err)
	}
	environment, err := security.BuildEnvironment(nil, nil, map[string]string{
		"AINFRA_TEST_HELPER": "wait",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := (childexec.Runner{CancellationGrace: 100 * time.Millisecond}).Run(ctx, childexec.Request{
		Executable:  executable,
		Args:        []string{"-test.run=TestHelperProcess"},
		WorkingRoot: t.TempDir(),
		WorkingDir:  ".",
		Environment: environment,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Cancelled {
		t.Error("cancelled result not reported")
	}
}

func mustExecutable(t *testing.T) string {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHelperProcess(t *testing.T) {
	mode := os.Getenv("AINFRA_TEST_HELPER")
	if mode == "" {
		return
	}
	switch mode {
	case "echo":
		for _, argument := range os.Args {
			if argument == "--" {
				continue
			}
			if strings.HasPrefix(argument, "-test.") {
				continue
			}
			fmt.Printf("arg=%q\n", argument)
		}
	case "wait":
		select {}
	default:
		os.Exit(2)
	}
	os.Exit(0)
}
