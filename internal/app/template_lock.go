package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
	"github.com/projectious-work/ainfra/internal/template"
)

// TemplateLockRequest selects a deployment for initial local template locking.
type TemplateLockRequest struct {
	Target      string
	ProjectPath string
	ConfigPath  string
}

// TemplateLockOptions supplies explicit host policy and time facts.
type TemplateLockOptions struct {
	WorkingDirectory string
	CacheDirectory   string
	Environment      map[string]string
	Now              func() time.Time
	GOOS             string
	HomeDirectory    string
	XDGConfigHome    string
	RunDirectory     string
	GitPath          string
	AcquireGit       func(context.Context, source.Reference, string, string) (source.GitAcquisition, error)
}

// TemplateLock resolves, validates, materializes, and locks one local source.
// Git sources remain unavailable until the Phase 3 Git adapter is composed.
func TemplateLock(request TemplateLockRequest, options TemplateLockOptions) (output.Template, error) {
	return mutateTemplateLock(context.Background(), request, options, false, true)
}

// TemplateUpdate explicitly replaces an existing template binding after
// resolving and validating the currently requested source.
func TemplateUpdate(request TemplateLockRequest, options TemplateLockOptions) (output.Template, error) {
	return mutateTemplateLock(context.Background(), request, options, true, true)
}

func mutateTemplateLock(
	ctx context.Context,
	request TemplateLockRequest,
	options TemplateLockOptions,
	allowUpdate bool,
	publish bool,
) (output.Template, error) {
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
	options, err = resolveTemplateConfiguration(request, deployment.Target.Root, options)
	if err != nil {
		return output.Template{}, fmt.Errorf("resolve template configuration: %w", err)
	}
	return resolveTemplateLockMutation(ctx, deployment, options, allowUpdate, publish)
}

func resolveTemplateLockMutation(ctx context.Context, deployment project.Deployment,
	options TemplateLockOptions, allowUpdate, publish bool,
) (output.Template, error) {
	result, document, err := planTemplateLockMutation(ctx, deployment, options, allowUpdate)
	if err != nil {
		return output.Template{}, err
	}
	if !publish || !result.Changed {
		return result, nil
	}
	if err := lockfile.Write(filepath.Join(deployment.Target.Root, lockfile.Filename), document); err != nil {
		return output.Template{}, err
	}
	return result, nil
}

func planTemplateLockMutation(ctx context.Context, deployment project.Deployment,
	options TemplateLockOptions, allowUpdate bool,
) (output.Template, lockfile.Document, error) {
	reference, err := source.Parse(deployment.Template.Source, deployment.Template.Ref)
	if err != nil {
		return output.Template{}, lockfile.Document{}, fmt.Errorf("parse template source: %w", err)
	}
	materialized, immutable, err := resolveTemplate(ctx, reference, deployment.Target.Root, options)
	if err != nil {
		return output.Template{}, lockfile.Document{}, err
	}
	contract, err := template.LoadMaterialized(materialized.Path)
	if err != nil {
		return output.Template{}, lockfile.Document{}, fmt.Errorf("validate materialized template: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	document := lockfile.New(lockfile.Template{
		Source: reference.Display, RequestedRef: reference.RequestedRef,
		Resolved:     immutable,
		Subdirectory: reference.Subdirectory, Version: contract.Version,
		Digest: materialized.Digest, ResolvedAt: now(),
	})
	lockPath := filepath.Join(deployment.Target.Root, lockfile.Filename)
	if existing, readErr := lockfile.Read(lockPath); readErr == nil {
		if lockfile.Equivalent(existing, document) {
			return templateLockResult(document, false), document, nil
		}
		if !allowUpdate {
			return output.Template{}, lockfile.Document{}, errors.New(
				"template lock already exists with a different binding; use 'ainfra template update'",
			)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return output.Template{}, lockfile.Document{}, readErr
	} else if allowUpdate {
		return output.Template{}, lockfile.Document{}, errors.New("template update requires an existing ainfra.lock")
	}
	return templateLockResult(document, true), document, nil
}

func resolveTemplate(
	ctx context.Context,
	reference source.Reference,
	deploymentRoot string,
	options TemplateLockOptions,
) (source.Materialized, string, error) {
	if reference.Kind == source.KindLocal {
		resolved, err := source.ResolveLocal(reference, deploymentRoot)
		if err != nil {
			return source.Materialized{}, "", fmt.Errorf("resolve local template source: %w", err)
		}
		materialized, err := source.MaterializeLocal(resolved.Path, options.CacheDirectory)
		if err != nil {
			return source.Materialized{}, "", fmt.Errorf("materialize local template source: %w", err)
		}
		return materialized, reference.Canonical, nil
	}
	if options.AcquireGit == nil {
		return source.Materialized{}, "", errors.New("git template acquisition is unavailable")
	}
	acquired, err := options.AcquireGit(
		ctx, reference, options.CacheDirectory, options.GitPath,
	)
	if err != nil {
		return source.Materialized{}, "", fmt.Errorf("acquire Git template source: %w", err)
	}
	return acquired.Materialized, acquired.Commit, nil
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
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return TemplateLockOptions{}, fmt.Errorf("read home directory: %w", err)
	}
	return TemplateLockOptions{
		WorkingDirectory: workingDirectory,
		CacheDirectory:   filepath.Join(cacheDirectory, "ainfra"),
		RunDirectory:     filepath.Join(homeDirectory, ".local", "state", "ainfra", "runs"),
		HomeDirectory:    homeDirectory,
		XDGConfigHome:    os.Getenv("XDG_CONFIG_HOME"),
		GOOS:             runtime.GOOS,
		Environment:      supportedEnvironment(),
		AcquireGit:       acquireHostGit,
	}, nil
}

