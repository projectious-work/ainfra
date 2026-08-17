package migration_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/projectious-work/ainfra/internal/migration"
)

func TestAnalyzeV1TemplateIsDeterministicAndNonMutating(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "template-example")
	if err := os.CopyFS(source, os.DirFS("../../spec/examples/v1/template-example")); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(source, "ainfra-template.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := migration.Analyze(source, migration.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := migration.Analyze(source, migration.VersionV1)
	if err != nil || first.SourceVersion != migration.VersionV1 ||
		first.TargetVersion != migration.VersionV1 || len(first.Changes) != 0 ||
		!reflect.DeepEqual(first, second) {
		t.Fatalf("migration plans: first=%+v second=%+v err=%v", first, second, err)
	}
	after, err := os.ReadFile(filepath.Join(source, "ainfra-template.yaml"))
	if err != nil || string(before) != string(after) {
		t.Fatalf("migration analysis changed source: %v", err)
	}
}

func TestAnalyzeRejectsUnsupportedAndUnsafeSources(t *testing.T) {
	t.Parallel()
	if _, err := migration.Analyze(t.TempDir(), "v2"); err == nil {
		t.Fatal("unsupported target succeeded")
	}
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(root, "linked")
	if err := os.Symlink(target, linked); err != nil {
		t.Fatal(err)
	}
	if _, err := migration.Analyze(linked, migration.VersionV1); err == nil {
		t.Fatal("symlink source succeeded")
	}
}
