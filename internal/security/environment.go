package security

import (
	"fmt"
	"sort"
	"strings"
)

// Environment is an immutable, explicit child-process environment.
type Environment struct {
	values []string
}

// BuildEnvironment selects only named variables from an injected parent.
func BuildEnvironment(parent []string, allowed []string, explicit map[string]string) (Environment, error) {
	wanted := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		if !validEnvironmentName(name) {
			return Environment{}, &Refusal{Policy: "environment", Reason: "invalid allowed variable name"}
		}
		wanted[name] = struct{}{}
	}
	selected := make(map[string]string, len(wanted)+len(explicit))
	seen := make(map[string]struct{}, len(parent))
	for _, entry := range parent {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || !validEnvironmentName(name) {
			return Environment{}, &Refusal{Policy: "environment", Reason: "malformed parent entry"}
		}
		if _, duplicate := seen[name]; duplicate {
			return Environment{}, &Refusal{Policy: "environment", Reason: "duplicate parent variable"}
		}
		seen[name] = struct{}{}
		if _, include := wanted[name]; include {
			selected[name] = value
		}
	}
	for name, value := range explicit {
		if !validEnvironmentName(name) {
			return Environment{}, &Refusal{Policy: "environment", Reason: "invalid explicit variable name"}
		}
		selected[name] = value
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, fmt.Sprintf("%s=%s", name, selected[name]))
	}
	return Environment{values: values}, nil
}

// Values returns a defensive copy suitable for os/exec.Cmd.Env.
func (environment Environment) Values() []string {
	values := make([]string, len(environment.values))
	copy(values, environment.values)
	return values
}

func validEnvironmentName(name string) bool {
	if name == "" || strings.Contains(name, "=") {
		return false
	}
	for index, character := range name {
		if character == '_' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}
