package security_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/projectious-work/ainfra/internal/security"
)

func TestResolveContained(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	inside := filepath.Join(root, "inside")
	if err := os.Mkdir(inside, 0o700); err != nil {
		t.Fatal(err)
	}
	resolved, err := security.ResolveContained(root, "inside")
	if err != nil {
		t.Fatalf("ResolveContained: %v", err)
	}
	if resolved != inside {
		t.Errorf("resolved = %q, want %q", resolved, inside)
	}
	if _, err := security.ResolveContained(root, "../outside"); err == nil {
		t.Error("traversal was accepted")
	}
}

func TestResolveContainedRejectsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions differ on Windows")
	}
	t.Parallel()
	root := t.TempDir()
	target := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := security.ResolveContained(root, link); err == nil {
		t.Error("symlink was accepted")
	}
}

func TestBuildEnvironment(t *testing.T) {
	t.Parallel()
	environment, err := security.BuildEnvironment(
		[]string{"HOME=/secret/home", "LANG=C", "TOKEN=secret"},
		[]string{"LANG"},
		map[string]string{"AINFRA_CHILD": "1"},
	)
	if err != nil {
		t.Fatalf("BuildEnvironment: %v", err)
	}
	want := []string{"AINFRA_CHILD=1", "LANG=C"}
	if !reflect.DeepEqual(environment.Values(), want) {
		t.Errorf("environment = %q, want %q", environment.Values(), want)
	}
}

func TestEmptyEnvironmentIsExplicit(t *testing.T) {
	t.Parallel()
	environment, err := security.BuildEnvironment(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if environment.Values() == nil || len(environment.Values()) != 0 {
		t.Errorf("empty environment = %#v", environment.Values())
	}
}

func TestCreatePrivateFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "private")
	file, err := security.CreatePrivateFile(root, "private")
	if err != nil {
		t.Fatalf("CreatePrivateFile: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permissions = %o", info.Mode().Perm())
	}
	if _, err := security.CreatePrivateFile(root, "private"); err == nil {
		t.Error("existing file was overwritten")
	}
}

func TestRedactingWriterAcrossChunks(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	writer := security.NewRedactingWriter(&output, []string{"secret"})
	if _, err := writer.Write([]byte("before sec")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("ret after")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "before <redacted> after" {
		t.Errorf("redacted output = %q", output.String())
	}
}

func FuzzBuildEnvironment(f *testing.F) {
	f.Add("LANG=C")
	f.Add("TOKEN=secret")
	f.Fuzz(func(t *testing.T, entry string) {
		environment, err := security.BuildEnvironment([]string{entry}, nil, nil)
		if err == nil && len(environment.Values()) != 0 {
			t.Fatalf("empty allowlist leaked %q", environment.Values())
		}
	})
}
