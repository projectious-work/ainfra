// Package config resolves immutable, typed CLI configuration with provenance.
package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Layer identifies one normative configuration precedence layer.
type Layer string

const (
	// LayerDefault identifies compiled defaults.
	LayerDefault Layer = "compiled-default"
	// LayerSystem identifies the system configuration file.
	LayerSystem Layer = "system-file"
	// LayerUser identifies the user configuration file.
	LayerUser Layer = "user-file"
	// LayerProject identifies deployment-controlled configuration.
	LayerProject Layer = "project-file"
	// LayerExplicit identifies --config or AINFRA_CONFIG.
	LayerExplicit Layer = "explicit-file"
	// LayerEnvironment identifies supported environment variables.
	LayerEnvironment Layer = "environment"
	// LayerFlag identifies command-line flags.
	LayerFlag Layer = "command-line"
)

// Settings is the immutable effective configuration consumed by commands.
type Settings struct {
	UI          UI
	Logging     Logging
	Paths       Paths
	Executables Executables
}

// UI controls command presentation without changing result semantics.
type UI struct {
	Format         string
	OutputStyle    string
	Color          string
	NonInteractive bool
}

// Logging controls the minimum level and destination configuration.
type Logging struct {
	Level        string
	Destinations []Destination
}

// Destination is one closed logging destination configuration.
type Destination struct {
	Type     string   `yaml:"type"`
	Format   string   `yaml:"format,omitempty"`
	Path     string   `yaml:"path,omitempty"`
	Facility string   `yaml:"facility,omitempty"`
	Tag      string   `yaml:"tag,omitempty"`
	Rotation Rotation `yaml:"rotation,omitempty"`
}

// Rotation controls deterministic file-destination retention.
type Rotation struct {
	MaxSizeMiB int  `yaml:"maxSizeMiB,omitempty"`
	MaxBackups int  `yaml:"maxBackups,omitempty"`
	MaxAgeDays int  `yaml:"maxAgeDays,omitempty"`
	Compress   bool `yaml:"compress,omitempty"`
}

// Paths identifies local cache and run-evidence roots.
type Paths struct {
	Cache string
	Runs  string
}

// Executables identifies optional explicit child-tool paths.
type Executables struct {
	Tofu          string
	AnsibleRunner string
	Git           string
	SSH           string
}

// Entry reports one effective display-safe value and its winning source.
type Entry struct {
	Key        string
	Value      string
	Source     Layer
	Overridden []Layer
}

// FileReport records whether an optional configuration layer was loaded.
type FileReport struct {
	Layer  Layer
	Path   string
	Status string
}

// Effective combines immutable settings with deterministic provenance.
type Effective struct {
	Settings Settings
	Entries  []Entry
	Files    []FileReport
	Rejected []string
}

// Defaults returns the closed baseline configuration for local resolution.
func Defaults(cacheDirectory, runDirectory string) Settings {
	return Settings{
		UI: UI{Format: "text", OutputStyle: "auto", Color: "auto"},
		Logging: Logging{
			Level:        "warn",
			Destinations: []Destination{{Type: "stderr", Format: "text"}},
		},
		Paths: Paths{Cache: cacheDirectory, Runs: runDirectory},
	}
}

// EnvironmentPatch parses only the supported, non-secret AINFRA variables.
func EnvironmentPatch(environment map[string]string) (Patch, error) {
	var patch Patch
	assignString(&patch.UI.Format, environment, "AINFRA_FORMAT")
	assignString(&patch.UI.OutputStyle, environment, "AINFRA_OUTPUT_STYLE")
	assignString(&patch.UI.Color, environment, "AINFRA_COLOR")
	assignString(&patch.Logging.Level, environment, "AINFRA_LOG_LEVEL")
	assignString(&patch.Paths.Cache, environment, "AINFRA_CACHE_DIR")
	assignString(&patch.Paths.Runs, environment, "AINFRA_RUN_DIR")
	assignString(&patch.Executables.Tofu, environment, "AINFRA_TOFU_PATH")
	assignString(&patch.Executables.AnsibleRunner, environment, "AINFRA_ANSIBLE_RUNNER_PATH")
	assignString(&patch.Executables.Git, environment, "AINFRA_GIT_PATH")
	assignString(&patch.Executables.SSH, environment, "AINFRA_SSH_PATH")
	if _, present := environment["NO_COLOR"]; present {
		value := "never"
		patch.UI.Color = &value
	}
	if value, present := environment["AINFRA_NON_INTERACTIVE"]; present {
		parsed, err := parseBoolean("AINFRA_NON_INTERACTIVE", value)
		if err != nil {
			return Patch{}, err
		}
		patch.UI.NonInteractive = &parsed
	}
	if value, present := environment["AINFRA_LOG_SYSLOG"]; present {
		if _, err := parseBoolean("AINFRA_LOG_SYSLOG", value); err != nil {
			return Patch{}, err
		}
	}
	for name, value := range map[string]string{
		"AINFRA_CACHE_DIR":           environment["AINFRA_CACHE_DIR"],
		"AINFRA_RUN_DIR":             environment["AINFRA_RUN_DIR"],
		"AINFRA_LOG_FILE":            environment["AINFRA_LOG_FILE"],
		"AINFRA_TOFU_PATH":           environment["AINFRA_TOFU_PATH"],
		"AINFRA_ANSIBLE_RUNNER_PATH": environment["AINFRA_ANSIBLE_RUNNER_PATH"],
		"AINFRA_GIT_PATH":            environment["AINFRA_GIT_PATH"],
		"AINFRA_SSH_PATH":            environment["AINFRA_SSH_PATH"],
	} {
		if value != "" && !filepath.IsAbs(value) {
			return Patch{}, fmt.Errorf("%s must be an absolute path", name)
		}
	}
	return patch, validatePatch(patch)
}

func assignString(target **string, environment map[string]string, name string) {
	if value, present := environment[name]; present {
		copy := value
		*target = &copy
	}
}

func parseBoolean(name, value string) (bool, error) {
	if value == "" {
		return false, fmt.Errorf("%s must not be empty", name)
	}
	parsed, err := strconv.ParseBool(strings.ToLower(value))
	if err != nil || (strings.ToLower(value) != "true" && strings.ToLower(value) != "false") {
		return false, fmt.Errorf("%s must be true or false", name)
	}
	return parsed, nil
}

func validatePatch(patch Patch) error {
	return firstInvalid(
		validateEnum("ui.format", patch.UI.Format, "text", "json"),
		validateEnum("ui.outputStyle", patch.UI.OutputStyle, "auto", "rich", "plain"),
		validateEnum("ui.color", patch.UI.Color, "auto", "always", "never"),
		validateEnum("logging.level", patch.Logging.Level, "error", "warn", "info", "debug", "trace"),
	)
}

func validateEnum(field string, value *string, accepted ...string) error {
	if value == nil {
		return nil
	}
	for _, candidate := range accepted {
		if *value == candidate {
			return nil
		}
	}
	return fmt.Errorf("invalid %s %q", field, *value)
}

func firstInvalid(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func sortedEntries(values map[string]provenance) []Entry {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]Entry, 0, len(keys))
	for _, key := range keys {
		value := values[key]
		entries = append(entries, Entry{
			Key: key, Value: value.value, Source: value.source,
			Overridden: append([]Layer(nil), value.overridden...),
		})
	}
	return entries
}

type provenance struct {
	value      string
	source     Layer
	overridden []Layer
}
