// Package exec runs child processes without a shell under explicit policy.
package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/projectious-work/ainfra/internal/security"
)

// IOPolicy supplies sanitized child stream destinations.
type IOPolicy struct {
	Stdout    io.Writer
	Stderr    io.Writer
	RawStdout io.Writer
	RawStderr io.Writer
}

// Request is a complete immutable child-process request.
type Request struct {
	Executable      security.Executable
	Args            []string
	WorkingRoot     string
	WorkingDir      string
	Environment     security.Environment
	Stdin           io.Reader
	IO              IOPolicy
	SensitiveValues []string
}

// Result describes only the observed child-process outcome.
type Result struct {
	Started   bool
	ExitCode  int
	Cancelled bool
	Signalled bool
}

// Runner executes child processes with a bounded graceful cancellation period.
type Runner struct {
	CancellationGrace time.Duration
}

// Run executes one request without invoking a shell.
func (runner Runner) Run(ctx context.Context, request Request) (Result, error) {
	if err := request.Executable.VerifyUnchanged(); err != nil {
		return Result{}, fmt.Errorf("verify executable: %w", err)
	}
	directory, err := security.ResolveContained(request.WorkingRoot, request.WorkingDir)
	if err != nil {
		return Result{}, fmt.Errorf("verify working directory: %w", err)
	}
	stdout := request.IO.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := request.IO.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	redactedOut := security.NewRedactingWriter(stdout, request.SensitiveValues)
	redactedErr := security.NewRedactingWriter(stderr, request.SensitiveValues)

	// #nosec G204 -- this package is the sole process boundary; the executable
	// is immutable and reverified, argv remains a structured array, and no shell
	// is involved.
	command := exec.Command(request.Executable.Path(), append([]string(nil), request.Args...)...)
	command.Dir = directory
	command.Env = request.Environment.Values()
	command.Stdin = request.Stdin
	command.Stdout = retainedWriter{raw: request.IO.RawStdout, sanitized: redactedOut}
	command.Stderr = retainedWriter{raw: request.IO.RawStderr, sanitized: redactedErr}
	configureProcess(command)
	if err := command.Start(); err != nil {
		return Result{}, fmt.Errorf("start child process: %w", err)
	}

	waited := make(chan error, 1)
	go func() { waited <- command.Wait() }()
	result := Result{Started: true, ExitCode: 0}
	waitErr := runner.wait(ctx, command, waited, &result)
	closeErr := errors.Join(redactedOut.Close(), redactedErr.Close())
	if closeErr != nil {
		return result, fmt.Errorf("flush sanitized child output: %w", closeErr)
	}
	if waitErr == nil {
		return result, nil
	}
	var exitError *exec.ExitError
	if errors.As(waitErr, &exitError) {
		result.ExitCode = exitError.ExitCode()
		result.Signalled = exitError.ExitCode() < 0
		return result, nil
	}
	return result, fmt.Errorf("wait for child process: %w", waitErr)
}

// retainedWriter preserves exact child bytes before sending the same bytes
// through the sanitized output boundary. A raw retention failure stops the
// child stream instead of silently producing incomplete evidence.
type retainedWriter struct {
	raw       io.Writer
	sanitized io.Writer
}

func (writer retainedWriter) Write(contents []byte) (int, error) {
	if writer.raw != nil {
		written, err := writer.raw.Write(contents)
		if err != nil {
			return written, err
		}
		if written != len(contents) {
			return written, io.ErrShortWrite
		}
	}
	return writer.sanitized.Write(contents)
}

func (runner Runner) wait(ctx context.Context, command *exec.Cmd, waited <-chan error, result *Result) error {
	select {
	case err := <-waited:
		return err
	case <-ctx.Done():
		result.Cancelled = true
		interruptProcess(command)
	}
	grace := runner.CancellationGrace
	if grace <= 0 {
		grace = 2 * time.Second
	}
	timer := time.NewTimer(grace)
	defer timer.Stop()
	select {
	case err := <-waited:
		return err
	case <-timer.C:
		killProcess(command)
		return <-waited
	}
}
