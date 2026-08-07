// Package output owns semantic command results and their renderers.
package output

import "github.com/projectious-work/ainfra/internal/diagnostic"

// APIVersion is the current machine-result envelope version.
const APIVersion = "ainfra.result/v1"

// Command is a canonical machine-interface command identifier.
type Command string

const (
	// CommandHelp identifies static command help.
	CommandHelp Command = "help"
	// CommandInvocation identifies a failure before command dispatch.
	CommandInvocation Command = "invocation"
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
	if result == nil {
		panic("successful output requires a result")
	}
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
	if len(diagnostics) == 0 {
		panic("failed output requires a diagnostic")
	}
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
	Version                   string                    `json:"version"`
	Commit                    string                    `json:"commit"`
	BuiltAt                   string                    `json:"builtAt"`
	GoVersion                 string                    `json:"goVersion"`
	Platform                  string                    `json:"platform"`
	SupportedContractVersions SupportedContractVersions `json:"supportedContractVersions"`
}

// SupportedContractVersions identifies every contract accepted by the binary.
type SupportedContractVersions struct {
	DocumentAPIVersions          []string `json:"documentApiVersions"`
	ResultAPIVersions            []string `json:"resultApiVersions"`
	StandardOutputSchemaVersions []string `json:"standardOutputSchemaVersions"`
}

// Help is the closed semantic result for static command help.
type Help struct {
	Topic       string          `json:"topic"`
	Usage       string          `json:"usage"`
	Summary     string          `json:"summary"`
	Subcommands []HelpNamedItem `json:"subcommands"`
	Arguments   []HelpArgument  `json:"arguments"`
	Options     []HelpOption    `json:"options"`
}

// HelpNamedItem describes a static child command.
type HelpNamedItem struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// HelpArgument describes one positional argument.
type HelpArgument struct {
	Name       string `json:"name"`
	Required   bool   `json:"required"`
	Repeatable bool   `json:"repeatable"`
	Summary    string `json:"summary"`
}

// HelpOption describes one command option.
type HelpOption struct {
	Names     []string `json:"names"`
	ValueName string   `json:"valueName,omitempty"`
	Summary   string   `json:"summary"`
}
