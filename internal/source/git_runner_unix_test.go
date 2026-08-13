//go:build !windows

package source_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
)

func TestAcquireGitThroughControlledExecutable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	script := filepath.Join(root, "git-fake")
	commit := strings.Repeat("b", 40)
	contents := `#!/bin/sh
case " $* " in
  *" rev-parse "*) printf '%s\n' "` + commit + `" ;;
  *" checkout "*)
    mkdir -p repository/provider
    printf 'template\n' > repository/provider/ainfra-template.yaml
    ;;
esac
`
	if err := os.WriteFile(script, []byte(contents), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := security.ResolveExecutable(script)
	if err != nil {
		t.Fatal(err)
	}
	environment, err := security.BuildEnvironment(os.Environ(), []string{"PATH"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := source.Parse(
		"git::https://example.com/templates.git//provider",
		"main; touch should-not-exist",
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := source.AcquireGit(context.Background(), reference, source.GitOptions{
		Executable: executable, Environment: environment, CacheRoot: filepath.Join(root, "cache"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Commit != commit {
		t.Fatalf("commit = %q", result.Commit)
	}
	if _, err := os.Stat(filepath.Join(root, "should-not-exist")); !os.IsNotExist(err) {
		t.Fatalf("requested ref crossed argument boundary: %v", err)
	}
}
