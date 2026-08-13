package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/source"
)

func TestTemplateLockMaterializesAndPublishesLocalBinding(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	mustMkdir(t, filepath.Join(repository, ".git"))
	deployment := filepath.Join(repository, "deployments", "dev")
	templateRoot := filepath.Join(repository, "templates", "base")
	mustMkdir(t, deployment)
	writeTestFile(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: dev
spec:
  template:
    source: local:../../templates/base
`)
	writeLocalTemplate(t, templateRoot)
	cacheRoot := t.TempDir()
	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	options := app.TemplateLockOptions{
		WorkingDirectory: deployment, CacheDirectory: cacheRoot,
		Environment: map[string]string{}, Now: func() time.Time { return now },
	}
	first, err := app.TemplateLock(app.TemplateLockRequest{}, options)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Changed || first.Source != "local:../../templates/base" ||
		!strings.HasPrefix(first.ContentDigest, "sha256:") {
		t.Fatalf("unexpected first result: %#v", first)
	}
	document, err := lockfile.Read(filepath.Join(deployment, lockfile.Filename))
	if err != nil {
		t.Fatal(err)
	}
	if document.Template.Version != "1.2.3" ||
		document.Template.Digest != first.ContentDigest ||
		document.Template.ResolvedAt != now {
		t.Fatalf("unexpected lock: %#v", document)
	}
	second, err := app.TemplateLock(app.TemplateLockRequest{}, options)
	if err != nil {
		t.Fatal(err)
	}
	if second.Changed || second.ContentDigest != first.ContentDigest {
		t.Fatalf("unexpected repeated result: %#v", second)
	}
}

func TestTemplateLockRefusesChangedExistingBinding(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	mustMkdir(t, filepath.Join(repository, ".git"))
	deployment := filepath.Join(repository, "deployment")
	templateRoot := filepath.Join(repository, "template")
	mustMkdir(t, deployment)
	writeTestFile(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: dev
spec:
  template:
    source: local:../template
`)
	writeLocalTemplate(t, templateRoot)
	options := app.TemplateLockOptions{
		WorkingDirectory: deployment, CacheDirectory: t.TempDir(),
		Environment: map[string]string{},
	}
	if _, err := app.TemplateLock(app.TemplateLockRequest{}, options); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(templateRoot, "README.md"), "tofu changed\n")
	if _, err := app.TemplateLock(app.TemplateLockRequest{}, options); err == nil ||
		!strings.Contains(err.Error(), "template update") {
		t.Fatalf("changed binding error = %v", err)
	}
}

