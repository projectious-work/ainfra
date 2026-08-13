// Package lock owns canonical parsing and atomic publication of ainfra.lock.
package lock

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.yaml.in/yaml/v3"
)

const (
	// Filename is the canonical deployment lockfile basename.
	Filename   = "ainfra.lock"
	apiVersion = "ainfra.projectious.work/v1"
	kind       = "TemplateLock"
)

// Template records the complete immutable template binding.
type Template struct {
	Source       string    `yaml:"source"`
	RequestedRef string    `yaml:"requestedRef,omitempty"`
	Resolved     string    `yaml:"resolved"`
	Subdirectory string    `yaml:"subdirectory"`
	Version      string    `yaml:"version"`
	Digest       string    `yaml:"digest"`
	ResolvedAt   time.Time `yaml:"resolvedAt"`
}

// Document is the v1 template lock contract.
type Document struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Template   Template `yaml:"template"`
}

// New constructs a canonical v1 lock document.
func New(template Template) Document {
	template.ResolvedAt = template.ResolvedAt.UTC().Truncate(time.Second)
	return Document{APIVersion: apiVersion, Kind: kind, Template: template}
}

// Read strictly parses one lockfile.
func Read(path string) (Document, error) {
	directory, name := filepath.Dir(path), filepath.Base(path)
	root, err := os.OpenRoot(directory)
	if err != nil {
		return Document{}, fmt.Errorf("open template lock root: %w", err)
	}
	contents, readErr := root.ReadFile(name)
	closeErr := root.Close()
	err = errors.Join(readErr, closeErr)
	if err != nil {
		return Document{}, fmt.Errorf("read template lock: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("parse template lock: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Document{}, errors.New("template lock must contain exactly one document")
		}
		return Document{}, fmt.Errorf("parse trailing template lock document: %w", err)
	}
	if document.APIVersion != apiVersion || document.Kind != kind {
		return Document{}, errors.New("unsupported template lock contract")
	}
	if document.Template.Source == "" || document.Template.Resolved == "" ||
		document.Template.Version == "" || document.Template.Digest == "" ||
		document.Template.ResolvedAt.IsZero() {
		return Document{}, errors.New("template lock is incomplete")
	}
	return document, nil
}

// Equivalent reports whether an existing binding already matches a candidate.
// Resolution time is provenance and does not make unchanged content different.
func Equivalent(left, right Document) bool {
	return left.APIVersion == right.APIVersion && left.Kind == right.Kind &&
		left.Template.Source == right.Template.Source &&
		left.Template.RequestedRef == right.Template.RequestedRef &&
		left.Template.Resolved == right.Template.Resolved &&
		left.Template.Subdirectory == right.Template.Subdirectory &&
		left.Template.Version == right.Template.Version &&
		left.Template.Digest == right.Template.Digest
}

// Write atomically publishes a private canonical lockfile while excluding a
// concurrent writer. The caller decides whether replacing an existing lock is
// permitted.
func Write(path string, document Document) error {
	directory, name := filepath.Dir(path), filepath.Base(path)
	root, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open template lock root: %w", err)
	}
	defer func() { _ = root.Close() }()
	writerName := name + ".writer"
	writer, err := root.OpenFile(writerName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("acquire template lock writer: %w", err)
	}
	if closeErr := writer.Close(); closeErr != nil {
		_ = root.Remove(writerName)
		return fmt.Errorf("close template lock writer: %w", closeErr)
	}
	defer func() { _ = root.Remove(writerName) }()

	contents, err := yaml.Marshal(document)
	if err != nil {
		return fmt.Errorf("serialize template lock: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".ainfra-lock-")
	if err != nil {
		return fmt.Errorf("create template lock staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure template lock staging file: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write template lock staging file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync template lock staging file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close template lock staging file: %w", err)
	}
	if err := root.Rename(filepath.Base(temporaryPath), name); err != nil {
		return fmt.Errorf("publish template lock: %w", err)
	}
	return nil
}
