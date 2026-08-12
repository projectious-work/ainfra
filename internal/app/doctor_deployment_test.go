package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestDoctorDeploymentReportsValidatedContract(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "terraform.tfvars"), "opaque native bytes")
	write(t, filepath.Join(root, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
  inputs:
    tofu:
      variableFiles: [terraform.tfvars]
`)
	response, err := app.DoctorDeployment(
		app.DoctorDeploymentRequest{Target: root},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
			HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{},
		},
	)
	if err != nil {
		t.Fatalf("doctor deployment: %v", err)
	}
	if response.Result.Scope != "deployment" || response.Result.Summary.Pass != 3 ||
		len(response.Result.Findings) != 3 ||
		response.Result.EffectiveConfiguration != nil {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestDoctorDeploymentRejectsSelectorConflictBeforeLoading(t *testing.T) {
	t.Parallel()
	_, err := app.DoctorDeployment(
		app.DoctorDeploymentRequest{Target: "/target", ProjectPath: "/project"},
		app.DoctorEnvironmentOptions{Environment: map[string]string{}},
	)
	if err == nil {
		t.Fatal("conflicting selectors unexpectedly accepted")
	}
}

func TestDoctorDeploymentRejectsInvalidNativePointer(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ainfra.yaml"), []byte(`apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
  inputs:
    tofu:
      variableFiles: [../outside.tfvars]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := app.DoctorDeployment(
		app.DoctorDeploymentRequest{Target: root},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
			HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{},
		},
	)
	if err == nil {
		t.Fatal("traversing native pointer unexpectedly accepted")
	}
}
