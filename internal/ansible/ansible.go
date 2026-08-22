// Package ansible owns controlled Ansible Runner invocation and normalized
// outcome evaluation.
package ansible

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/security"
)

type Run func(context.Context, childexec.Request) (childexec.Result, error)

// Adapter binds Runner to an immutable executable and explicit environment.
type Adapter struct {
	Executable  security.Executable
	Environment security.Environment
	Run         Run
}

type Outcome struct {
	Result childexec.Result
	Stats  Stats
}

// Stats is the attributed playbook_on_stats evidence used for verification.
type Stats struct {
	Changed   map[string]int `json:"changed"`
	Dark      map[string]int `json:"dark"`
	Failures  map[string]int `json:"failures"`
	OK        map[string]int `json:"ok"`
	Processed map[string]int `json:"processed"`
	Skipped   map[string]int `json:"skipped"`
}

// EventRecord is a display-safe attribution selected from one native Runner
// job event. It deliberately excludes stdout and event_data values.
type EventRecord struct {
	Created string
	UUID    string
	Event   string
	Failure bool
}

// ReadEvents validates and classifies native Ansible Runner v2 job events.
func ReadEvents(root, artifactDir, version string) ([]EventRecord, error) {
	major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	if err != nil || major != 2 {
		return nil, fmt.Errorf("unsupported Ansible Runner evidence version %q", version)
	}
	directory, err := security.ResolveContained(root, artifactDir)
	if err != nil {
		return nil, err
	}
	artifactRoot, err := os.OpenRoot(directory)
	if err != nil {
		return nil, fmt.Errorf("open Ansible artifacts: %w", err)
	}
	defer func() { _ = artifactRoot.Close() }()
	records := make([]EventRecord, 0)
	files := 0
	var visit func(string, int) error
	visit = func(relative string, depth int) error {
		if depth > 8 {
			return errors.New("ansible artifact tree exceeds depth limit")
		}
		handle, err := artifactRoot.Open(relative)
		if err != nil {
			return err
		}
		entries, readErr := handle.ReadDir(-1)
		closeErr := handle.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("symlink in Ansible artifacts")
			}
			path := filepath.Join(relative, entry.Name())
			if entry.IsDir() {
				if err := visit(path, depth+1); err != nil {
					return err
				}
				continue
			}
			if filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			files++
			if files > 100000 {
				return errors.New("too many Ansible event artifacts")
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > 4<<20 {
				return errors.New("invalid Ansible event artifact")
			}
			contents, err := artifactRoot.ReadFile(path)
			if err != nil {
				return errors.New("invalid Ansible event artifact")
			}
			var event struct {
				Created string `json:"created"`
				UUID    string `json:"uuid"`
				Event   string `json:"event"`
			}
			decoder := json.NewDecoder(bytes.NewReader(contents))
			if decoder.Decode(&event) != nil || event.Event == "" {
				return errors.New("invalid Ansible event artifact")
			}
			failure := event.Event == "runner_on_failed" || event.Event == "runner_on_unreachable" ||
				event.Event == "runner_on_async_failed" || event.Event == "error"
			records = append(records, EventRecord{Created: event.Created, UUID: event.UUID,
				Event: event.Event, Failure: failure})
		}
		return nil
	}
	if err := visit(".", 0); err != nil {
		return nil, fmt.Errorf("read Ansible events: %w", err)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Created == records[j].Created {
			return records[i].UUID < records[j].UUID
		}
		return records[i].Created < records[j].Created
	})
	return records, nil
}

func (adapter Adapter) Version(ctx context.Context, root string) (string, error) {
	var output bytes.Buffer
	_, err := adapter.execute(ctx, root, ".", []string{"--version"},
		childexec.IOPolicy{Stdout: &boundedWriter{destination: &output, remaining: 1 << 20}})
	if err != nil {
		return "", err
	}
	fields := strings.Fields(output.String())
	version := ""
	if len(fields) == 1 {
		version = fields[0]
	} else if len(fields) >= 2 && fields[0] == "ansible-runner" {
		version = fields[1]
	} else {
		return "", errors.New("invalid Ansible Runner version response")
	}
	major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	if err != nil || major != 2 {
		return "", fmt.Errorf("unsupported Ansible Runner version %q", version)
	}
	return version, nil
}

// Configure runs one declared native playbook through Ansible Runner.
func (adapter Adapter) Configure(ctx context.Context, root, privateDataDir, projectDir,
	inventoryPath, playbook, artifactDir string, variableFiles []string, check bool) (Outcome, error) {
	for _, path := range []string{privateDataDir, projectDir, artifactDir} {
		if err := requireDirectory(root, path); err != nil {
			return Outcome{}, err
		}
	}
	for _, path := range append([]string{inventoryPath, filepath.Join(projectDir, playbook)}, variableFiles...) {
		if err := requireFile(root, path); err != nil {
			return Outcome{}, err
		}
	}
	cmdline := make([]string, 0, len(variableFiles)*2+2)
	for _, path := range variableFiles {
		cmdline = append(cmdline, "-e", "@"+filepath.Join(root, filepath.Clean(path)))
	}
	if check {
		cmdline = append(cmdline, "--check", "--diff")
	}
	arguments := []string{"run", filepath.Clean(privateDataDir), "--project-dir", filepath.Clean(projectDir),
		"--inventory", filepath.Clean(inventoryPath), "--artifact-dir", filepath.Clean(artifactDir),
		"--playbook", filepath.Clean(playbook)}
	if len(cmdline) > 0 {
		arguments = append(arguments, "--cmdline", quoteArguments(cmdline))
	}
	result, err := adapter.execute(ctx, root, ".", arguments, childexec.IOPolicy{})
	outcome := Outcome{Result: result}
	if err != nil {
		return outcome, err
	}
	stats, err := ReadStats(root, artifactDir)
	if err != nil {
		return outcome, err
	}
	outcome.Stats = stats
	return outcome, nil
}

