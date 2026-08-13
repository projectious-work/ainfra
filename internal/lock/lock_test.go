package lock_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/lock"
)

func TestWriteReadAndEquivalent(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), lock.Filename)
	document := lock.New(lock.Template{
		Source: "local:../template", Resolved: "local:../template",
		Version: "1.0.0", Digest: "sha256:" + strings.Repeat("a", 64),
		ResolvedAt: time.Date(2026, 8, 13, 12, 34, 56, 123, time.FixedZone("test", 3600)),
	})
	if err := lock.Write(path, document); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "resolvedAt: 2026-08-13T11:34:56Z") {
		t.Fatalf("non-canonical timestamp:\n%s", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("lock mode = %o", info.Mode().Perm())
	}
	read, err := lock.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	changedTime := document
	changedTime.Template.ResolvedAt = changedTime.Template.ResolvedAt.Add(time.Hour)
	if !lock.Equivalent(read, changedTime) {
		t.Fatal("resolution time made identical lock unequal")
	}
}

func TestWriteExcludesConcurrentWriter(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), lock.Filename)
	if err := os.WriteFile(path+".writer", []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}
	document := lock.New(lock.Template{
		Source: "local:x", Resolved: "local:x", Version: "1.0.0",
		Digest: "sha256:" + strings.Repeat("b", 64), ResolvedAt: time.Now(),
	})
	if err := lock.Write(path, document); err == nil {
		t.Fatal("concurrent writer unexpectedly accepted")
	}
}
