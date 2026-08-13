package source_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	childexec "github.com/projectious-work/ainfra/internal/exec"
	"github.com/projectious-work/ainfra/internal/source"
)

func TestAcquireGitUsesArgumentArraysAndMaterializesCommit(t *testing.T) {
	t.Parallel()
	reference, err := source.Parse(
		"git::https://user:secret@example.com/templates.git//provider",
		"release/v1; touch unsafe",
	)
	if err != nil {
		t.Fatal(err)
	}
	var requests []childexec.Request
	commit := strings.Repeat("a", 40)
	result, err := source.AcquireGit(context.Background(), reference, source.GitOptions{
		CacheRoot: t.TempDir(),
		Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
			requests = append(requests, request)
			if containsArgument(request.Args, "rev-parse") {
				_, _ = fmt.Fprintln(request.IO.Stdout, commit)
			}
			if containsArgument(request.Args, "checkout") {
				root := filepath.Join(request.WorkingRoot, "repository", "provider")
				writeGitTemplateTree(t, root)
			}
			return childexec.Result{Started: true}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Commit != commit || result.Materialized.Digest == "" {
		t.Fatalf("unexpected acquisition: %#v", result)
	}
	if len(requests) != 4 || !reflect.DeepEqual(requests[1].Args, []string{
		"-C", "repository", "fetch", "--depth=1", "--no-tags", "--",
		reference.Repository, reference.RequestedRef,
	}) {
		t.Fatalf("fetch arguments = %#v", requests[1].Args)
	}
	for _, invocation := range result.Invocations {
		if strings.Contains(strings.Join(invocation.Arguments, " "), "secret") {
			t.Fatalf("invocation leaked credential: %#v", invocation)
		}
	}
	if _, err := os.Stat(result.Materialized.Path); err != nil {
		t.Fatal(err)
	}
}

func TestAcquireGitRejectsInvalidRevisionOutput(t *testing.T) {
	t.Parallel()
	reference, err := source.Parse("git::https://example.com/templates.git", "main")
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.AcquireGit(context.Background(), reference, source.GitOptions{
		CacheRoot: t.TempDir(),
		Run: func(_ context.Context, request childexec.Request) (childexec.Result, error) {
			if containsArgument(request.Args, "rev-parse") {
				_, _ = fmt.Fprintln(request.IO.Stdout, "not-a-commit")
			}
			return childexec.Result{Started: true}, nil
		},
	})
	if err == nil {
		t.Fatal("invalid immutable revision unexpectedly accepted")
	}
}

func TestAcquireGitPropagatesFailureAndCancellation(t *testing.T) {
	t.Parallel()
	reference, err := source.Parse("git::ssh://git@example.com/templates.git", "main")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []childexec.Result{
		{Started: true, ExitCode: 9}, {Started: true, Cancelled: true},
	} {
		_, err := source.AcquireGit(context.Background(), reference, source.GitOptions{
			CacheRoot: t.TempDir(),
			Run: func(context.Context, childexec.Request) (childexec.Result, error) {
				return result, nil
			},
		})
		if err == nil {
			t.Fatalf("result %#v unexpectedly accepted", result)
		}
	}
}

func containsArgument(arguments []string, target string) bool {
	for _, argument := range arguments {
		if argument == target {
			return true
		}
	}
	return false
}

func writeGitTemplateTree(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTreeFile(t, root, "ainfra-template.yaml", []byte("template"), 0o600)
}
