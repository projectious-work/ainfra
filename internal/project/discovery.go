// Package project discovers and identifies local ainfra deployments.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectious-work/ainfra/internal/security"
)

// ManifestName is the only deployment manifest basename discovered by ainfra.
const ManifestName = "ainfra.yaml"

// TargetSource records which normative precedence layer selected a deployment.
type TargetSource string

const (
	// SourceExplicit identifies a positional deployment target.
	SourceExplicit TargetSource = "explicit"
	// SourceProjectFlag identifies the global --project option.
	SourceProjectFlag TargetSource = "project-flag"
	// SourceEnvironment identifies AINFRA_PROJECT.
	SourceEnvironment TargetSource = "environment"
	// SourceAncestor identifies nearest-ancestor discovery from the working dir.
	SourceAncestor TargetSource = "ancestor"
)

// Target is the canonical identity of a discovered local deployment.
type Target struct {
	Root         string
	ManifestPath string
	Source       TargetSource
}

// ResolveOptions supplies immutable inputs to deployment discovery.
type ResolveOptions struct {
	WorkingDirectory string
	ExplicitPath     string
	ProjectPath      string
	EnvironmentPath  string
}

// Resolve applies the normative deployment-target precedence and returns one
// canonical root/manifest pair. It never searches descendants or global roots.
func Resolve(options ResolveOptions) (Target, error) {
	workingDirectory := options.WorkingDirectory
	if workingDirectory == "" {
		var err error
		workingDirectory, err = os.Getwd()
		if err != nil {
			return Target{}, fmt.Errorf("read working directory: %w", err)
		}
	}

	selections := []struct {
		path   string
		source TargetSource
	}{
		{path: options.ExplicitPath, source: SourceExplicit},
		{path: options.ProjectPath, source: SourceProjectFlag},
		{path: options.EnvironmentPath, source: SourceEnvironment},
	}
	for _, selection := range selections {
		if selection.path != "" {
			return resolveExplicit(workingDirectory, selection.path, selection.source)
		}
	}
	return discoverAncestor(workingDirectory)
}

func resolveExplicit(workingDirectory, candidate string, source TargetSource) (Target, error) {
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(workingDirectory, candidate)
	}
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return Target{}, fmt.Errorf("resolve deployment target: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return Target{}, fmt.Errorf("inspect deployment target: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Target{}, &security.Refusal{
			Policy: "symlink", Path: absolute,
			Reason: "deployment target must not be a symbolic link",
		}
	}

	manifestPath := absolute
	if info.IsDir() {
		manifestPath = filepath.Join(absolute, ManifestName)
	} else if filepath.Base(absolute) != ManifestName {
		return Target{}, fmt.Errorf("deployment target file must be named %s", ManifestName)
	}
	return canonicalTarget(manifestPath, source)
}

func discoverAncestor(workingDirectory string) (Target, error) {
	current, err := filepath.Abs(workingDirectory)
	if err != nil {
		return Target{}, fmt.Errorf("resolve working directory: %w", err)
	}
	for {
		manifestPath := filepath.Join(current, ManifestName)
		if _, statErr := os.Lstat(manifestPath); statErr == nil {
			return canonicalTarget(manifestPath, SourceAncestor)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return Target{}, fmt.Errorf("inspect deployment manifest: %w", statErr)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return Target{}, fmt.Errorf("no %s found in working directory ancestors", ManifestName)
		}
		current = parent
	}
}

func canonicalTarget(manifestPath string, source TargetSource) (Target, error) {
	if err := security.RequireRegular(manifestPath); err != nil {
		return Target{}, fmt.Errorf("validate deployment manifest: %w", err)
	}
	root := filepath.Dir(manifestPath)
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return Target{}, fmt.Errorf("canonicalize deployment root: %w", err)
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil {
		return Target{}, fmt.Errorf("resolve deployment root: %w", err)
	}
	canonicalManifest, err := security.ResolveContained(canonicalRoot, ManifestName)
	if err != nil {
		return Target{}, fmt.Errorf("contain deployment manifest: %w", err)
	}
	return Target{
		Root: canonicalRoot, ManifestPath: canonicalManifest, Source: source,
	}, nil
}
