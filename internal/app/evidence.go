package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/projectious-work/ainfra/internal/ansible"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
)

var errStructuredEvidenceUnavailable = errors.New("retained structured evidence is unavailable")

// EvidenceRequest selects retained, ainfra-owned run evidence.
type EvidenceRequest struct {
	Target, ProjectPath, ConfigPath, RunID string
	Source                                 string
	Errors                                 bool
	Raw                                    bool
	Stream                                 string
}

func evidenceDeployment(request EvidenceRequest, options PlanHostOptions) (project.Deployment, string, error) {
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return project.Deployment{}, "", errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return project.Deployment{}, "", fmt.Errorf("load deployment contract: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{ConfigPath: request.ConfigPath}, deployment.Target.Root, options)
	if err != nil {
		return project.Deployment{}, "", fmt.Errorf("resolve evidence configuration: %w", err)
	}
	return deployment, settings.Paths.Runs, nil
}

// Logs reads the append-only ainfra event source without consulting state or
// operational log sinks.
func Logs(request EvidenceRequest, options PlanHostOptions) (output.Logs, error) {
	if request.RunID == "" {
		return output.Logs{}, errors.New("logs requires --run RUN_ID")
	}
	if request.Raw {
		return rawLogs(request, options)
	}
	if request.Source == "" {
		return combinedLogs(request, options)
	}
	if request.Source == "ansible-runner" {
		return ansibleLogs(request, options)
	}
	if request.Source == "opentofu" {
		return unavailableOpenTofuLogs(request, options)
	}
	if request.Source != "ainfra" {
		return output.Logs{}, errors.New("retained structured evidence for selected child source is unavailable")
	}
	deployment, runsRoot, err := evidenceDeployment(request, options)
	if err != nil {
		return output.Logs{}, err
	}
	root, err := security.ResolveContained(runsRoot, request.RunID)
	if err != nil {
		return output.Logs{}, errors.New("retained run ID is invalid")
	}
	contents, err := readPrivateArtifact(root, "events.jsonl", 4<<20)
	if err != nil {
		return output.Logs{}, fmt.Errorf("read retained events: %w", err)
	}
	records := make([]string, 0)
	for _, line := range bytes.Split(bytes.TrimSpace(contents), []byte{'\n'}) {
		var event runstate.ExecutionEvent
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&event) != nil || event.SchemaVersion != 1 {
			return output.Logs{}, errors.New("retained event stream is invalid")
		}
		if request.Errors && event.State != "failed" && event.State != "cancelled" &&
			event.State != "inspection-required" {
			continue
		}
		records = append(records, fmt.Sprintf("%s %s %s", event.OccurredAt, event.Operation, event.State))
	}
	view := "timeline"
	if request.Errors {
		view = "errors"
	}
	return output.Logs{Deployment: output.Deployment{Name: deployment.Metadata.Name, Root: deployment.Target.Root},
		RunID: request.RunID, View: view, Source: "ainfra", StructuredFiltering: "available",
		DisplayedRecords: len(records), Evidence: []output.Evidence{{Kind: "run-events", Path: "events.jsonl"}},
		Records: records}, nil
}

func unavailableOpenTofuLogs(request EvidenceRequest, options PlanHostOptions) (output.Logs, error) {
	deployment, runsRoot, err := evidenceDeployment(request, options)
	if err != nil {
		return output.Logs{}, err
	}
	root, err := security.ResolveContained(runsRoot, request.RunID)
	if err != nil {
		return output.Logs{}, errors.New("retained run ID is invalid")
	}
	evidence := make([]output.Evidence, 0, 2)
	for _, stream := range []string{"stdout", "stderr"} {
		name := "opentofu." + stream
		present, inspectErr := privateArtifactPresent(root, name)
		if inspectErr != nil {
			return output.Logs{}, fmt.Errorf("inspect retained OpenTofu stream: %w", inspectErr)
		}
		if present {
			evidence = append(evidence, output.Evidence{Kind: "raw-engine-stream",
				Engine: "opentofu", Path: name, Sensitive: true})
		}
	}
	view := "timeline"
	if request.Errors {
		view = "errors"
	}
	return output.Logs{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, RunID: request.RunID, View: view,
		Source: "opentofu", StructuredFiltering: "unavailable",
		DisplayedRecords: 0, Evidence: evidence, Records: []string{}}, nil
}

