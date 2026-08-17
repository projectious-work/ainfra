package doctor

import (
	"context"
	"fmt"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

// EnvironmentRegistry returns the closed initial environment-check catalog.
func EnvironmentRegistry() Registry {
	definitions := []Definition{
		platformCheck(),
		executableCheck("ansible-runner", "AINFRA-E2202", false),
		executableCheck("git", "AINFRA-E2203", false),
		executableCheck("ssh", "AINFRA-E2204", false),
		executableCheck("tofu", "AINFRA-E2201", true),
	}
	registry, err := NewRegistry(definitions...)
	if err != nil {
		panic(err)
	}
	return registry
}

func platformCheck() Definition {
	return Definition{
		ID: "environment.platform", Scope: ScopeEnvironment,
		Run: func(_ context.Context, input Input, _ Capabilities) diagnostic.Diagnostic {
			supportedOS := input.GOOS == "linux" || input.GOOS == "darwin"
			supportedArch := input.GOARCH == "amd64" || input.GOARCH == "arm64"
			if !supportedOS || !supportedArch {
				return diagnostic.Diagnostic{
					Code: "AINFRA-E2200", Severity: diagnostic.SeverityError,
					Status: "fail", Component: "environment",
					Message:    fmt.Sprintf("unsupported platform %s/%s", input.GOOS, input.GOARCH),
					NextAction: "Use Linux or macOS on amd64 or arm64.",
				}
			}
			return diagnostic.Diagnostic{
				Code: "AINFRA-E2200", Severity: diagnostic.SeverityInfo,
				Status: "pass", Component: "environment",
				Message: fmt.Sprintf("supported platform %s/%s", input.GOOS, input.GOARCH),
			}
		},
	}
}

func executableCheck(name, code string, required bool) Definition {
	return Definition{
		ID:    "environment.executable." + name,
		Scope: ScopeEnvironment, ChildToolNeeds: []string{name},
		Run: func(ctx context.Context, input Input, capabilities Capabilities) diagnostic.Diagnostic {
			if capabilities.InspectExecutable == nil {
				return unavailableExecutable(name, code, required, "executable discovery is unavailable")
			}
			fact, err := capabilities.InspectExecutable(
				ctx, name, input.Executables[name],
			)
			if err != nil {
				return unavailableExecutable(name, code, required, err.Error())
			}
			return diagnostic.Diagnostic{
				Code: code, Severity: diagnostic.SeverityInfo, Status: "pass",
				Component: name, Path: fact.Path, Evidence: fact.Version,
				Message: fmt.Sprintf("%s executable is available", name),
			}
		},
	}
}

func unavailableExecutable(name, code string, required bool, evidence string) diagnostic.Diagnostic {
	severity := diagnostic.SeverityInfo
	status := "skip"
	message := fmt.Sprintf("%s executable is unavailable; dependent checks are skipped", name)
	if required {
		severity = diagnostic.SeverityError
		status = "fail"
		message = fmt.Sprintf("required %s executable is unavailable", name)
	}
	return diagnostic.Diagnostic{
		Code: code, Severity: severity, Status: status, Component: name,
		Message: message, Evidence: evidence,
		NextAction: fmt.Sprintf("Install %s or configure its absolute executable path.", name),
	}
}