func acquireHostGit(
	ctx context.Context,
	reference source.Reference,
	cacheRoot string,
	configuredPath string,
) (source.GitAcquisition, error) {
	gitPath := configuredPath
	if gitPath == "" {
		var err error
		gitPath, err = exec.LookPath("git")
		if err != nil {
			return source.GitAcquisition{}, fmt.Errorf("discover git executable: %w", err)
		}
	}
	executable, err := security.ResolveExecutable(gitPath)
	if err != nil {
		return source.GitAcquisition{}, fmt.Errorf("validate git executable: %w", err)
	}
	environment, err := security.BuildEnvironment(
		os.Environ(), []string{"HOME", "PATH", "SSH_AUTH_SOCK"},
		map[string]string{"GIT_TERMINAL_PROMPT": "0"},
	)
	if err != nil {
		return source.GitAcquisition{}, fmt.Errorf("build git environment: %w", err)
	}
	return source.AcquireGit(ctx, reference, source.GitOptions{
		Executable: executable, Environment: environment, CacheRoot: cacheRoot,
	})
}

func resolveTemplateConfiguration(
	request TemplateLockRequest,
	deploymentRoot string,
	options TemplateLockOptions,
) (TemplateLockOptions, error) {
	if options.HomeDirectory == "" {
		if request.ConfigPath != "" {
			return TemplateLockOptions{}, errors.New("home directory is required for explicit configuration")
		}
		return options, nil
	}
	explicit := request.ConfigPath
	if explicit == "" {
		explicit = options.Environment["AINFRA_CONFIG"]
	}
	if explicit != "" && !filepath.IsAbs(explicit) {
		explicit = filepath.Join(options.WorkingDirectory, explicit)
	}
	files, err := config.Files(config.LocationOptions{
		GOOS: options.GOOS, HomeDirectory: options.HomeDirectory,
		XDGConfigHome: options.XDGConfigHome, DeploymentRoot: deploymentRoot,
		ExplicitPath: explicit,
	})
	if err != nil {
		return TemplateLockOptions{}, err
	}
	effective, err := config.Resolve(config.ResolveOptions{
		Defaults: config.Defaults(options.CacheDirectory, options.RunDirectory),
		Files:    files, Environment: options.Environment,
	})
	if err != nil {
		return TemplateLockOptions{}, err
	}
	options.CacheDirectory = effective.Settings.Paths.Cache
	options.GitPath = effective.Settings.Executables.Git
	return options, nil
}
