package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
)

// ReviewOptions supplies current trusted values for immediate reverification.
type ReviewOptions struct {
	ID            string
	RunsRoot      string
	CacheRoot     string
	Deployment    project.Deployment
	Lock          lockfile.Document
	Executable    security.Executable
	EngineVersion string
}

// Reviewed is an exact saved plan whose immutable bindings were reverified.
type Reviewed struct {
	Root   string
	Record PlanRecord
}

// LoadReviewed strictly loads and immediately reverifies an apply plan.
func LoadReviewed(options ReviewOptions) (Reviewed, error) {
	return LoadReviewedIntent(options, "apply")
}

// LoadReviewedIntent strictly loads and immediately reverifies a plan whose
// recorded intent exactly matches the requested mutating operation.
func LoadReviewedIntent(options ReviewOptions, intent string) (Reviewed, error) {
	if intent != "apply" && intent != "destroy" {
		return Reviewed{}, errors.New("unsupported reviewed plan intent")
	}
	if !validID(options.ID) {
		return Reviewed{}, errors.New("invalid reviewed plan ID")
	}
	root, err := security.ResolveContained(options.RunsRoot, options.ID)
	if err != nil {
		return Reviewed{}, fmt.Errorf("resolve reviewed plan: %w", err)
	}
	var record PlanRecord
	if err := readStrictJSON(root, "plan-record.json", &record); err != nil {
		return Reviewed{}, fmt.Errorf("read reviewed plan record: %w", err)
	}
	if record.SchemaVersion != 1 || record.RunID != options.ID || record.Intent != intent {
		return Reviewed{}, fmt.Errorf("reviewed plan does not authorize %s", intent)
	}
	if record.Deployment.Name != options.Deployment.Metadata.Name {
		return Reviewed{}, errors.New("reviewed plan deployment binding is stale")
	}
	manifestDigest, err := digestFile(options.Deployment.Target.ManifestPath)
	if err != nil || manifestDigest != record.Deployment.Digest {
		return Reviewed{}, errors.New("reviewed plan deployment bytes changed")
	}
	paths := append([]string(nil), options.Deployment.Inputs.TofuBackendConfigFiles...)
	paths = append(paths, options.Deployment.Inputs.TofuVariableFiles...)
	paths = append(paths, options.Deployment.Inputs.AnsibleVariableFiles...)
	if options.Deployment.SSH.KnownHosts != "" {
		paths = append(paths, options.Deployment.SSH.KnownHosts)
	}
	if len(paths) != len(record.Inputs) {
		return Reviewed{}, errors.New("reviewed plan input binding set changed")
	}
	for index, relative := range paths {
		path, resolveErr := security.ResolveContained(options.Deployment.Target.Root, relative)
		if resolveErr != nil {
			return Reviewed{}, fmt.Errorf("reviewed plan input %q changed", relative)
		}
		digest, digestErr := digestFile(path)
		if digestErr != nil || record.Inputs[index].Path != filepath.ToSlash(relative) ||
			record.Inputs[index].Digest != digest {
			return Reviewed{}, fmt.Errorf("reviewed plan input %q changed", relative)
		}
	}
	if record.Template.Source != options.Lock.Template.Source ||
		record.Template.Digest != options.Lock.Template.Digest ||
		(record.Template.Commit != "" && record.Template.Commit != gitCommit(options.Lock.Template.Resolved)) {
		return Reviewed{}, errors.New("reviewed plan template lock changed")
	}
	cacheKey := strings.TrimPrefix(record.Template.Digest, "sha256:")
	if len(cacheKey) != 64 || cacheKey == record.Template.Digest {
		return Reviewed{}, errors.New("reviewed plan template digest is invalid")
	}
	cachePath := filepath.Join(options.CacheRoot, "templates", "sha256", cacheKey)
	if digest, digestErr := source.TreeDigest(cachePath); digestErr != nil || digest != record.Template.Digest {
		return Reviewed{}, errors.New("reviewed plan template cache binding changed")
	}
	workspace := filepath.Join(root, "workspace")
	if digest, digestErr := source.EngineWorkspaceDigest(workspace, cachePath); digestErr != nil ||
		digest != record.Template.Digest {
		return Reviewed{}, errors.New("reviewed plan workspace template bytes changed")
	}
	if err := options.Executable.VerifyUnchanged(); err != nil ||
		options.Executable.Digest() != record.Engine.ExecutableDigest ||
		options.EngineVersion != record.Engine.Version {
		return Reviewed{}, errors.New("reviewed plan OpenTofu executable changed")
	}
	if record.Engine.Name != "opentofu" || record.Plan.Path != "plan.tfplan" ||
		record.Plan.SummaryPath != "plan.json" {
		return Reviewed{}, errors.New("reviewed plan record has unsupported bindings")
	}
	planPath, err := security.ResolveContained(root, record.Plan.Path)
	if err != nil {
		return Reviewed{}, errors.New("reviewed saved plan is unavailable")
	}
	if digest, digestErr := digestFile(planPath); digestErr != nil || digest != record.Plan.Digest {
		return Reviewed{}, errors.New("reviewed saved plan bytes changed")
	}
	if summary, resolveErr := security.ResolveContained(root, record.Plan.SummaryPath); resolveErr != nil || security.RequireRegular(summary) != nil {
		return Reviewed{}, errors.New("reviewed plan summary is unavailable")
	}
	return Reviewed{Root: root, Record: record}, nil
}

