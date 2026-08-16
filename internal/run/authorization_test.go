package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runstate "github.com/projectious-work/ainfra/internal/run"
)

func TestAuthorizationBindingAndEvidenceAreClosedAndExclusive(t *testing.T) {
	t.Parallel()
	runsRoot := t.TempDir()
	const runID = "20260816T120000Z-0123456789abcdef"
	runRoot := filepath.Join(runsRoot, runID)
	if err := os.Mkdir(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	digest := "sha256:" + strings.Repeat("a", 64)
	record := `{"schemaVersion":1,"runId":"` + runID +
		`","intent":"apply","deployment":{"name":"example","digest":"sha256:x"},` +
		`"template":{"source":"local:x","digest":"sha256:x"},"inputs":[],` +
		`"engine":{"name":"opentofu","version":"1","executableDigest":"sha256:x"},` +
		`"plan":{"path":"plan.tfplan","digest":"` + digest +
		`","summaryPath":"plan.json"}}`
	if err := os.WriteFile(filepath.Join(runRoot, "plan-record.json"),
		[]byte(record), 0o600); err != nil {
		t.Fatal(err)
	}
	binding, err := runstate.LoadAuthorizationBinding(runsRoot, runID, "example", "apply")
	if err != nil || binding.PlanID != runID || binding.PlanDigest != digest {
		t.Fatalf("authorization binding=%+v err=%v", binding, err)
	}
	if _, err := runstate.LoadAuthorizationBinding(runsRoot, runID, "other", "apply"); err == nil {
		t.Fatal("cross-deployment authorization binding succeeded")
	}
	evidence := runstate.NewAuthorizationRecord("approval-1", "apply", runID, digest,
		"agent-1", "operator-1", "2026-08-16T12:05:00Z",
		time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC))
	if err := runstate.RecordAuthorization(runsRoot, evidence); err != nil {
		t.Fatal(err)
	}
	if err := runstate.RecordAuthorization(runsRoot, evidence); err == nil {
		t.Fatal("authorization evidence overwrite succeeded")
	}
	info, err := os.Stat(filepath.Join(runRoot, "authorization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("authorization evidence mode=%v", info.Mode())
	}
	configureEvidence := runstate.NewAuthorizationRecord("approval-2", "configure", runID,
		digest, "agent-1", "operator-1", "2026-08-16T12:05:00Z",
		time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC))
	if err := runstate.RecordOperationAuthorization(runsRoot,
		"authorization-configure.json", configureEvidence); err != nil {
		t.Fatal(err)
	}
	if err := runstate.RecordOperationAuthorization(runsRoot,
		"authorization-configure.json", configureEvidence); err == nil {
		t.Fatal("configure authorization evidence overwrite succeeded")
	}
	if err := runstate.RecordOperationAuthorization(runsRoot,
		"../authorization.json", configureEvidence); err == nil {
		t.Fatal("arbitrary authorization evidence name succeeded")
	}
}
