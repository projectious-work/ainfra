package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/project"
)

func TestResolveDirectoryAndManifestAreEquivalent(t *testing.T) {
	t.Parallel()
	root := deployment(t)

	fromDirectory, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: t.TempDir(), ExplicitPath: root,
	})
	if err != nil {
		t.Fatalf("resolve directory: %v", err)
	}
	fromManifest, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: t.TempDir(),
		ExplicitPath:     filepath.Join(root, project.ManifestName),
	})
	if err != nil {
		t.Fatalf("resolve manifest: %v", err)
	}
	if fromDirectory.Root != fromManifest.Root ||
		fromDirectory.ManifestPath != fromManifest.ManifestPath {
		t.Fatalf("targets differ: %#v != %#v", fromDirectory, fromManifest)
	}
}

func TestResolveUsesNormativePrecedence(t *testing.T) {
	t.Parallel()
	explicit := deployment(t)
	flag := deployment(t)
	environment := deployment(t)
	ancestor := deployment(t)
	nested := filepath.Join(ancestor, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}

	target, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: nested,
		ExplicitPath:     explicit,
		ProjectPath:      flag,
		EnvironmentPath:  environment,
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if target.Source != project.SourceExplicit || target.Root != explicit {
		t.Fatalf("unexpected selection: %#v", target)
	}
}

func TestResolveDiscoversNearestAncestor(t *testing.T) {
	t.Parallel()
	root := deployment(t)
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	target, err := project.Resolve(project.ResolveOptions{WorkingDirectory: nested})
	if err != nil {
		t.Fatalf("resolve ancestor: %v", err)
	}
	if target.Source != project.SourceAncestor || target.Root != root {
		t.Fatalf("unexpected ancestor: %#v", target)
	}
}

func TestResolveExplicitDirectoryDoesNotSearchChildren(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o700); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, child)

	if _, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: t.TempDir(), ExplicitPath: root,
	}); err == nil {
		t.Fatal("explicit directory without a manifest unexpectedly resolved")
	}
}

func TestResolveRejectsSymlinkManifest(t *testing.T) {
	t.Parallel()
	realRoot := deployment(t)
	root := t.TempDir()
	if err := os.Symlink(
		filepath.Join(realRoot, project.ManifestName),
		filepath.Join(root, project.ManifestName),
	); err != nil {
		t.Fatal(err)
	}
	if _, err := project.Resolve(project.ResolveOptions{
		WorkingDirectory: t.TempDir(), ExplicitPath: root,
	}); err == nil {
		t.Fatal("symlink manifest unexpectedly resolved")
	}
}

func deployment(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeManifest(t, root)
	return root
}

func writeManifest(t *testing.T, root string) {
	t.Helper()
	contents := []byte("apiVersion: ainfra.projectious.work/v1\nkind: Deployment\n")
	if err := os.WriteFile(filepath.Join(root, project.ManifestName), contents, 0o600); err != nil {
		t.Fatal(err)
	}
}
