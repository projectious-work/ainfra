package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/project"
)

func TestLoadStrictManifestAndPreserveInputOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"first.tfvars", "second.tfvars", "backend.hcl"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("opaque\x00bytes"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
  inputs:
    tofu:
      variableFiles: [first.tfvars, second.tfvars]
      backendConfigFiles: [backend.hcl]
`
	writeDeploymentManifest(t, root, manifest)

	deployment, err := project.Load(project.ResolveOptions{ExplicitPath: root})
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	want := []string{"first.tfvars", "second.tfvars"}
	if strings.Join(deployment.Inputs.TofuVariableFiles, ",") != strings.Join(want, ",") {
		t.Fatalf("input order changed: %#v", deployment.Inputs.TofuVariableFiles)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeDeploymentManifest(t, root, `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
  unexpected: true
spec:
  template:
    source: local:../template
`)
	if _, err := project.Load(project.ResolveOptions{ExplicitPath: root}); err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
}

func TestLoadRejectsTraversalDuplicateAndSymlinkInputs(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		prepare func(*testing.T, string)
		paths   string
	}{
		"traversal": {paths: "[../outside.tfvars]"},
		"duplicate": {
			prepare: func(t *testing.T, root string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, "input.tfvars"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			paths: "[input.tfvars, input.tfvars]",
		},
		"symlink": {
			prepare: func(t *testing.T, root string) {
				t.Helper()
				outside := filepath.Join(t.TempDir(), "outside.tfvars")
				if err := os.WriteFile(outside, nil, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(root, "input.tfvars")); err != nil {
					t.Fatal(err)
				}
			},
			paths: "[input.tfvars]",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if test.prepare != nil {
				test.prepare(t, root)
			}
			manifest := `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    source: local:../template
  inputs:
    tofu:
      variableFiles: ` + test.paths + "\n"
			writeDeploymentManifest(t, root, manifest)
			if _, err := project.Load(project.ResolveOptions{ExplicitPath: root}); err == nil {
				t.Fatal("unsafe input unexpectedly accepted")
			}
		})
	}
}

func writeDeploymentManifest(t *testing.T, root, contents string) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(root, project.ManifestName), []byte(contents), 0o600,
	); err != nil {
		t.Fatal(err)
	}
}
