package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/projectious-work/ainfra/internal/source"
)

func TestMaterializeLocalPublishesVerifiedPrivateCache(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "README.md", []byte("template"), 0o644)
	writeTreeFile(t, templateRoot, "bin/check", []byte("#!/bin/sh\n"), 0o755)
	writeTreeFile(t, templateRoot, ".git/config", []byte("excluded"), 0o600)
	cacheRoot := t.TempDir()

	first, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if first.Reused || first.Path == "" || first.Digest == "" {
		t.Fatalf("unexpected first result: %#v", first)
	}
	if _, err := os.Stat(filepath.Join(first.Path, ".git")); !os.IsNotExist(err) {
		t.Fatalf("Git metadata was materialized: %v", err)
	}
	for _, relative := range []string{"README.md", "bin/check"} {
		info, err := os.Stat(filepath.Join(first.Path, relative))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("materialized mode for %s is not private: %o", relative, info.Mode().Perm())
		}
	}
	second, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Reused || second.Path != first.Path || second.Digest != first.Digest {
		t.Fatalf("unexpected reuse result: %#v", second)
	}
}

func TestMaterializeLocalRejectsPoisonedCacheEntry(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "file", []byte("trusted"), 0o600)
	cacheRoot := t.TempDir()
	result, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(result.Path, "file"), []byte("poisoned"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := source.MaterializeLocal(templateRoot, cacheRoot); err == nil {
		t.Fatal("poisoned cache entry unexpectedly reused")
	}
}

func TestMaterializeLocalRejectsPermissionPoisonedCacheEntry(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "file", []byte("trusted"), 0o600)
	cacheRoot := t.TempDir()
	result, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(result.Path, "file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := source.MaterializeLocal(templateRoot, cacheRoot); err == nil {
		t.Fatal("permission-poisoned cache entry unexpectedly reused")
	}
}

func TestMaterializeLocalRejectsSymlinkedCacheEntry(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "file", []byte("trusted"), 0o600)
	digest, err := source.TreeDigest(templateRoot)
	if err != nil {
		t.Fatal(err)
	}
	cacheRoot := t.TempDir()
	entries := filepath.Join(cacheRoot, "templates", "sha256")
	if err := os.MkdirAll(entries, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(templateRoot, filepath.Join(entries, digest[len("sha256:"):])); err != nil {
		t.Fatal(err)
	}
	if _, err := source.MaterializeLocal(templateRoot, cacheRoot); err == nil {
		t.Fatal("symlinked cache entry unexpectedly reused")
	}
}

func TestMaterializeLocalRejectsUnsafeSource(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	if err := os.Symlink("outside", filepath.Join(templateRoot, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := source.MaterializeLocal(templateRoot, t.TempDir()); err == nil {
		t.Fatal("unsafe source unexpectedly materialized")
	}
}

func TestMaterializeLocalRejectsOverlappingRoots(t *testing.T) {
	t.Parallel()
	templateRoot := t.TempDir()
	writeTreeFile(t, templateRoot, "file", []byte("trusted"), 0o600)
	if _, err := source.MaterializeLocal(templateRoot, filepath.Join(templateRoot, "cache")); err == nil {
		t.Fatal("cache nested in source unexpectedly accepted")
	}
}
