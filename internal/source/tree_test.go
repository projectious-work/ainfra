package source_test

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/projectious-work/ainfra/internal/source"
)

func TestTreeDigestUsesNormativeFramingAndOrdering(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTreeFile(t, root, "z.txt", []byte("last"), 0o600)
	writeTreeFile(t, root, "bin/tool", []byte("first"), 0o700)
	writeTreeFile(t, root, ".git/config", []byte("excluded"), 0o600)

	got, err := source.TreeDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	want := normativeDigest([]digestFixture{
		{path: "bin/tool", executable: true, contents: []byte("first")},
		{path: "z.txt", contents: []byte("last")},
	})
	if got != want {
		t.Fatalf("digest = %q, want %q", got, want)
	}
}

func TestTreeDigestBindsContentPathAndExecutableBit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := writeTreeFile(t, root, "tool", []byte("contents"), 0o600)
	baseline, err := source.TreeDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := source.TreeDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if baseline == executable {
		t.Fatal("executable-bit change did not change digest")
	}
	if err := os.Rename(path, filepath.Join(root, "renamed")); err != nil {
		t.Fatal(err)
	}
	renamed, err := source.TreeDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if executable == renamed {
		t.Fatal("path change did not change digest")
	}
}

func TestEngineWorkspaceDigestExcludesOnlyOpenTofuInitialization(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTreeFile(t, root, "main.tf", []byte("resource {}"), 0o600)
	want, err := source.TreeDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".terraform"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeTreeFile(t, root, ".terraform/provider", []byte("binary"), 0o700)
	writeTreeFile(t, root, ".terraform.lock.hcl", []byte("generated"), 0o600)
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "main.tf", []byte("resource {}"), 0o600)
	got, err := source.EngineWorkspaceDigest(root, templateRoot)
	if err != nil || got != want {
		t.Fatalf("digest=%q want=%q err=%v", got, want, err)
	}
	writeTreeFile(t, root, "injected.tf", []byte("resource {}"), 0o600)
	if got, err := source.EngineWorkspaceDigest(root, templateRoot); err != nil || got == want {
		t.Fatalf("injected configuration was not bound: digest=%q err=%v", got, err)
	}
}

func TestTreeDigestRejectsUnsafeTrees(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		prepare func(*testing.T, string)
	}{
		{name: "ainfra runtime", prepare: func(t *testing.T, root string) {
			writeTreeFile(t, root, ".ainfra/run.json", nil, 0o600)
		}},
		{name: "opentofu runtime", prepare: func(t *testing.T, root string) {
			writeTreeFile(t, root, "tofu/.terraform/providers", nil, 0o600)
		}},
		{name: "state", prepare: func(t *testing.T, root string) {
			writeTreeFile(t, root, "tofu/terraform.tfstate.backup", nil, 0o600)
		}},
		{name: "symlink", prepare: func(t *testing.T, root string) {
			if err := os.Symlink("target", filepath.Join(root, "link")); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "hard link", prepare: func(t *testing.T, root string) {
			target := writeTreeFile(t, root, "target", nil, 0o600)
			if err := os.Link(target, filepath.Join(root, "duplicate")); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "non NFC path", prepare: func(t *testing.T, root string) {
			writeTreeFile(t, root, "cafe\u0301.txt", nil, 0o600)
		}},
	}
	if runtime.GOOS != "windows" {
		tests = append(tests, struct {
			name    string
			prepare func(*testing.T, string)
		}{name: "fifo", prepare: func(t *testing.T, root string) {
			if err := makeFIFO(filepath.Join(root, "pipe")); err != nil {
				t.Fatal(err)
			}
		}})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			test.prepare(t, root)
			if digest, err := source.TreeDigest(root); err == nil {
				t.Fatalf("TreeDigest() = %q, want refusal", digest)
			}
		})
	}
}

func TestTreeDigestRejectsNonDirectoryRoot(t *testing.T) {
	t.Parallel()
	root := writeTreeFile(t, t.TempDir(), "file", nil, 0o600)
	if _, err := source.TreeDigest(root); err == nil {
		t.Fatal("regular-file root unexpectedly accepted")
	}
}

func TestRequirePrivateTreeRejectsPermissionAndSymlinkDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	file := writeTreeFile(t, root, "nested/file", []byte("trusted"), 0o600)
	for _, path := range []string{root, filepath.Join(root, "nested")} {
		if err := os.Chmod(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := source.RequirePrivateTree(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := source.RequirePrivateTree(root); err == nil {
		t.Fatal("world-readable file unexpectedly accepted")
	}
	if err := os.Chmod(file, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("file", filepath.Join(root, "nested", "link")); err != nil {
		t.Fatal(err)
	}
	if err := source.RequirePrivateTree(root); err == nil {
		t.Fatal("symlinked entry unexpectedly accepted")
	}
}

func FuzzTreeDigestFile(f *testing.F) {
	f.Add([]byte("template"), false)
	f.Add([]byte{0, 1, 2, 255}, true)
	f.Fuzz(func(t *testing.T, contents []byte, executable bool) {
		root := t.TempDir()
		mode := os.FileMode(0o600)
		if executable {
			mode = 0o700
		}
		writeTreeFile(t, root, "nested/file", contents, mode)
		first, err := source.TreeDigest(root)
		if err != nil {
			t.Fatal(err)
		}
		second, err := source.TreeDigest(root)
		if err != nil {
			t.Fatal(err)
		}
		if first != second {
			t.Fatalf("digest is not deterministic: %q != %q", first, second)
		}
	})
}

type digestFixture struct {
	path       string
	executable bool
	contents   []byte
}

func normativeDigest(files []digestFixture) string {
	digest := sha256.New()
	_, _ = digest.Write([]byte("ainfra-template-tree-v1\x00"))
	var integer [8]byte
	for _, file := range files {
		binary.BigEndian.PutUint64(integer[:], uint64(len([]byte(file.path))))
		_, _ = digest.Write(integer[:])
		_, _ = digest.Write([]byte(file.path))
		if file.executable {
			_, _ = digest.Write([]byte{1})
		} else {
			_, _ = digest.Write([]byte{0})
		}
		binary.BigEndian.PutUint64(integer[:], uint64(len(file.contents)))
		_, _ = digest.Write(integer[:])
		_, _ = digest.Write(file.contents)
	}
	return fmt.Sprintf("sha256:%x", digest.Sum(nil))
}

func writeTreeFile(t *testing.T, root, relative string, contents []byte, mode os.FileMode) string {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, contents, mode); err != nil {
		t.Fatal(err)
	}
	return target
}
