package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"

	"github.com/projectious-work/ainfra/internal/security"
	"go.yaml.in/yaml/v3"
)

const configAPIVersion = "ainfra.projectious.work/v1"

// File identifies one ordered optional or required configuration layer.
type File struct {
	Layer    Layer
	Path     string
	Required bool
}

// ResolveOptions supplies every configuration layer without global mutation.
type ResolveOptions struct {
	Defaults          Settings
	Files             []File
	Environment       map[string]string
	Flags             Patch
	DiagnosticProject bool
}

// LocationOptions supplies host facts used to construct the four normative
// file layers. Environment is injected so tests never depend on host state.
type LocationOptions struct {
	GOOS           string
	HomeDirectory  string
	XDGConfigHome  string
	DeploymentRoot string
	ExplicitPath   string
}

// Files returns the normative system, user, project, and explicit file layers
// in ascending precedence. It never searches the working directory.
func Files(options LocationOptions) ([]File, error) {
	goos := options.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	if options.HomeDirectory == "" {
		return nil, fmt.Errorf("home directory is required")
	}
	var systemPath, userPath string
	switch goos {
	case "linux":
		systemPath = "/etc/ainfra/config.yaml"
		userRoot := options.XDGConfigHome
		if userRoot == "" {
			userRoot = filepath.Join(options.HomeDirectory, ".config")
		}
		userPath = filepath.Join(userRoot, "ainfra", "config.yaml")
	case "darwin":
		systemPath = "/Library/Application Support/ainfra/config.yaml"
		userPath = filepath.Join(
			options.HomeDirectory, "Library", "Application Support", "ainfra", "config.yaml",
		)
	default:
		return nil, fmt.Errorf("unsupported host operating system %q", goos)
	}
	files := []File{
		{Layer: LayerSystem, Path: systemPath},
		{Layer: LayerUser, Path: userPath},
	}
	if options.DeploymentRoot != "" {
		files = append(files, File{
			Layer: LayerProject,
			Path:  filepath.Join(options.DeploymentRoot, "ainfra.config.yaml"),
		})
	}
	if options.ExplicitPath != "" {
		files = append(files, File{
			Layer: LayerExplicit, Path: options.ExplicitPath, Required: true,
		})
	}
	return files, nil
}

// Patch is a sparse typed configuration layer. Pointer fields distinguish an
// explicit zero value from an absent value during deterministic merging.
type Patch struct {
	APIVersion  string           `yaml:"apiVersion"`
	Kind        string           `yaml:"kind"`
	UI          UIPatch          `yaml:"ui,omitempty"`
	Logging     LoggingPatch     `yaml:"logging,omitempty"`
	Paths       PathsPatch       `yaml:"paths,omitempty"`
	Executables ExecutablesPatch `yaml:"executables,omitempty"`
}

// UIPatch is the sparse UI portion of a configuration layer.
type UIPatch struct {
	Format         *string `yaml:"format,omitempty"`
	OutputStyle    *string `yaml:"outputStyle,omitempty"`
	Color          *string `yaml:"color,omitempty"`
	NonInteractive *bool   `yaml:"nonInteractive,omitempty"`
}

// LoggingPatch is the sparse logging portion of a configuration layer.
type LoggingPatch struct {
	Level        *string        `yaml:"level,omitempty"`
	Destinations *[]Destination `yaml:"destinations,omitempty"`
}

// PathsPatch is the sparse storage portion of a configuration layer.
type PathsPatch struct {
	Cache *string `yaml:"cache,omitempty"`
	Runs  *string `yaml:"runs,omitempty"`
}

// ExecutablesPatch is the sparse executable portion of a configuration layer.
type ExecutablesPatch struct {
	Tofu          *string `yaml:"tofu,omitempty"`
	AnsibleRunner *string `yaml:"ansibleRunner,omitempty"`
	Git           *string `yaml:"git,omitempty"`
	SSH           *string `yaml:"ssh,omitempty"`
}

