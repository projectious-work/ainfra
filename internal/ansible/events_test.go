package ansible_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/ansible"
)

func TestReadEventsClassifiesOnlyNativeFailureTypes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "artifacts", "job", "job_events")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	writeEvent(t, filepath.Join(directory, "1.json"),
		`{"created":"2026-08-14T18:00:00Z","uuid":"one","event":"runner_on_ok","stdout":"ERROR prose","event_data":{"secret":"ignored"}}`)
	writeEvent(t, filepath.Join(directory, "2.json"),
		`{"created":"2026-08-14T18:00:01Z","uuid":"two","event":"runner_on_unreachable","unknown":true}`)
	records, err := ansible.ReadEvents(root, "artifacts", "2.4.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Failure || !records[1].Failure ||
		records[1].Event != "runner_on_unreachable" {
		t.Fatalf("records=%+v", records)
	}
}

func TestReadEventsFailsClosedOnVersionAndUnsafeArtifacts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := ansible.ReadEvents(root, "artifacts", "3.0.0"); err == nil {
		t.Fatal("unsupported Runner major unexpectedly accepted")
	}
	target := filepath.Join(root, "target.json")
	writeEvent(t, target, `{"event":"runner_on_failed"}`)
	if err := os.Symlink(target, filepath.Join(root, "artifacts", "event.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ansible.ReadEvents(root, "artifacts", "2.4.1"); err == nil {
		t.Fatal("symlinked event unexpectedly accepted")
	}
}

func writeEvent(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