func TestTemplateLockAndUpdateGitBinding(t *testing.T) {
	t.Parallel()
	deployment := t.TempDir()
	writeTestFile(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: dev
spec:
  template:
    source: git::https://user:secret@example.com/templates.git?access_token=secret//base
    ref: main
`)
	templateRoot := filepath.Join(t.TempDir(), "base")
	writeLocalTemplate(t, templateRoot)
	cacheRoot := t.TempDir()
	materialized, err := source.MaterializeLocal(templateRoot, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.Repeat("a", 40)
	configuredGit := filepath.Join(t.TempDir(), "git")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	writeTestFile(t, configPath, `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
executables:
  git: `+configuredGit+`
`)
	options := app.TemplateLockOptions{
		WorkingDirectory: deployment, CacheDirectory: cacheRoot,
		HomeDirectory: t.TempDir(), GOOS: "linux",
		RunDirectory: filepath.Join(t.TempDir(), "runs"), Environment: map[string]string{},
		AcquireGit: func(
			_ context.Context, reference source.Reference, selectedCache, gitPath string,
		) (source.GitAcquisition, error) {
			if reference.RequestedRef != "main" || selectedCache != cacheRoot ||
				gitPath != configuredGit || !strings.Contains(reference.Repository, "secret") {
				t.Fatalf("unexpected acquisition: %#v cache=%q git=%q", reference, selectedCache, gitPath)
			}
			return source.GitAcquisition{Commit: commit, Materialized: materialized}, nil
		},
	}
	request := app.TemplateLockRequest{ConfigPath: configPath}
	if _, err := app.TemplateLock(request, options); err != nil {
		t.Fatal(err)
	}
	first, err := lockfile.Read(filepath.Join(deployment, lockfile.Filename))
	if err != nil {
		t.Fatal(err)
	}
	if first.Template.RequestedRef != "main" || first.Template.Resolved != commit ||
		first.Template.Subdirectory != "base" {
		t.Fatalf("unexpected Git lock: %#v", first)
	}
	if strings.Contains(first.Template.Source, "secret") ||
		!strings.Contains(first.Template.Source, "redacted") {
		t.Fatalf("Git lock persisted credentials: %#v", first.Template)
	}
	commit = strings.Repeat("b", 40)
	updated, err := app.TemplateUpdate(request, options)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Changed || updated.ResolvedRevision != commit {
		t.Fatalf("unexpected update: %#v", updated)
	}
}

func TestTemplateLockUsesTrustedExplicitConfiguration(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	mustMkdir(t, filepath.Join(repository, ".git"))
	deployment := filepath.Join(repository, "deployment")
	templateRoot := filepath.Join(repository, "base")
	mustMkdir(t, deployment)
	writeTestFile(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: configured
spec:
  template:
    source: local:../base
`)
	writeLocalTemplate(t, templateRoot)
	configuredCache := filepath.Join(t.TempDir(), "configured-cache")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	writeTestFile(t, configPath, `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
paths:
  cache: `+configuredCache+`
`)
	_, err := app.TemplateLock(app.TemplateLockRequest{
		Target: deployment, ConfigPath: configPath,
	}, app.TemplateLockOptions{
		WorkingDirectory: repository, HomeDirectory: t.TempDir(), GOOS: "linux",
		CacheDirectory: filepath.Join(t.TempDir(), "default-cache"),
		RunDirectory:   filepath.Join(t.TempDir(), "runs"), Environment: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(configuredCache, "templates", "sha256"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("configured cache entries=%d err=%v", len(entries), err)
	}
}

func TestTemplateLockRejectsProjectControlledCache(t *testing.T) {
	t.Parallel()
	repository := t.TempDir()
	mustMkdir(t, filepath.Join(repository, ".git"))
	deployment := filepath.Join(repository, "deployment")
	templateRoot := filepath.Join(repository, "base")
	mustMkdir(t, deployment)
	writeTestFile(t, filepath.Join(deployment, "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: rejected
spec:
  template:
    source: local:../base
`)
	writeLocalTemplate(t, templateRoot)
	writeTestFile(t, filepath.Join(deployment, "ainfra.config.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
paths:
  cache: /tmp/project-controlled
`)
	_, err := app.TemplateLock(app.TemplateLockRequest{Target: deployment}, app.TemplateLockOptions{
		WorkingDirectory: repository, HomeDirectory: t.TempDir(), GOOS: "linux",
		CacheDirectory: filepath.Join(t.TempDir(), "cache"),
		RunDirectory:   filepath.Join(t.TempDir(), "runs"), Environment: map[string]string{},
	})
	if err == nil || !strings.Contains(err.Error(), "prohibited keys") {
		t.Fatalf("project cache error = %v", err)
	}
}

func writeLocalTemplate(t *testing.T, root string) {
	t.Helper()
	for _, directory := range []string{
		root, filepath.Join(root, "tofu"), filepath.Join(root, "docs"),
		filepath.Join(root, "examples", "minimal"),
	} {
		mustMkdir(t, directory)
	}
	writeTestFile(t, filepath.Join(root, "ainfra-template.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Template
metadata:
  name: base
  version: 1.2.3
spec:
  engines:
    tofu:
      directory: tofu
      version: ">= 1.9.0"
  outputs:
    inventory: none
`)
	writeTestFile(t, filepath.Join(root, "README.md"), "run tofu plan\n")
	writeTestFile(t, filepath.Join(root, "docs", "variables.md"), "# Variables\n")
	writeTestFile(t, filepath.Join(root, "tofu", "main.tf"), "terraform {}\n")
	writeTestFile(t, filepath.Join(root, "examples", "minimal", "ainfra.yaml"), `apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: dev
spec:
  template:
    source: local:../..
`)
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