func privateArtifactPresent(root, name string) (bool, error) {
	directory, err := os.OpenRoot(root)
	if err != nil {
		return false, err
	}
	defer func() { _ = directory.Close() }()
	info, err := directory.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return false, errors.New("artifact is not a private regular file")
	}
	return true, nil
}

func combinedLogs(request EvidenceRequest, options PlanHostOptions) (output.Logs, error) {
	ainfraRequest := request
	ainfraRequest.Source = "ainfra"
	combined, err := Logs(ainfraRequest, options)
	if err != nil {
		return output.Logs{}, err
	}
	combined.Source = ""
	combined.StructuredFiltering = "partial"
	ansibleRequest := request
	ansibleRequest.Source = "ansible-runner"
	ansibleResult, ansibleErr := ansibleLogs(ansibleRequest, options)
	if ansibleErr == nil {
		combined.Records = append(combined.Records, ansibleResult.Records...)
		combined.Evidence = append(combined.Evidence, ansibleResult.Evidence...)
	} else if !errors.Is(ansibleErr, errStructuredEvidenceUnavailable) {
		return output.Logs{}, ansibleErr
	}
	sort.Strings(combined.Records)
	combined.DisplayedRecords = len(combined.Records)
	return combined, nil
}

func ansibleLogs(request EvidenceRequest, options PlanHostOptions) (output.Logs, error) {
	deployment, runsRoot, err := evidenceDeployment(request, options)
	if err != nil {
		return output.Logs{}, err
	}
	root, err := security.ResolveContained(runsRoot, request.RunID)
	if err != nil {
		return output.Logs{}, errors.New("retained run ID is invalid")
	}
	records := make([]string, 0)
	evidence := make([]output.Evidence, 0)
	for _, operation := range []string{"configure", "configure-check"} {
		privateData := filepath.Join("ansible-runner", operation)
		versionBytes, readErr := readPrivateArtifact(root, filepath.Join(privateData, "version"), 1024)
		if errors.Is(readErr, os.ErrNotExist) {
			continue
		}
		if readErr != nil {
			return output.Logs{}, fmt.Errorf("read retained Ansible Runner version: %w", readErr)
		}
		artifactDir := filepath.Join(privateData, "artifacts")
		events, readErr := ansible.ReadEvents(root, artifactDir, strings.TrimSpace(string(versionBytes)))
		if readErr != nil {
			return output.Logs{}, readErr
		}
		for _, event := range events {
			if request.Errors && !event.Failure {
				continue
			}
			records = append(records, fmt.Sprintf("%s ansible-runner %s %s",
				event.Created, event.Event, event.UUID))
		}
		evidence = append(evidence, output.Evidence{Kind: "runner-artifacts",
			Engine: "ansible-runner", Path: filepath.ToSlash(artifactDir), Sensitive: true})
	}
	if len(evidence) == 0 {
		return output.Logs{}, fmt.Errorf("retained Ansible Runner: %w", errStructuredEvidenceUnavailable)
	}
	view := "timeline"
	if request.Errors {
		view = "errors"
	}
	return output.Logs{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, RunID: request.RunID, View: view,
		Source: "ansible-runner", StructuredFiltering: "available",
		DisplayedRecords: len(records), Evidence: evidence, Records: records}, nil
}

