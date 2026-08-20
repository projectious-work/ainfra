package app_test

import (
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestDoctorTemplateReportsLocalContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../spec/examples/v1/template-example")
	if err != nil {
		t.Fatal(err)
	}
	response, err := app.DoctorTemplate(
		app.DoctorTemplateRequest{Target: root},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
			HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{},
		},
	)
	if err != nil {
		t.Fatalf("doctor template: %v", err)
	}
	if response.Result.Scope != "template" || response.Result.Summary.Pass != 7 ||
		response.Result.Summary.Skip != 1 || len(response.Result.Findings) != 8 {
		t.Fatalf("response = %#v", response)
	}
}

func TestDoctorTemplateReportsCleanRoomAuthoringProof(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../spec/examples/v1/template-clean-room")
	if err != nil {
		t.Fatal(err)
	}
	response, err := app.DoctorTemplate(
		app.DoctorTemplateRequest{Target: root},
		app.DoctorEnvironmentOptions{
			GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
			HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
			Environment: map[string]string{},
		},
	)
	if err != nil {
		t.Fatalf("doctor template: %v", err)
	}
	if response.Result.Summary.Pass != 7 || response.Result.Summary.Skip != 1 ||
		response.Result.Summary.Fail != 0 || len(response.Result.Findings) != 8 {
		t.Fatalf("response = %#v", response)
	}
}
