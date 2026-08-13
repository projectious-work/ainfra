// Package diagnostic defines stable, safe findings presented to ainfra users.
package diagnostic

import "regexp"

var codePattern = regexp.MustCompile(`^AINFRA-E[0-9]{4}$`)

// Severity classifies the importance of a diagnostic.
type Severity string

const (
	// SeverityInfo reports useful information that is not a failure.
	SeverityInfo Severity = "info"
	// SeverityWarning reports a condition that needs attention.
	SeverityWarning Severity = "warning"
	// SeverityError reports a failed request or refused operation.
	SeverityError Severity = "error"
)

// Diagnostic is a stable machine-readable user finding.
type Diagnostic struct {
	Code             string   `json:"code"`
	Severity         Severity `json:"severity"`
	Check            string   `json:"check,omitempty"`
	Scope            string   `json:"scope,omitempty"`
	Status           string   `json:"status,omitempty"`
	Reconciliation   string   `json:"reconciliation,omitempty"`
	Message          string   `json:"message"`
	Component        string   `json:"component,omitempty"`
	Path             string   `json:"path,omitempty"`
	Evidence         string   `json:"evidence,omitempty"`
	NextAction       string   `json:"nextAction,omitempty"`
	DocumentationURL string   `json:"documentationUrl,omitempty"`
}

// ValidCode reports whether code follows the stable ainfra diagnostic format.
func ValidCode(code string) bool {
	return codePattern.MatchString(code)
}
