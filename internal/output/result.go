// Package output owns semantic command results and their renderers.
package output

import "github.com/projectious-work/ainfra/internal/diagnostic"

// APIVersion is the current machine-result envelope version.
const APIVersion = "ainfra.result/v1"

// Command is a canonical machine-interface command identifier.
type Command string

const (
	// CommandVersion identifies the version command.
	CommandVersion Command = "version"
)

// Envelope is the closed top-level machine-result contract.
type Envelope struct {
	APIVersion  string                  `json:"apiVersion"`
	Command     Command                 `json:"command"`
	OK          bool                    `json:"ok"`
	Result      any                     `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// Success constructs a successful result with a non-null semantic value.
func Success(command Command, result any) Envelope {
	return Envelope{
		APIVersion:  APIVersion,
		Command:     command,
		OK:          true,
		Result:      result,
		Diagnostics: []diagnostic.Diagnostic{},
	}
}

// Failure constructs an unsuccessful result with at least one diagnostic.
func Failure(command Command, diagnostics ...diagnostic.Diagnostic) Envelope {
	return Envelope{
		APIVersion:  APIVersion,
		Command:     command,
		OK:          false,
		Result:      nil,
		Diagnostics: diagnostics,
	}
}

// Version is the semantic result for the version command.
type Version struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuiltAt   string `json:"builtAt"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}