// Resolve loads and merges configuration in the exact order supplied, then
// applies environment and flag layers. The returned value owns all slices.
func Resolve(options ResolveOptions) (Effective, error) {
	settings := cloneSettings(options.Defaults)
	provenanceValues := defaultProvenance(settings)
	report := Effective{}
	for _, file := range options.Files {
		patch, status, err := loadFile(file)
		report.Files = append(report.Files, FileReport{
			Layer: file.Layer, Path: file.Path, Status: status,
		})
		if err != nil {
			return Effective{}, err
		}
		if status != "loaded" {
			continue
		}
		if file.Layer == LayerProject {
			rejected := prohibitedProjectKeys(patch)
			report.Rejected = append(report.Rejected, rejected...)
			if len(rejected) > 0 && !options.DiagnosticProject {
				return Effective{}, fmt.Errorf(
					"project configuration contains prohibited keys: %v", rejected,
				)
			}
			patch = safeProjectPatch(patch)
		}
		apply(&settings, provenanceValues, patch, file.Layer, filepath.Dir(file.Path))
	}
	environmentPatch, err := EnvironmentPatch(options.Environment)
	if err != nil {
		return Effective{}, err
	}
	apply(&settings, provenanceValues, environmentPatch, LayerEnvironment, "")
	if err := validatePatch(options.Flags); err != nil {
		return Effective{}, err
	}
	apply(&settings, provenanceValues, options.Flags, LayerFlag, "")
	report.Settings = cloneSettings(settings)
	report.Entries = sortedEntries(provenanceValues)
	sort.Strings(report.Rejected)
	return report, nil
}

