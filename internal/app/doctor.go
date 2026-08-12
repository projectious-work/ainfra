package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/doctor"
	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
)

// DoctorEnvironmentRequest contains explicit CLI selections after parsing.
type DoctorEnvironmentRequest struct {
	ConfigPath  string
	ProjectPath string
	Format      *string
	OutputStyle *string
	Color       *string
}

// DoctorEnvironmentOptions supplies immutable host facts to the use case.
type DoctorEnvironmentOptions struct {
	GOOS              string
	GOARCH            string
	WorkingDirectory  string
	HomeDirectory     string
	XDGConfigHome     string
	CacheDirectory    string
	RunDirectory      string
	Environment       map[string]string
	InspectExecutable func(
		context.Context, string, string,
	) (doctor.ExecutableFact, error)
}

// DoctorEnvironmentResponse carries the semantic result and resolved display
// settings. Rendering remains owned by the command/output layers.
type DoctorEnvironmentResponse struct {
	Result      output.Doctor
	Format      string
	OutputStyle string
	Color       string
}

// DoctorEnvironment validates local configuration and host support without
// invoking child tools, using network access, or changing the filesystem.
func DoctorEnvironment(
	request DoctorEnvironmentRequest,
	options DoctorEnvironmentOptions,
) (DoctorEnvironmentResponse, error) {
	if options.GOOS == "" {
		options.GOOS = runtime.GOOS
	}
	if options.GOARCH == "" {
		options.GOARCH = runtime.GOARCH
	}
	projectRoot, err := selectedProjectRoot(request.ProjectPath, options)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	effective, err := resolveDoctorConfiguration(request, options, projectRoot)
	if err != nil {
		return DoctorEnvironmentResponse{}, err
	}
	report := doctor.EnvironmentRegistry().Run(
		context.Background(), doctor.ScopeEnvironment,
		doctor.Input{
			GOOS: options.GOOS, GOARCH: options.GOARCH,
			Executables: map[string]string{
				"tofu":           effective.Settings.Executables.Tofu,
				"ansible-runner": effective.Settings.Executables.AnsibleRunner,
				"git":            effective.Settings.Executables.Git,
				"ssh":            effective.Settings.Executables.SSH,
			},
		},
		doctor.Capabilities{InspectExecutable: options.InspectExecutable},
	)
	result := output.Doctor{
		Scope: "environment",
		Summary: output.DoctorSummary{
			Pass: report.Summary.Pass, Skip: report.Summary.Skip,
			Warning: report.Summary.Warning, Fail: report.Summary.Fail,
		},
		Findings:               report.Findings,
		EffectiveConfiguration: renderConfiguration(effective),
	}
	applyDiscoveredExecutables(result.EffectiveConfiguration, report.Findings)
	return DoctorEnvironmentResponse{
		Result: result, Format: effective.Settings.UI.Format,
		OutputStyle: effective.Settings.UI.OutputStyle,
		Color:       effective.Settings.UI.Color,
	}, nil
}

func resolveDoctorConfiguration(
	request DoctorEnvironmentRequest,
	options DoctorEnvironmentOptions,
	projectRoot string,
) (config.Effective, error) {
	explicitConfig := request.ConfigPath
	if explicitConfig == "" {
		explicitConfig = options.Environment["AINFRA_CONFIG"]
	}
	if explicitConfig != "" && !filepath.IsAbs(explicitConfig) {
		explicitConfig = filepath.Join(options.WorkingDirectory, explicitConfig)
	}
	files, err := config.Files(config.LocationOptions{
		GOOS: options.GOOS, HomeDirectory: options.HomeDirectory,
		XDGConfigHome: options.XDGConfigHome, DeploymentRoot: projectRoot,
		ExplicitPath: explicitConfig,
	})
	if err != nil {
		return config.Effective{}, err
	}
	flags := config.Patch{UI: config.UIPatch{
		Format: request.Format, OutputStyle: request.OutputStyle, Color: request.Color,
	}}
	effective, err := config.Resolve(config.ResolveOptions{
		Defaults: config.Defaults(options.CacheDirectory, options.RunDirectory),
		Files:    files, Environment: options.Environment, Flags: flags,
		DiagnosticProject: true,
	})
	if err != nil {
		return config.Effective{}, err
	}
	return effective, nil
}

func applyDiscoveredExecutables(
	effective *output.EffectiveConfiguration,
	findings []diagnostic.Diagnostic,
) {
	if effective == nil {
		return
	}
	for _, finding := range findings {
		if finding.Status != "pass" || finding.Path == "" ||
			!strings.HasPrefix(finding.Check, "environment.executable.") {
			continue
		}
		name := strings.TrimPrefix(finding.Check, "environment.executable.")
		key := "executables." + name
		if name == "ansible-runner" {
			key = "executables.ansibleRunner"
		}
		value := effective.Values[key]
		if value.Source == string(config.LayerDefault) {
			value.DisplayValue = finding.Path
			value.Source = "discovered"
			value.OverriddenSources = []string{string(config.LayerDefault)}
			effective.Values[key] = value
		}
	}
}

func selectedProjectRoot(explicit string, options DoctorEnvironmentOptions) (string, error) {
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if explicit != "" && environmentPath != "" && explicit != environmentPath {
		return "", errors.New("--project conflicts with AINFRA_PROJECT")
	}
	target, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: options.WorkingDirectory,
		ProjectPath:      explicit, EnvironmentPath: environmentPath,
	})
	if err == nil {
		return target.Root, nil
	}
	if explicit != "" || environmentPath != "" {
		return "", err
	}
	return "", nil
}

