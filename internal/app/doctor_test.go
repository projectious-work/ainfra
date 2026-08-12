package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestDoctorEnvironmentUsesProjectConfigSafely(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	deployment := t.TempDir()
	write(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
`)
	write(t, filepath.Join(deployment, "ainfra.config.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
ui:
  outputStyle: plain
paths:
  cache: prohibited
`)
	response, err := app.DoctorEnvironment(
		app.DoctorEnvironmentRequest{ProjectPath: deployment},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
			HomeDirectory: home, CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{},
		},
	)
	if err != nil {
		t.Fatalf("doctor environment: %v", err)
	}
	configuration := response.Result.EffectiveConfiguration
	if response.OutputStyle != "plain" || configuration == nil ||
		len(configuration.RejectedProjectSettings) != 1 ||
		configuration.RejectedProjectSettings[0].Key != "paths.cache" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if configuration.Values["paths.cache"].DisplayValue != "/cache" {
		t.Fatal("prohibited project path was applied")
	}
}

func TestDoctorEnvironmentRejectsConflictingProjectSelectors(t *testing.T) {
	t.Parallel()
	_, err := app.DoctorEnvironment(
		app.DoctorEnvironmentRequest{ProjectPath: "/flag"},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "amd64", WorkingDirectory: t.TempDir(),
			HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{"AINFRA_PROJECT": "/environment"},
		},
	)
	if err == nil {
		t.Fatal("conflicting project selectors unexpectedly accepted")
	}
}

func write(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