func loadFile(file File) (Patch, string, error) {
	if file.Path == "" {
		return Patch{}, "absent", nil
	}
	if err := security.RequireRegular(file.Path); err != nil {
		if !file.Required && errors.Is(err, os.ErrNotExist) {
			return Patch{}, "absent", nil
		}
		return Patch{}, "rejected", fmt.Errorf("load %s config: %w", file.Layer, err)
	}
	contents, err := os.ReadFile(file.Path)
	if err != nil {
		return Patch{}, "rejected", fmt.Errorf("read %s config: %w", file.Layer, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var patch Patch
	if err := decoder.Decode(&patch); err != nil {
		return Patch{}, "rejected", fmt.Errorf("parse %s config: %w", file.Layer, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Patch{}, "rejected", fmt.Errorf("%s config must contain one document", file.Layer)
		}
		return Patch{}, "rejected", fmt.Errorf("parse trailing %s config: %w", file.Layer, err)
	}
	if patch.APIVersion != configAPIVersion || patch.Kind != "CLIConfig" {
		return Patch{}, "rejected", fmt.Errorf("unsupported %s config contract", file.Layer)
	}
	if err := validatePatch(patch); err != nil {
		return Patch{}, "rejected", err
	}
	return patch, "loaded", nil
}

func safeProjectPatch(patch Patch) Patch {
	return Patch{UI: patch.UI, Logging: LoggingPatch{Level: patch.Logging.Level}}
}

func prohibitedProjectKeys(patch Patch) []string {
	var rejected []string
	if patch.Logging.Destinations != nil {
		rejected = append(rejected, "logging.destinations")
	}
	if patch.Paths.Cache != nil {
		rejected = append(rejected, "paths.cache")
	}
	if patch.Paths.Runs != nil {
		rejected = append(rejected, "paths.runs")
	}
	for key, value := range map[string]*string{
		"executables.ansibleRunner": patch.Executables.AnsibleRunner,
		"executables.git":           patch.Executables.Git,
		"executables.ssh":           patch.Executables.SSH,
		"executables.tofu":          patch.Executables.Tofu,
	} {
		if value != nil {
			rejected = append(rejected, key)
		}
	}
	sort.Strings(rejected)
	return rejected
}

func apply(settings *Settings, values map[string]provenance, patch Patch, layer Layer, base string) {
	setString(&settings.UI.Format, patch.UI.Format, "ui.format", layer, values, base, false)
	setString(&settings.UI.OutputStyle, patch.UI.OutputStyle, "ui.outputStyle", layer, values, base, false)
	setString(&settings.UI.Color, patch.UI.Color, "ui.color", layer, values, base, false)
	if patch.UI.NonInteractive != nil {
		settings.UI.NonInteractive = *patch.UI.NonInteractive
		setProvenance(values, "ui.nonInteractive", strconv.FormatBool(*patch.UI.NonInteractive), layer)
	}
	setString(&settings.Logging.Level, patch.Logging.Level, "logging.level", layer, values, base, false)
	if patch.Logging.Destinations != nil {
		settings.Logging.Destinations = append([]Destination(nil), (*patch.Logging.Destinations)...)
		setProvenance(values, "logging.destinations", fmt.Sprintf("%d configured", len(*patch.Logging.Destinations)), layer)
	}
	setString(&settings.Paths.Cache, patch.Paths.Cache, "paths.cache", layer, values, base, true)
	setString(&settings.Paths.Runs, patch.Paths.Runs, "paths.runs", layer, values, base, true)
	setString(&settings.Executables.Tofu, patch.Executables.Tofu, "executables.tofu", layer, values, base, true)
	setString(&settings.Executables.AnsibleRunner, patch.Executables.AnsibleRunner, "executables.ansibleRunner", layer, values, base, true)
	setString(&settings.Executables.Git, patch.Executables.Git, "executables.git", layer, values, base, true)
	setString(&settings.Executables.SSH, patch.Executables.SSH, "executables.ssh", layer, values, base, true)
}

func setString(target *string, value *string, key string, layer Layer, values map[string]provenance, base string, path bool) {
	if value == nil {
		return
	}
	resolved := *value
	if path && base != "" && resolved != "" && !filepath.IsAbs(resolved) {
		resolved = filepath.Join(base, resolved)
	}
	*target = resolved
	setProvenance(values, key, resolved, layer)
}

func setProvenance(values map[string]provenance, key, value string, layer Layer) {
	previous, exists := values[key]
	if exists && previous.source != layer {
		previous.overridden = append(previous.overridden, previous.source)
	}
	previous.value = value
	previous.source = layer
	values[key] = previous
}

func defaultProvenance(settings Settings) map[string]provenance {
	values := map[string]provenance{}
	setProvenance(values, "ui.format", settings.UI.Format, LayerDefault)
	setProvenance(values, "ui.outputStyle", settings.UI.OutputStyle, LayerDefault)
	setProvenance(values, "ui.color", settings.UI.Color, LayerDefault)
	setProvenance(values, "ui.nonInteractive", strconv.FormatBool(settings.UI.NonInteractive), LayerDefault)
	setProvenance(values, "logging.level", settings.Logging.Level, LayerDefault)
	setProvenance(values, "logging.destinations", fmt.Sprintf("%d configured", len(settings.Logging.Destinations)), LayerDefault)
	setProvenance(values, "paths.cache", settings.Paths.Cache, LayerDefault)
	setProvenance(values, "paths.runs", settings.Paths.Runs, LayerDefault)
	setProvenance(values, "executables.tofu", settings.Executables.Tofu, LayerDefault)
	setProvenance(values, "executables.ansibleRunner", settings.Executables.AnsibleRunner, LayerDefault)
	setProvenance(values, "executables.git", settings.Executables.Git, LayerDefault)
	setProvenance(values, "executables.ssh", settings.Executables.SSH, LayerDefault)
	return values
}

func cloneSettings(settings Settings) Settings {
	settings.Logging.Destinations = append([]Destination(nil), settings.Logging.Destinations...)
	return settings
}