func renderConfiguration(effective config.Effective) *output.EffectiveConfiguration {
	values := make(map[string]output.EffectiveConfigurationValue, len(effective.Entries))
	for _, entry := range effective.Entries {
		overridden := make([]string, len(entry.Overridden))
		for index, layer := range entry.Overridden {
			overridden[index] = string(layer)
		}
		values[entry.Key] = output.EffectiveConfigurationValue{
			DisplayValue: entry.Value, Source: string(entry.Source),
			OverriddenSources: overridden,
		}
	}
	files := make([]output.ConfigurationFile, len(effective.Files))
	for index, file := range effective.Files {
		files[index] = output.ConfigurationFile{
			Layer: string(file.Layer), Path: file.Path, Status: file.Status,
		}
	}
	rejected := make([]output.RejectedProjectSetting, len(effective.Rejected))
	projectPath := "ainfra.config.yaml"
	for _, file := range effective.Files {
		if file.Layer == config.LayerProject {
			projectPath = file.Path
		}
	}
	for index, key := range effective.Rejected {
		rejected[index] = output.RejectedProjectSetting{
			Key: key, Path: projectPath, Code: "AINFRA-E2107",
			Message: "project configuration cannot set " + key,
		}
	}
	sort.Slice(rejected, func(left, right int) bool {
		return strings.Compare(rejected[left].Key, rejected[right].Key) < 0
	})
	return &output.EffectiveConfiguration{
		Values: values, Files: files, RejectedProjectSettings: rejected,
	}
}

// HostDoctorEnvironmentOptions reads only the closed host facts needed by the
// environment doctor. The caller chooses when to invoke it, preserving static
// help's no-side-effect contract.
func HostDoctorEnvironmentOptions() (DoctorEnvironmentOptions, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return DoctorEnvironmentOptions{}, fmt.Errorf("read working directory: %w", err)
	}
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return DoctorEnvironmentOptions{}, fmt.Errorf("read home directory: %w", err)
	}
	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return DoctorEnvironmentOptions{}, fmt.Errorf("read cache directory: %w", err)
	}
	environment := supportedEnvironment()
	return DoctorEnvironmentOptions{
		WorkingDirectory: workingDirectory, HomeDirectory: homeDirectory,
		XDGConfigHome:     environment["XDG_CONFIG_HOME"],
		CacheDirectory:    filepath.Join(cacheDirectory, "ainfra"),
		RunDirectory:      filepath.Join(homeDirectory, ".local", "state", "ainfra", "runs"),
		Environment:       environment,
		InspectExecutable: inspectExecutable(workingDirectory),
	}, nil
}

func inspectExecutable(workingDirectory string) func(
	context.Context, string, string,
) (doctor.ExecutableFact, error) {
	return func(
		ctx context.Context, name, configuredPath string,
	) (doctor.ExecutableFact, error) {
		path := configuredPath
		if path == "" {
			var err error
			path, err = exec.LookPath(name)
			if err != nil {
				return doctor.ExecutableFact{}, fmt.Errorf("discover %s: %w", name, err)
			}
		}
		executable, err := security.ResolveExecutable(path)
		if err != nil {
			return doctor.ExecutableFact{}, fmt.Errorf("validate %s executable: %w", name, err)
		}
		environment, err := security.BuildEnvironment(nil, nil, nil)
		if err != nil {
			return doctor.ExecutableFact{}, err
		}
		versionContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		var version bytes.Buffer
		result, err := (childexec.Runner{}).Run(versionContext, childexec.Request{
			Executable: executable, Args: versionArguments(name),
			WorkingRoot: workingDirectory, WorkingDir: workingDirectory,
			Environment: environment,
			IO:          childexec.IOPolicy{Stdout: &version, Stderr: &version},
		})
		if err != nil {
			return doctor.ExecutableFact{}, fmt.Errorf("query %s version: %w", name, err)
		}
		if result.ExitCode != 0 || result.Cancelled {
			return doctor.ExecutableFact{}, fmt.Errorf("query %s version failed", name)
		}
		reported := strings.TrimSpace(version.String())
		if reported == "" {
			return doctor.ExecutableFact{}, fmt.Errorf("%s reported an empty version", name)
		}
		return doctor.ExecutableFact{Path: executable.Path(), Version: reported}, nil
	}
}

func versionArguments(name string) []string {
	switch name {
	case "tofu":
		return []string{"version"}
	case "ssh":
		return []string{"-V"}
	default:
		return []string{"--version"}
	}
}

func supportedEnvironment() map[string]string {
	names := []string{
		"AINFRA_CONFIG", "AINFRA_PROJECT", "AINFRA_FORMAT",
		"AINFRA_OUTPUT_STYLE", "AINFRA_COLOR", "NO_COLOR",
		"AINFRA_NON_INTERACTIVE", "AINFRA_LOG_LEVEL", "AINFRA_LOG_FORMAT",
		"AINFRA_LOG_FILE", "AINFRA_LOG_SYSLOG", "AINFRA_CACHE_DIR",
		"AINFRA_RUN_DIR", "AINFRA_TOFU_PATH", "AINFRA_ANSIBLE_RUNNER_PATH",
		"AINFRA_GIT_PATH", "AINFRA_SSH_PATH", "XDG_CONFIG_HOME",
	}
	values := make(map[string]string, len(names))
	for _, name := range names {
		if value, found := os.LookupEnv(name); found {
			values[name] = value
		}
	}
	return values
}
