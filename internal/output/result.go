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
	// CommandInit identifies local deployment initialization.
	CommandInit Command = "init"
	// CommandDoctorEnvironment identifies local environment diagnostics.
	CommandDoctorEnvironment Command = "doctor.environment"
	// CommandDoctorDeployment identifies local deployment diagnostics.
	CommandDoctorDeployment Command = "doctor.deployment"
	// CommandDoctorTemplate identifies local resolved-template diagnostics.
	CommandDoctorTemplate Command = "doctor.template"
	// CommandDoctorRun identifies retained local run diagnostics.
	CommandDoctorRun Command = "doctor.run"
	// CommandDoctorAll identifies the complete applicable diagnostic set.
	CommandDoctorAll Command = "doctor.all"
	// CommandTemplateLock identifies initial immutable template locking.
	CommandTemplateLock Command = "template.lock"
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

// PartialFailure constructs an unsuccessful command that still produced a
// schema-valid meaningful result, such as a complete doctor report.
func PartialFailure(
	command Command,
	result any,
	diagnostics ...diagnostic.Diagnostic,
) Envelope {
	if result == nil || len(diagnostics) == 0 {
		panic("partial failure requires a result and diagnostic")
	}
	return Envelope{
		APIVersion: APIVersion, Command: command, OK: false, Result: result,
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

// Doctor is the semantic result shared by every doctor scope.
type Doctor struct {
	Scope                  string                  `json:"scope"`
	Summary                DoctorSummary           `json:"summary"`
	Findings               []diagnostic.Diagnostic `json:"findings"`
	EffectiveConfiguration *EffectiveConfiguration `json:"effectiveConfiguration,omitempty"`
}

// DoctorSummary counts every registered check outcome.
type DoctorSummary struct {
	Pass    int `json:"pass"`
	Skip    int `json:"skip"`
	Warning int `json:"warning"`
	Fail    int `json:"fail"`
}

// Init is the semantic result of creating a minimal deployment contract.
type Init struct {
	Deployment   Deployment `json:"deployment"`
	CreatedPaths []string   `json:"createdPaths"`
}

// Template is the semantic result of a template lock mutation.
type Template struct {
	Source           string `json:"source"`
	ResolvedRevision string `json:"resolvedRevision"`
	ContentDigest    string `json:"contentDigest"`
	Changed          bool   `json:"changed"`
}

// Deployment identifies one canonical local deployment.
type Deployment struct {
	Name string `json:"name"`
	Root string `json:"root"`
}

// EffectiveConfiguration is the display-safe configuration/provenance view.
type EffectiveConfiguration struct {
	Values                  map[string]EffectiveConfigurationValue `json:"values"`
	Files                   []ConfigurationFile                    `json:"files"`
	RejectedProjectSettings []RejectedProjectSetting               `json:"rejectedProjectSettings"`
}

// EffectiveConfigurationValue reports one value without exposing secrets.
type EffectiveConfigurationValue struct {
	DisplayValue      string   `json:"displayValue"`
	Source            string   `json:"source"`
	SourceDetail      string   `json:"sourceDetail,omitempty"`
	OverriddenSources []string `json:"overriddenSources"`
}

// ConfigurationFile reports one normative file layer.
type ConfigurationFile struct {
	Layer  string `json:"layer"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

// RejectedProjectSetting explains one prohibited repository-controlled key.
type RejectedProjectSetting struct {
	Key     string `json:"key"`
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
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
