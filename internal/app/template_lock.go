package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/source"
	"github.com/projectious-work/ainfra/internal/template"
)

// TemplateLockRequest selects a deployment for initial local template locking.
type TemplateLockRequest struct {
	Target      string
	ProjectPath string
}

// TemplateLockOptions supplies explicit host policy and time facts.
type TemplateLockOptions struct {
	WorkingDirectory string
	CacheDirectory   string
	Environment      map[string]string
	Now              func() time.Time
}

// TemplateLock resolves, validates, materializes, and locks one local source.
// Git sources remain unavailable until the Phase 3 Git adapter is composed.
func TemplateLock(request TemplateLockRequest, options TemplateLockOptions) (output.Template, error) {
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return output.Template{}, errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: options.WorkingDirectory, ExplicitPath: request.Target,
		ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath,
	})
	if err != nil {
		return output.Template{}, fmt.Errorf("load deployment contract: %w", err)
	}
	reference, err := source.Parse(deployment.Template.Source, deployment.Template.Ref)
	if err != nil {
		return output.Template{}, fmt.Errorf("parse template source: %w", err)
	}
	if reference.Kind != source.KindLocal {
		return output.Template{}, errors.New("git template locking requires the Phase 3 Git adapter")
	}
	resolved, err := source.ResolveLocal(reference, deployment.Target.Root)
	if err != nil {
		return output.Template{}, fmt.Errorf("resolve local template source: %w", err)
	}
	materialized, err := source.MaterializeLocal(resolved.Path, options.CacheDirectory)
	if err != nil {
		return output.Template{}, fmt.Errorf("materialize local template source: %w", err)
	}
	contract, err := template.LoadMaterialized(materialized.Path)
	if err != nil {
		return output.Template{}, fmt.Errorf("validate materialized template: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	document := lockfile.New(lockfile.Template{
		Source: reference.Canonical, Resolved: reference.Canonical,
		Subdirectory: reference.Subdirectory, Version: contract.Version,
		Digest: materialized.Digest, ResolvedAt: now(),
	})
	lockPath := filepath.Join(deployment.Target.Root, lockfile.Filename)
	if existing, readErr := lockfile.Read(lockPath); readErr == nil {
		if lockfile.Equivalent(existing, document) {
			return templateLockResult(document, false), nil
		}
		return output.Template{}, errors.New(
			"template lock already exists with a different binding; use 'ainfra template update'",
		)
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return output.Template{}, readErr
	}
	if err := lockfile.Write(lockPath, document); err != nil {
		return output.Template{}, err
	}
	return templateLockResult(document, true), nil
}

func templateLockResult(document lockfile.Document, changed bool) output.Template {
	return output.Template{
		Source: document.Template.Source, ResolvedRevision: document.Template.Resolved,
		ContentDigest: document.Template.Digest, Changed: changed,
	}
}

// HostTemplateLockOptions reads the closed host facts needed by local locking.
func HostTemplateLockOptions() (TemplateLockOptions, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return TemplateLockOptions{}, fmt.Errorf("read working directory: %w", err)
	}
	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return TemplateLockOptions{}, fmt.Errorf("read cache directory: %w", err)
	}
	return TemplateLockOptions{
		WorkingDirectory: workingDirectory,
		CacheDirectory:   filepath.Join(cacheDirectory, "ainfra"),
		Environment:      map[string]string{"AINFRA_PROJECT": os.Getenv("AINFRA_PROJECT")},
	}, nil
}
