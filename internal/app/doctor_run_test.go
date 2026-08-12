package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestDoctorRunSkipsWhenNoEvidenceIsRetained(t *testing.T) {
	t.Parallel()
	root := runDeployment(t)
	response, err := app.DoctorRun(
		app.DoctorRunRequest{Target: root}, runDoctorOptions(t),
	)
	if err != nil {
		t.Fatalf("doctor run: %v", err)
	}
	if response.Result.Summary.Skip != 1 ||
		response.Result.Findings[0].Status != "skip" {
		t.Fatalf("unexpected response: %#v", response.Result)
	}
}

func TestDoctorRunValidatesLatestRetainedEvidence(t *testing.T) {
	t.Parallel()
	root := runDeployment(t)
	runRoot := filepath.Join(root, ".ainfra", "runs", "20260812T120000Z-example")
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(runRoot, "run.json"), "{}\n")
	write(t, filepath.Join(runRoot, "events.jsonl"), "{}\n")
	if err := os.Chmod(filepath.Join(runRoot, "run.json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(runRoot, "events.jsonl"), 0o600); err != nil {
		t.Fatal(err)
	}
	response, err := app.DoctorRun(
		app.DoctorRunRequest{Target: root}, runDoctorOptions(t),
	)
	if err != nil {
		t.Fatalf("doctor run: %v", err)
	}
	if response.Result.Summary.Pass != 1 ||
		response.Result.Findings[0].Status != "pass" {
		t.Fatalf("unexpected response: %#v", response.Result)
	}
}

func TestDoctorRunRejectsNonOwnerOnlyEvidence(t *testing.T) {
	t.Parallel()
	root := runDeployment(t)
	runRoot := filepath.Join(root, ".ainfra", "runs", "20260812T120000Z-example")
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(runRoot, "run.json"), "{}\n")
	write(t, filepath.Join(runRoot, "events.jsonl"), "{}\n")
	if err := os.Chmod(filepath.Join(runRoot, "run.json"), 0o644); err != nil {
		t.Fatal(err)
	}
	response, err := app.DoctorRun(
		app.DoctorRunRequest{Target: root}, runDoctorOptions(t),
	)
	if err != nil {
		t.Fatalf("doctor run: %v", err)
	}
	if response.Result.Summary.Fail != 1 ||
		response.Result.Findings[0].Status != "fail" {
		t.Fatalf("unexpected response: %#v", response.Result)
	}
}

func TestDoctorRunFailsIncompleteLatestEvidence(t *testing.T) {
	t.Parallel()
	root := runDeployment(t)
	runRoot := filepath.Join(root, ".ainfra", "runs", "20260812T120000Z-example")
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(runRoot, "run.json"), "{}\n")
	response, err := app.DoctorRun(
		app.DoctorRunRequest{Target: root}, runDoctorOptions(t),
	)
	if err != nil {
		t.Fatalf("doctor run: %v", err)
	}
	if response.Result.Summary.Fail != 1 ||
		response.Result.Findings[0].Status != "fail" {
		t.Fatalf("unexpected response: %#v", response.Result)
	}
}

func runDeployment(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: run-example
spec:
  template:
    source: local:../template
`)
	return root
}

func runDoctorOptions(t *testing.T) app.DoctorEnvironmentOptions {
	t.Helper()
	return app.DoctorEnvironmentOptions{
		GOOS: "linux", GOARCH: "arm64", WorkingDirectory: t.TempDir(),
		HomeDirectory: t.TempDir(), CacheDirectory: "/cache", RunDirectory: "/runs",
		Environment: map[string]string{},
	}
}
