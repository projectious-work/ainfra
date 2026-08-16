// Command ainfra provides the ainfra command-line interface.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/command"
	"github.com/projectious-work/ainfra/internal/config"
	"github.com/projectious-work/ainfra/internal/initialize"
	operational "github.com/projectious-work/ainfra/internal/logging"
	"github.com/projectious-work/ainfra/internal/mcpserver"
	"github.com/projectious-work/ainfra/internal/output"
)

const (
	developmentVersion = "0.0.0-dev"
	unknownCommit      = "unknown"
	reproducibleEpoch  = "1970-01-01T00:00:00Z"
)

var injectedVersion string

func main() {
	build := readBuild()
	logger, closeLogger, err := hostOperationalLogger(os.Args[1:], os.Stderr)
	if err != nil {
		_, _ = os.Stderr.WriteString("AINFRA-E0003: initialize operational logging: " + err.Error() + "\n")
		os.Exit(int(command.ExitOperationFailed))
	}
	code := command.Run(os.Args[1:], command.Options{
		Build:       build,
		Operational: &logger,
		Initialize:  initialize.Create,
		TemplateLock: func(request app.TemplateLockRequest) (output.Template, error) {
			options, err := app.HostTemplateLockOptions()
			if err != nil {
				return output.Template{}, err
			}
			return app.TemplateLock(request, options)
		},
		TemplateUpdate: func(request app.TemplateLockRequest) (output.Template, error) {
			options, err := app.HostTemplateLockOptions()
			if err != nil {
				return output.Template{}, err
			}
			return app.TemplateUpdate(request, options)
		},
		Plan: func(request app.PlanRequest) (output.Plan, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Plan{}, err
			}
			return app.Plan(context.Background(), request, options)
		},
		Apply: func(request app.ApplyRequest) (output.Execution, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Execution{}, err
			}
			return app.Apply(context.Background(), request, options)
		},
		Destroy: func(request app.DestroyRequest) (output.Execution, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Execution{}, err
			}
			return app.Destroy(context.Background(), request, options)
		},
		Logs: func(request app.EvidenceRequest) (output.Logs, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Logs{}, err
			}
			return app.Logs(request, options)
		},
		Status: func(request app.EvidenceRequest) (output.Status, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Status{}, err
			}
			return app.Status(request, options)
		},
		Output: func(request app.ArtifactRequest) (output.Artifact, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Artifact{}, err
			}
			return app.CollectOutput(context.Background(), request, options)
		},
		Inventory: func(request app.ArtifactRequest) (output.Artifact, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Artifact{}, err
			}
			return app.GenerateInventory(context.Background(), request, options)
		},
		Configure: func(request app.ConfigureRequest) (output.Execution, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Execution{}, err
			}
			return app.Configure(context.Background(), request, options)
		},
		Deploy: func(request app.DeployRequest) (output.Execution, error) {
			options, err := app.HostPlanOptions()
			if err != nil {
				return output.Execution{}, err
			}
			return app.Deploy(context.Background(), request, options)
		},
		MCPServe: func(ctx context.Context, request app.MCPServeRequest) error {
			return mcpserver.Serve(ctx, request, mcpserver.Options{Build: build,
				Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
		},
		DoctorEnvironment: func(
			request app.DoctorEnvironmentRequest,
		) (app.DoctorEnvironmentResponse, error) {
			options, err := app.HostDoctorEnvironmentOptions()
			if err != nil {
				return app.DoctorEnvironmentResponse{}, err
			}
			return app.DoctorEnvironment(request, options)
		},
		DoctorDeployment: func(
			request app.DoctorDeploymentRequest,
		) (app.DoctorEnvironmentResponse, error) {
			options, err := app.HostDoctorEnvironmentOptions()
			if err != nil {
				return app.DoctorEnvironmentResponse{}, err
			}
			return app.DoctorDeployment(request, options)
		},
		DoctorTemplate: func(
			request app.DoctorTemplateRequest,
		) (app.DoctorEnvironmentResponse, error) {
			options, err := app.HostDoctorEnvironmentOptions()
			if err != nil {
				return app.DoctorEnvironmentResponse{}, err
			}
			return app.DoctorTemplate(request, options)
		},
		DoctorRun: func(
			request app.DoctorRunRequest,
		) (app.DoctorEnvironmentResponse, error) {
			options, err := app.HostDoctorEnvironmentOptions()
			if err != nil {
				return app.DoctorEnvironmentResponse{}, err
			}
			return app.DoctorRun(request, options)
		},
		DoctorAll: func(
			request app.DoctorAllRequest,
		) (app.DoctorEnvironmentResponse, error) {
			options, err := app.HostDoctorEnvironmentOptions()
			if err != nil {
				return app.DoctorEnvironmentResponse{}, err
			}
			return app.DoctorAll(request, options)
		},
		IO: command.IO{
			Stdin:      os.Stdin,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			IsTerminal: isTerminal(os.Stdout),
		},
	})
	if err := closeLogger(); err != nil && code == command.ExitSuccess {
		_, _ = os.Stderr.WriteString("AINFRA-E0003: close operational logging: " + err.Error() + "\n")
		code = command.ExitOperationFailed
	}
	os.Exit(int(code))
}

