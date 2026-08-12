// Package doctor defines capability-aware, deterministic diagnostic checks.
package doctor

import (
	"context"
	"fmt"
	"sort"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

// Scope identifies the local subject examined by a doctor check.
type Scope string

const (
	// ScopeEnvironment examines host configuration and capabilities.
	ScopeEnvironment Scope = "environment"
	// ScopeDeployment examines an ainfra deployment.
	ScopeDeployment Scope = "deployment"
	// ScopeTemplate examines an already-resolved local template.
	ScopeTemplate Scope = "template"
	// ScopeRun examines retained local run evidence.
	ScopeRun Scope = "run"
)

// Capabilities are explicit effects a check may request. Checks receive no
// implicit filesystem, process, environment, or network authority.
type Capabilities struct {
	InspectExecutable func(
		context.Context, string, string,
	) (ExecutableFact, error)
}

// ExecutableFact is a validated local executable identity and reported version.
type ExecutableFact struct {
	Path    string
	Version string
}

// Input is the immutable state shared with applicable checks.
type Input struct {
	GOOS        string
	GOARCH      string
	Executables map[string]string
}

// Definition describes one check independently from its execution order.
type Definition struct {
	ID             string
	Scope          Scope
	Prerequisites  []string
	ChildToolNeeds []string
	Applicable     func(Input) bool
	Run            func(context.Context, Input, Capabilities) diagnostic.Diagnostic
}

// Report contains stable findings and their computed summary.
type Report struct {
	Findings []diagnostic.Diagnostic
	Summary  Summary
}

// Summary counts all registered applicable and skipped check outcomes.
type Summary struct {
	Pass    int
	Skip    int
	Warning int
	Fail    int
}

// Registry validates check metadata once and executes checks deterministically.
type Registry struct {
	definitions []Definition
}

// NewRegistry constructs a registry and rejects duplicate or incomplete checks.
func NewRegistry(definitions ...Definition) (Registry, error) {
	copied := append([]Definition(nil), definitions...)
	sort.Slice(copied, func(left, right int) bool {
		if copied[left].Scope != copied[right].Scope {
			return copied[left].Scope < copied[right].Scope
		}
		return copied[left].ID < copied[right].ID
	})
	for index, definition := range copied {
		if definition.ID == "" || definition.Scope == "" || definition.Run == nil {
			return Registry{}, fmt.Errorf("doctor check %d is incomplete", index)
		}
		if index > 0 && copied[index-1].Scope == definition.Scope &&
			copied[index-1].ID == definition.ID {
			return Registry{}, fmt.Errorf("duplicate doctor check %q", definition.ID)
		}
		copied[index].Prerequisites = sortedCopy(definition.Prerequisites)
		copied[index].ChildToolNeeds = sortedCopy(definition.ChildToolNeeds)
	}
	return Registry{definitions: copied}, nil
}

// Run executes one scope without short-circuiting on failed findings.
func (registry Registry) Run(
	ctx context.Context,
	scope Scope,
	input Input,
	capabilities Capabilities,
) Report {
	report := Report{Findings: []diagnostic.Diagnostic{}}
	for _, definition := range registry.definitions {
		if definition.Scope != scope {
			continue
		}
		if definition.Applicable != nil && !definition.Applicable(input) {
			continue
		}
		finding := definition.Run(ctx, input, capabilities)
		finding.Check = definition.ID
		finding.Scope = string(definition.Scope)
		if finding.Reconciliation == "" {
			finding.Reconciliation = "not_available"
		}
		report.Findings = append(report.Findings, finding)
		increment(&report.Summary, finding.Status)
	}
	sort.SliceStable(report.Findings, func(left, right int) bool {
		leftFinding, rightFinding := report.Findings[left], report.Findings[right]
		if statusOrder(leftFinding.Status) != statusOrder(rightFinding.Status) {
			return statusOrder(leftFinding.Status) < statusOrder(rightFinding.Status)
		}
		if leftFinding.Check != rightFinding.Check {
			return leftFinding.Check < rightFinding.Check
		}
		if leftFinding.Path != rightFinding.Path {
			return leftFinding.Path < rightFinding.Path
		}
		return leftFinding.Code < rightFinding.Code
	})
	return report
}

func increment(summary *Summary, status string) {
	switch status {
	case "pass":
		summary.Pass++
	case "skip":
		summary.Skip++
	case "warning":
		summary.Warning++
	case "fail":
		summary.Fail++
	}
}

func statusOrder(status string) int {
	switch status {
	case "fail":
		return 0
	case "warning":
		return 1
	case "skip":
		return 2
	default:
		return 3
	}
}

func sortedCopy(values []string) []string {
	copy := append([]string(nil), values...)
	sort.Strings(copy)
	return copy
}