// ExecutionEvent is append-only evidence written around mutating execution.
type ExecutionEvent struct {
	SchemaVersion int    `json:"schemaVersion"`
	Operation     string `json:"operation"`
	State         string `json:"state"`
	OccurredAt    string `json:"occurredAt"`
	ExitCode      *int   `json:"exitCode,omitempty"`
}

// AppendExecutionEvent durably appends one private lifecycle event.
func AppendExecutionEvent(reviewed Reviewed, state string, at time.Time, exitCode *int) error {
	return appendOperationEvent(reviewed, "apply", state, at, exitCode, state == "started")
}

// AppendDestroyEvent records an exact reviewed destroy execution boundary.
func AppendDestroyEvent(reviewed Reviewed, state string, at time.Time, exitCode *int) error {
	return appendOperationEvent(reviewed, "destroy", state, at, exitCode, state == "started")
}

// BeginOperation durably starts a post-apply stage exactly once.
func BeginOperation(reviewed Reviewed, operation string, at time.Time) error {
	if operation != "output" && operation != "inventory" && operation != "configure" && operation != "configure-check" {
		return errors.New("invalid post-apply operation")
	}
	directory, err := os.OpenRoot(reviewed.Root)
	if err != nil {
		return err
	}
	contents, readErr := directory.ReadFile("events.jsonl")
	closeErr := directory.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return err
	}
	for _, line := range bytes.Split(bytes.TrimSpace(contents), []byte{'\n'}) {
		var event ExecutionEvent
		if (operation == "configure" || operation == "configure-check") &&
			json.Unmarshal(line, &event) == nil && event.Operation == operation && event.State == "started" {
			return fmt.Errorf("%s already started for this run", operation)
		}
	}
	return appendOperationEvent(reviewed, operation, "started", at, nil, false)
}

// AppendOperationEvent records a terminal post-apply stage event.
func AppendOperationEvent(reviewed Reviewed, operation, state string, at time.Time, exitCode *int) error {
	return appendOperationEvent(reviewed, operation, state, at, exitCode, false)
}

func appendOperationEvent(reviewed Reviewed, operation, state string, at time.Time, exitCode *int, exclusive bool) error {
	if state != "started" && state != "succeeded" && state != "failed" &&
		state != "cancelled" && state != "inspection-required" {
		return errors.New("invalid execution event state")
	}
	contents, err := json.Marshal(ExecutionEvent{SchemaVersion: 1, Operation: operation,
		State: state, OccurredAt: at.UTC().Format(time.RFC3339Nano), ExitCode: exitCode})
	if err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_APPEND
	if exclusive {
		flags |= os.O_CREATE | os.O_EXCL
	}
	root, err := os.OpenRoot(reviewed.Root)
	if err != nil {
		return fmt.Errorf("open reviewed run: %w", err)
	}
	file, err := root.OpenFile("events.jsonl", flags, 0o600)
	rootCloseErr := root.Close()
	if err != nil {
		return fmt.Errorf("open execution evidence: %w", errors.Join(err, rootCloseErr))
	}
	if rootCloseErr != nil {
		_ = file.Close()
		return rootCloseErr
	}
	_, writeErr := file.Write(append(contents, '\n'))
	return errors.Join(writeErr, file.Sync(), file.Close())
}

// RequireSuccessfulApply refuses post-apply stages unless durable evidence
// proves that the exact reviewed plan completed successfully.
func RequireSuccessfulApply(reviewed Reviewed) error {
	directory, err := os.OpenRoot(reviewed.Root)
	if err != nil {
		return err
	}
	contents, readErr := directory.ReadFile("events.jsonl")
	closeErr := directory.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return errors.New("successful apply evidence is unavailable")
	}
	if len(contents) > 4<<20 {
		return errors.New("apply evidence exceeds size limit")
	}
	lines := bytes.Split(bytes.TrimSpace(contents), []byte{'\n'})
	started, succeeded := false, false
	for _, line := range lines {
		var event ExecutionEvent
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&event); err != nil || event.SchemaVersion != 1 {
			return errors.New("apply evidence is invalid")
		}
		if event.Operation != "apply" {
			continue
		}
		switch event.State {
		case "started":
			started = true
		case "succeeded":
			succeeded = started
		case "failed", "cancelled", "inspection-required":
			succeeded = false
		}
	}
	if !succeeded {
		return errors.New("reviewed plan has no successful apply evidence")
	}
	return nil
}

func readStrictJSON(root, name string, target any) error {
	directory, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	contents, readErr := directory.ReadFile(name)
	closeErr := directory.Close()
	err = errors.Join(readErr, closeErr)
	if err != nil {
		return err
	}
	if len(contents) > 1<<20 {
		return errors.New("record exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("record must contain exactly one JSON value")
	}
	return nil
}
