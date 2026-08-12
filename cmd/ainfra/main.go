// Command ainfra provides the ainfra command-line interface.
package main

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/command"
	"github.com/projectious-work/ainfra/internal/initialize"
)

const (
	developmentVersion = "0.0.0-dev"
	unknownCommit      = "unknown"
	reproducibleEpoch  = "1970-01-01T00:00:00Z"
)

var injectedVersion string

func main() {
	build := readBuild()
	code := command.Run(os.Args[1:], command.Options{
		Build:      build,
		Initialize: initialize.Create,
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
	os.Exit(int(code))
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
