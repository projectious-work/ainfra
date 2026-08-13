package initialize_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/initialize"
)

func TestCreateMinimalDeployment(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "example-deployment")
	result, err := initialize.Create(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "example-deployment" || result.Root != root ||
		len(result.CreatedPaths) != 2 {
		t.Fatalf("result=%#v", result)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "ainfra.yaml"))
	if err != nil || !strings.Contains(string(manifest), "name: example-deployment") ||
		strings.Contains(strings.ToLower(string(manifest)), "password") {
		t.Fatalf("manifest=%q err=%v", manifest, err)
	}
	ignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil || strings.Count(string(ignore), "# ainfra BEGIN") != 1 {
		t.Fatalf("ignore=%q err=%v", ignore, err)
	}
}

func TestCreateDetectsConflictBeforeWriting(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "conflict-deployment")
	if err := os.Mkdir(root, 0o750); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "ainfra.yaml")
	if err := os.WriteFile(manifest, []byte("owned\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := initialize.Create(root); err == nil {
		t.Fatal("existing manifest unexpectedly overwritten")
	}
	contents, err := os.ReadFile(manifest)
	if err != nil || string(contents) != "owned\n" {
		t.Fatalf("manifest changed: %q err=%v", contents, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf("write occurred before conflict return: %v", err)
	}
}

func TestCreateRejectsIncompleteIgnoreBlockBeforeWriting(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "ignore-conflict")
	if err := os.Mkdir(root, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, ".gitignore"), []byte("# ainfra BEGIN\n"), 0o644,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := initialize.Create(root); err == nil {
		t.Fatal("incomplete marker unexpectedly accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "ainfra.yaml")); !os.IsNotExist(err) {
		t.Fatalf("manifest written before conflict return: %v", err)
	}
}