// VerifyConverged requires complete expected-host coverage and zero changes,
// failures, or unreachable hosts in check-mode Runner evidence.
func VerifyConverged(stats Stats, expectedHosts []string) error {
	expected := append([]string(nil), expectedHosts...)
	sort.Strings(expected)
	if positive(stats.Dark) || positive(stats.Failures) || positive(stats.Changed) {
		return errors.New("ansible check reported changed, failed, or unreachable hosts")
	}
	processed := make([]string, 0, len(stats.Processed))
	for host, count := range stats.Processed {
		if count < 1 {
			return errors.New("ansible check reported an invalid processed-host count")
		}
		processed = append(processed, host)
	}
	sort.Strings(processed)
	if strings.Join(expected, "\x00") != strings.Join(processed, "\x00") {
		return errors.New("ansible check did not process the exact expected host set")
	}
	return nil
}

func positive(values map[string]int) bool {
	for _, value := range values {
		if value > 0 {
			return true
		}
	}
	return false
}

// ReadStats selects exactly one terminal playbook_on_stats event.
func ReadStats(root, artifactDir string) (Stats, error) {
	directory, err := security.ResolveContained(root, artifactDir)
	if err != nil {
		return Stats{}, err
	}
	artifactRoot, err := os.OpenRoot(directory)
	if err != nil {
		return Stats{}, fmt.Errorf("open ansible artifact root: %w", err)
	}
	defer func() { _ = artifactRoot.Close() }()
	var found *Stats
	files := 0
	var visit func(string, int) error
	visit = func(relative string, depth int) error {
		if depth > 8 {
			return errors.New("ansible artifact tree exceeds depth limit")
		}
		directoryFile, err := artifactRoot.Open(relative)
		if err != nil {
			return err
		}
		entries, readErr := directoryFile.ReadDir(-1)
		closeErr := directoryFile.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				return errors.New("symlink in ansible artifacts")
			}
			path := filepath.Join(relative, entry.Name())
			if entry.IsDir() {
				if err := visit(path, depth+1); err != nil {
					return err
				}
				continue
			}
			if filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			files++
			if files > 100000 {
				return errors.New("too many ansible event artifacts")
			}
			contents, err := artifactRoot.ReadFile(path)
			if err != nil || len(contents) > 4<<20 {
				return errors.New("invalid ansible event artifact")
			}
			var event struct {
				Event     string          `json:"event"`
				EventData json.RawMessage `json:"event_data"`
			}
			if json.Unmarshal(contents, &event) != nil || event.Event != "playbook_on_stats" {
				continue
			}
			if found != nil {
				return errors.New("multiple ansible stats events")
			}
			var stats Stats
			if err := json.Unmarshal(event.EventData, &stats); err != nil {
				return errors.New("invalid ansible stats event")
			}
			found = &stats
		}
		return nil
	}
	err = visit(".", 0)
	if err != nil {
		return Stats{}, fmt.Errorf("read Ansible artifacts: %w", err)
	}
	if found == nil {
		return Stats{}, errors.New("ansible-runner omitted playbook_on_stats evidence")
	}
	return *found, nil
}

func (adapter Adapter) execute(ctx context.Context, root, directory string, arguments []string, policy childexec.IOPolicy) (childexec.Result, error) {
	if err := adapter.Executable.VerifyUnchanged(); err != nil {
		return childexec.Result{}, err
	}
	run := adapter.Run
	if run == nil {
		run = (childexec.Runner{}).Run
	}
	result, err := run(ctx, childexec.Request{Executable: adapter.Executable, Args: arguments,
		WorkingRoot: root, WorkingDir: directory, Environment: adapter.Environment, IO: policy})
	if err != nil {
		return result, err
	}
	if result.Cancelled {
		return result, context.Canceled
	}
	if result.ExitCode != 0 {
		return result, fmt.Errorf("ansible-runner exited with status %d", result.ExitCode)
	}
	return result, nil
}

func requireDirectory(root, relative string) error {
	path, err := security.ResolveContained(root, relative)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("contained directory required: %s", relative)
	}
	return nil
}

func requireFile(root, relative string) error {
	path, err := security.ResolveContained(root, relative)
	if err != nil {
		return err
	}
	return security.RequireRegular(path)
}

func quoteArguments(arguments []string) string {
	quoted := make([]string, len(arguments))
	for index, argument := range arguments {
		quoted[index] = strconv.Quote(argument)
	}
	return strings.Join(quoted, " ")
}

type boundedWriter struct {
	destination *bytes.Buffer
	remaining   int
}

func (writer *boundedWriter) Write(contents []byte) (int, error) {
	if len(contents) > writer.remaining {
		return 0, errors.New("ansible output exceeds size limit")
	}
	written, err := writer.destination.Write(contents)
	writer.remaining -= written
	return written, err
}
