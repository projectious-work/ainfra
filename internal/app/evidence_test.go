package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/app"
)

func TestLogsFiltersTypedAinfraErrors(t *testing.T) {
	t.Parallel()
	deployment, runs, id := evidenceFixture(t)
	result, err := app.Logs(app.EvidenceRequest{Target: deployment, RunID: id,
		Errors: true}, evidenceOptions(t, runs))
	if err != nil {
		t.Fatal(err)
	}
	if result.StructuredFiltering != "available" || result.View != "errors" ||
		result.DisplayedRecords != 1 {
		t.Fatalf("logs=%+v", result)
	}
}

func TestLogsRefusesUnavailableChildEvidenceAndPublicFiles(t *testing.T) {
	t.Parallel()
	deployment, runs, id := evidenceFixture(t)
	options := evidenceOptions(t, runs)
	if _, err := app.Logs(app.EvidenceRequest{Target: deployment, RunID: id,
		Source: "opentofu"}, options); err == nil {
		t.Fatal("missing child evidence unexpectedly accepted")
	}
	if err := os.Chmod(filepath.Join(runs, id, "events.jsonl"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Logs(app.EvidenceRequest{Target: deployment, RunID: id}, options); err == nil {
		t.Fatal("public retained evidence unexpectedly accepted")
	}
}

func TestStatusReportsInterruptedRecoveryWithoutReadingState(t *testing.T) {
	t.Parallel()
	deployment, runs, _ := evidenceFixture(t)
	result, err := app.Status(app.EvidenceRequest{Target: deployment}, evidenceOptions(t, runs))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Runs) != 1 || result.Runs[0].ExecutionOutcome != "interrupted" ||
		result.Runs[0].Recovery == nil || !result.Runs[0].Recovery.InspectionRequired ||
		result.Runs[0].Recovery.AutomaticRetryAllowed {
		t.Fatalf("status=%+v", result)
	}
}

func TestRawLogsRequirePrivateExplicitChildStream(t *testing.T) {
	t.Parallel()
	deployment, runs, id := evidenceFixture(t)
	write(t, filepath.Join(runs, id, "opentofu.stderr"), "token=unredacted-test-value\n")
	result, err := app.Logs(app.EvidenceRequest{Target: deployment, RunID: id,
		Raw: true, Source: "opentofu", Stream: "stderr"}, evidenceOptions(t, runs))
	if err != nil || len(result.Records) != 1 ||
		result.Records[0] != "token=unredacted-test-value\n" || !result.Evidence[0].Sensitive {
		t.Fatalf("raw result=%+v err=%v", result, err)
	}
	if _, err := app.Logs(app.EvidenceRequest{Target: deployment, RunID: id,
		Raw: true, Source: "ainfra", Stream: "stderr"}, evidenceOptions(t, runs)); err == nil {
		t.Fatal("raw ainfra source unexpectedly accepted")
	}
}

func evidenceFixture(t *testing.T) (string, string, string) {
	t.Helper()
	deployment := runDeployment(t)
	runs, id := t.TempDir(), "20260814T170000Z-0123456789abcdef"
	root := filepath.Join(runs, id)
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "run.json"), `{"schemaVersion":1,"runId":"`+id+`","operation":"plan","state":"succeeded","createdAt":"2026-08-14T17:00:00Z","planRecord":"plan-record.json"}`)
	write(t, filepath.Join(root, "plan-record.json"), `{"schemaVersion":1,"runId":"`+id+`","intent":"apply","deployment":{"name":"run-example","digest":"sha256:x"},"template":{"source":"local:x","digest":"sha256:x"},"inputs":[],"engine":{"name":"opentofu","version":"1.10.0","executableDigest":"sha256:x"},"plan":{"path":"plan.tfplan","digest":"sha256:x","summaryPath":"plan.json"}}`)
	write(t, filepath.Join(root, "events.jsonl"),
		"{\"schemaVersion\":1,\"operation\":\"apply\",\"state\":\"started\",\"occurredAt\":\"2026-08-14T17:00:01Z\"}\n"+
			"{\"schemaVersion\":1,\"operation\":\"apply\",\"state\":\"inspection-required\",\"occurredAt\":\"2026-08-14T17:00:02Z\"}\n")
	return deployment, runs, id
}

func evidenceOptions(t *testing.T, runs string) app.PlanHostOptions {
	t.Helper()
	return app.PlanHostOptions{GOOS: "linux", WorkingDirectory: t.TempDir(),
		HomeDirectory: t.TempDir(), CacheDirectory: t.TempDir(), RunDirectory: runs,
		Environment: map[string]string{}}
}
