package doctor_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/doctor"
)

func TestEnvironmentRegistryIsCompleteAndDeterministic(t *testing.T) {
	t.Parallel()
	found := map[string]string{"tofu": "/bin/tofu", "git": "/bin/git"}
	capabilities := doctor.Capabilities{
		InspectExecutable: func(
			_ context.Context, name, configured string,
		) (doctor.ExecutableFact, error) {
			if configured != "" {
				return doctor.ExecutableFact{Path: configured, Version: "v1"}, nil
			}
			path, ok := found[name]
			if !ok {
				return doctor.ExecutableFact{}, errors.New("not found")
			}
			return doctor.ExecutableFact{Path: path, Version: "v1"}, nil
		},
	}
	report := doctor.EnvironmentRegistry().Run(
		context.Background(), doctor.ScopeEnvironment,
		doctor.Input{GOOS: "linux", GOARCH: "arm64", Executables: map[string]string{}},
		capabilities,
	)
	if report.Summary.Pass != 3 || report.Summary.Skip != 2 ||
		report.Summary.Warning != 0 || report.Summary.Fail != 0 {
		t.Fatalf("summary = %#v", report.Summary)
	}
	checks := make([]string, len(report.Findings))
	for index, finding := range report.Findings {
		checks[index] = finding.Status + ":" + finding.Check
		if finding.Status == "skip" && finding.NextAction == "" {
			t.Fatalf("skip lacks next action: %#v", finding)
		}
	}
	want := []string{
		"skip:environment.executable.ansible-runner",
		"skip:environment.executable.ssh",
		"pass:environment.executable.git",
		"pass:environment.executable.tofu",
		"pass:environment.platform",
	}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("ordered checks = %#v, want %#v", checks, want)
	}
}

func TestRequiredTofuUnavailableFailsWithoutStoppingOtherChecks(t *testing.T) {
	t.Parallel()
	report := doctor.EnvironmentRegistry().Run(
		context.Background(), doctor.ScopeEnvironment,
		doctor.Input{GOOS: "plan9", GOARCH: "mips", Executables: map[string]string{}},
		doctor.Capabilities{InspectExecutable: func(
			context.Context, string, string,
		) (doctor.ExecutableFact, error) {
			return doctor.ExecutableFact{}, errors.New("missing")
		}},
	)
	if report.Summary.Fail != 2 || len(report.Findings) != 5 {
		t.Fatalf("incomplete report: %#v", report)
	}
}

func TestRegistryRejectsDuplicateChecks(t *testing.T) {
	t.Parallel()
	definition := doctor.Definition{
		ID: "environment.same", Scope: doctor.ScopeEnvironment,
		Run: func(context.Context, doctor.Input, doctor.Capabilities) diagnostic.Diagnostic {
			return diagnostic.Diagnostic{}
		},
	}
	if _, err := doctor.NewRegistry(definition, definition); err == nil {
		t.Fatal("duplicate registry entry accepted")
	}
}
