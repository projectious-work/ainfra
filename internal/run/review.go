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
	if record.SchemaVersion != 1 || record.RunID != options.ID || record.Intent != "apply" {
		return Reviewed{}, errors.New("reviewed plan does not authorize apply")
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
	if state != "started" && state != "succeeded" && state != "failed" &&
		state != "cancelled" && state != "inspection-required" {
		return errors.New("invalid execution event state")
	}
	contents, err := json.Marshal(ExecutionEvent{SchemaVersion: 1, Operation: "apply",
		State: state, OccurredAt: at.UTC().Format(time.RFC3339Nano), ExitCode: exitCode})
	if err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_APPEND
	if state == "started" {
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