func rawLogs(request EvidenceRequest, options PlanHostOptions) (output.Logs, error) {
	if request.Errors {
		return output.Logs{}, errors.New("--raw and --errors are mutually exclusive")
	}
	if request.Source != "opentofu" && request.Source != "ansible-runner" {
		return output.Logs{}, errors.New("--raw requires an explicit child-engine --source")
	}
	if request.Stream != "stdout" && request.Stream != "stderr" && request.Stream != "events" {
		return output.Logs{}, errors.New("--raw requires --stream stdout, stderr, or events")
	}
	deployment, runsRoot, err := evidenceDeployment(request, options)
	if err != nil {
		return output.Logs{}, err
	}
	root, err := security.ResolveContained(runsRoot, request.RunID)
	if err != nil {
		return output.Logs{}, errors.New("retained run ID is invalid")
	}
	name := request.Source + "." + request.Stream
	contents, err := readPrivateArtifact(root, name, 64<<20)
	if err != nil {
		return output.Logs{}, fmt.Errorf("read retained raw stream: %w", err)
	}
	return output.Logs{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, RunID: request.RunID, View: "timeline",
		Source: request.Source, StructuredFiltering: "unavailable",
		DisplayedRecords: 1, Evidence: []output.Evidence{{Kind: "raw-engine-stream",
			Engine: request.Source, Path: name, Sensitive: true}}, Records: []string{string(contents)}}, nil
}

// Status derives lifecycle outcomes only from strict ainfra-owned records.
func Status(request EvidenceRequest, options PlanHostOptions) (output.Status, error) {
	deployment, runsRoot, err := evidenceDeployment(request, options)
	if err != nil {
		return output.Status{}, err
	}
	return statusForDeployment(deployment, runsRoot)
}

func statusForDeployment(deployment project.Deployment, runsRoot string) (output.Status, error) {
	entries, err := os.ReadDir(runsRoot)
	if errors.Is(err, os.ErrNotExist) {
		return output.Status{Deployment: output.Deployment{Name: deployment.Metadata.Name,
			Root: deployment.Target.Root}, Runs: []output.RunStatus{}}, nil
	}
	if err != nil {
		return output.Status{}, fmt.Errorf("read retained runs: %w", err)
	}
	runs := make([]output.RunStatus, 0)
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		root, resolveErr := security.ResolveContained(runsRoot, entry.Name())
		if resolveErr != nil {
			continue
		}
		var record runstate.RunRecord
		contents, readErr := readPrivateArtifact(root, "run.json", 1<<20)
		if readErr != nil || json.Unmarshal(contents, &record) != nil || record.SchemaVersion != 1 {
			continue
		}
		var plan runstate.PlanRecord
		planBytes, readErr := readPrivateArtifact(root, "plan-record.json", 1<<20)
		if readErr != nil || json.Unmarshal(planBytes, &plan) != nil ||
			plan.Deployment.Name != deployment.Metadata.Name {
			continue
		}
		status := output.RunStatus{RunID: record.RunID, Operation: record.Operation,
			ExecutionOutcome: record.State, StartedAt: record.CreatedAt}
		if events, readErr := readPrivateArtifact(root, "events.jsonl", 4<<20); readErr == nil {
			for _, line := range bytes.Split(bytes.TrimSpace(events), []byte{'\n'}) {
				var event runstate.ExecutionEvent
				if json.Unmarshal(line, &event) != nil {
					continue
				}
				status.Operation, status.ExecutionOutcome = event.Operation, event.State
				if event.State != "started" {
					status.FinishedAt = event.OccurredAt
				}
				if event.State == "cancelled" || event.State == "inspection-required" {
					status.ExecutionOutcome = "interrupted"
					status.Recovery = &output.Recovery{AutomaticRetryAllowed: false,
						InspectionRequired: true, NextCommands: []string{"ainfra doctor run " + deployment.Metadata.Name}}
				}
			}
		}
		runs = append(runs, status)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt > runs[j].StartedAt })
	return output.Status{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, Runs: runs}, nil
}