func hostOperationalLogger(arguments []string, stderr io.Writer) (operational.Logger, func() error, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return operational.Logger{}, nil, err
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return operational.Logger{}, nil, err
	}
	explicit := os.Getenv("AINFRA_CONFIG")
	for index, argument := range arguments {
		if argument == "--config" && index+1 < len(arguments) {
			explicit = arguments[index+1]
		}
	}
	if explicit != "" && !filepath.IsAbs(explicit) {
		working, workingErr := os.Getwd()
		if workingErr != nil {
			return operational.Logger{}, nil, workingErr
		}
		explicit = filepath.Join(working, explicit)
	}
	files, err := config.Files(config.LocationOptions{GOOS: runtime.GOOS,
		HomeDirectory: home, XDGConfigHome: os.Getenv("XDG_CONFIG_HOME"), ExplicitPath: explicit})
	if err != nil {
		return operational.Logger{}, nil, err
	}
	environment := make(map[string]string)
	for _, entry := range os.Environ() {
		if name, value, found := strings.Cut(entry, "="); found {
			environment[name] = value
		}
	}
	if err := applyLoggingFlags(arguments, environment); err != nil {
		return operational.Logger{}, nil, err
	}
	effective, err := config.Resolve(config.ResolveOptions{Defaults: config.Defaults(
		filepath.Join(cache, "ainfra"), filepath.Join(cache, "ainfra", "runs")),
		Files: files, Environment: environment})
	if err != nil {
		return operational.Logger{}, nil, err
	}
	// Structured command output owns both process streams. Keep the default
	// operational stderr sink from contaminating its machine-readable contract;
	// explicitly configured file and syslog sinks remain active.
	if structuredOutput(arguments) {
		stderr = io.Discard
	}
	return operational.Build(effective.Settings.Logging, stderr)
}

func structuredOutput(arguments []string) bool {
	for index, argument := range arguments {
		if argument == "--format" && index+1 < len(arguments) {
			if arguments[index+1] == "json" || arguments[index+1] == "yaml" {
				return true
			}
			continue
		}
		if value, found := strings.CutPrefix(argument, "--format="); found {
			if value == "json" || value == "yaml" {
				return true
			}
		}
	}
	return false
}

func applyLoggingFlags(arguments []string, environment map[string]string) error {
	verbosity, explicitLevel := 0, ""
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		switch argument {
		case "-v":
			verbosity++
		case "-vv":
			verbosity += 2
		case "-vvv":
			verbosity += 3
		case "-vvvv":
			return errors.New("verbosity may not exceed -vvv")
		case "--syslog":
			environment["AINFRA_LOG_SYSLOG"] = "true"
		case "--log-level", "--log-format", "--log-file":
			if index+1 >= len(arguments) {
				return fmt.Errorf("%s requires a value", argument)
			}
			index++
			value := arguments[index]
			switch argument {
			case "--log-level":
				explicitLevel = value
			case "--log-format":
				environment["AINFRA_LOG_FORMAT"] = value
			case "--log-file":
				environment["AINFRA_LOG_FILE"] = value
			}
		}
	}
	if verbosity > 3 {
		return errors.New("verbosity may not exceed -vvv")
	}
	if verbosity > 0 && explicitLevel != "" {
		return errors.New("-v and --log-level are mutually exclusive")
	}
	if explicitLevel != "" {
		environment["AINFRA_LOG_LEVEL"] = explicitLevel
	} else if verbosity > 0 {
		environment["AINFRA_LOG_LEVEL"] = []string{"", "info", "debug", "trace"}[verbosity]
	}
	return nil
}

func readBuild() app.Build {
	build := app.Build{
		Version: developmentVersion,
		Commit:  unknownCommit,
		BuiltAt: reproducibleEpoch,
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return build
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		build.Version = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			build.Commit = setting.Value
		case "vcs.time":
			if _, err := time.Parse(time.RFC3339, setting.Value); err == nil {
				build.BuiltAt = setting.Value
			}
		}
	}
	if injectedVersion != "" {
		build.Version = injectedVersion
	}
	return build
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb"
}
