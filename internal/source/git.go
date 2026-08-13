package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/security"
)

var gitCommitPattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// GitRun executes one already-policy-bound Git request.
type GitRun func(context.Context, childexec.Request) (childexec.Result, error)

// GitOptions supplies the explicit executable, environment, and cache policy.
type GitOptions struct {
	Executable  security.Executable
	Environment security.Environment
	CacheRoot   string
	Run         GitRun
}

// GitInvocation is a display-safe record of one acquisition boundary call.
type GitInvocation struct {
	Executable  string
	Arguments   []string
	WorkingRoot string
	WorkingDir  string
}

// GitAcquisition identifies one immutable verified Git materialization.
type GitAcquisition struct {
	Commit       string
	Materialized Materialized
	Invocations  []GitInvocation
}

// AcquireGit resolves an explicit ref to an immutable commit, checks out only
// into private staging, and publishes verified template content to the shared
// digest-addressed cache. It never configures a persistent remote.
func AcquireGit(ctx context.Context, reference Reference, options GitOptions) (GitAcquisition, error) {
	if reference.Kind != KindGit {
		return GitAcquisition{}, errors.New("git acquisition requires a Git reference")
	}
	if options.CacheRoot == "" {
		return GitAcquisition{}, errors.New("git acquisition requires an explicit cache root")
	}
	acquisitionRoot := filepath.Join(options.CacheRoot, "acquisition")
	if err := os.MkdirAll(acquisitionRoot, 0o700); err != nil {
		return GitAcquisition{}, fmt.Errorf("create Git acquisition root: %w", err)
	}
	if err := os.Chmod(acquisitionRoot, 0o700); err != nil { // #nosec G302 -- private directory requires owner execute permission.
		return GitAcquisition{}, fmt.Errorf("secure Git acquisition root: %w", err)
	}
	staging, err := os.MkdirTemp(acquisitionRoot, ".staging-")
	if err != nil {
		return GitAcquisition{}, fmt.Errorf("create Git acquisition staging: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()
	if err := os.Chmod(staging, 0o700); err != nil { // #nosec G302 -- private directory requires owner execute permission.
		return GitAcquisition{}, fmt.Errorf("secure Git acquisition staging: %w", err)
	}
	if err := os.Mkdir(filepath.Join(staging, "repository"), 0o700); err != nil {
		return GitAcquisition{}, fmt.Errorf("create Git working directory: %w", err)
	}

	run := options.Run
	if run == nil {
		runner := childexec.Runner{}
		run = runner.Run
	}
	invocations := make([]GitInvocation, 0, 4)
	runGit := func(arguments []string, stdout *bytes.Buffer) error {
		displayArguments := append([]string(nil), arguments...)
		for index, argument := range displayArguments {
			if argument == reference.Repository {
				displayArguments[index] = strings.TrimPrefix(reference.Display, "git::")
				if reference.Subdirectory != "" {
					displayArguments[index] = strings.TrimSuffix(
						displayArguments[index], "//"+reference.Subdirectory,
					)
				}
			}
		}
		invocations = append(invocations, GitInvocation{
			Executable: options.Executable.Path(), Arguments: displayArguments,
			WorkingRoot: staging, WorkingDir: ".",
		})
		result, runErr := run(ctx, childexec.Request{
			Executable: options.Executable, Args: arguments,
			WorkingRoot: staging, WorkingDir: ".", Environment: options.Environment,
			IO:              childexec.IOPolicy{Stdout: stdout},
			SensitiveValues: []string{reference.Repository},
		})
		if runErr != nil {
			return runErr
		}
		if result.Cancelled {
			return context.Canceled
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("git exited with status %d", result.ExitCode)
		}
		return nil
	}
	if err := runGit([]string{"-C", "repository", "init", "--no-template"}, nil); err != nil {
		return GitAcquisition{}, fmt.Errorf("initialize Git acquisition: %w", err)
	}
	if err := runGit([]string{
		"-C", "repository", "fetch", "--depth=1", "--no-tags", "--",
		reference.Repository, reference.RequestedRef,
	}, nil); err != nil {
		return GitAcquisition{}, fmt.Errorf("fetch Git template ref: %w", err)
	}
	var revision bytes.Buffer
	if err := runGit([]string{
		"-C", "repository", "rev-parse", "--verify", "FETCH_HEAD^{commit}",
	}, &revision); err != nil {
		return GitAcquisition{}, fmt.Errorf("resolve immutable Git commit: %w", err)
	}
	commit := strings.TrimSpace(revision.String())
	if !gitCommitPattern.MatchString(commit) {
		return GitAcquisition{}, errors.New("git returned an invalid immutable commit")
	}
	if err := runGit([]string{
		"-C", "repository", "-c", "advice.detachedHead=false",
		"checkout", "--detach", commit, "--",
	}, nil); err != nil {
		return GitAcquisition{}, fmt.Errorf("checkout immutable Git commit: %w", err)
	}
	selected := filepath.Join(staging, "repository", filepath.FromSlash(reference.Subdirectory))
	materialized, err := materializeLocal(selected, options.CacheRoot, true)
	if err != nil {
		return GitAcquisition{}, fmt.Errorf("materialize Git template: %w", err)
	}
	return GitAcquisition{
		Commit: commit, Materialized: materialized, Invocations: invocations,
	}, nil
}
