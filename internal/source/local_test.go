package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/source"
)

func TestResolveLocalAllowsDeploymentAndRepositoryRoots(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	if err := os.Mkdir(filepath.Join(repository, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	deployment := filepath.Join(repository, "deployments", "dev")
	templateRoot := filepath.Join(repository, "templates", "base")
	for _, path := range []string{deployment, templateRoot, filepath.Join(deployment, "inline")} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		value string
		want  string
	}{
		{value: "local:inline", want: filepath.Join(deployment, "inline")},
		{value: "local:../../templates/base", want: templateRoot},
	} {
		reference, err := source.Parse(test.value, "")
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := source.ResolveLocal(reference, deployment)
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Path != test.want || resolved.ApprovedRoot != repository && test.value != "local:inline" {
			t.Fatalf("resolution = %#v, want path %q", resolved, test.want)
		}
	}
}

func TestResolveLocalRejectsOutsideApprovedRootsAndSymlinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	repository := filepath.Join(root, "repository")
	deployment := filepath.Join(repository, "deployment")
	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(deployment, 0o700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(deployment, outside)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := source.Parse("local:"+filepath.ToSlash(relative), "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.ResolveLocal(reference, deployment); err == nil {
		t.Fatal("outside source unexpectedly accepted")
	}
	if err := os.Symlink(outside, filepath.Join(deployment, "linked")); err != nil {
		t.Fatal(err)
	}
	reference, err = source.Parse("local:linked", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.ResolveLocal(reference, deployment); err == nil {
		t.Fatal("symlinked source unexpectedly accepted")
	}
}
