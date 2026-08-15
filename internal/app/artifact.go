package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	"github.com/projectious-work/ainfra/internal/inventory"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/template"
	"github.com/projectious-work/ainfra/internal/tofu"
)

// ArtifactRequest selects one successfully applied retained run.
type ArtifactRequest struct {
	Target, ProjectPath, ConfigPath, RunID string
}

type artifactContext struct {
	deployment project.Deployment
	contract   template.Contract
	reviewed   runstate.Reviewed
	adapter    tofu.Adapter
	settings   config.Settings
}

// CollectOutput validates and atomically persists the declared standardized
// output. It never persists any other provider output.
func CollectOutput(ctx context.Context, request ArtifactRequest, options PlanHostOptions) (output.Artifact, error) {
	deployment, contract, err := resolveArtifactApplicability(request, options)
	if err != nil {
		return output.Artifact{}, err
	}
	if contract.Inventory == "none" {
		return notApplicableArtifact(deployment, request.RunID, "template declares inventory mode none"), nil
	}
	resolved, unlock, err := resolveLockedArtifactContext(ctx, request, options)
	if err != nil {
		return output.Artifact{}, err
	}
	defer func() { _ = unlock() }()
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	if err := runstate.BeginOperation(resolved.reviewed, "output", now()); err != nil {
		return output.Artifact{}, err
	}
	directory := filepath.Join("workspace", filepath.FromSlash(resolved.contract.Tofu.Directory))
	contents, _, err := resolved.adapter.Output(ctx, resolved.reviewed.Root, directory, resolved.contract.Inventory)
	if err != nil {
		return failArtifact(resolved.reviewed, "output", now(), fmt.Errorf("collect standardized OpenTofu output: %w", err))
	}
	standard, err := inventory.Parse(contents)
	if err != nil {
		return failArtifact(resolved.reviewed, "output", now(), err)
	}
	canonical, err := json.MarshalIndent(standard, "", "  ")
	if err != nil {
		return failArtifact(resolved.reviewed, "output", now(), err)
	}
	canonical = append(canonical, '\n')
	digest, err := runstate.PublishArtifact(resolved.reviewed.Root, "output.json", canonical)
	if err != nil {
		return failArtifact(resolved.reviewed, "output", now(), err)
	}
	if err := runstate.AppendOperationEvent(resolved.reviewed, "output", "succeeded", now(), nil); err != nil {
		return output.Artifact{}, err
	}
	return artifactResult(resolved, request.RunID, "standard-output", "output.json", digest), nil
}

// GenerateInventory transforms only retained validated output into stable YAML.
func GenerateInventory(ctx context.Context, request ArtifactRequest, options PlanHostOptions) (output.Artifact, error) {
	deployment, contract, err := resolveArtifactApplicability(request, options)
	if err != nil {
		return output.Artifact{}, err
	}
	if contract.Inventory == "none" {
		return notApplicableArtifact(deployment, request.RunID, "template declares inventory mode none"), nil
	}
	resolved, unlock, err := resolveLockedArtifactContext(ctx, request, options)
	if err != nil {
		return output.Artifact{}, err
	}
	defer func() { _ = unlock() }()
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	if err := runstate.BeginOperation(resolved.reviewed, "inventory", now()); err != nil {
		return output.Artifact{}, err
	}
	contents, err := readPrivateArtifact(resolved.reviewed.Root, "output.json", 16<<20)
	if err != nil {
		return failArtifact(resolved.reviewed, "inventory", now(), fmt.Errorf("read validated output: %w", err))
	}
	standard, err := inventory.Parse(contents)
	if err != nil {
		return failArtifact(resolved.reviewed, "inventory", now(), fmt.Errorf("revalidate retained output: %w", err))
	}
	yamlContents, err := inventory.YAML(standard)
	if err != nil {
		return failArtifact(resolved.reviewed, "inventory", now(), err)
	}
	digest, err := runstate.PublishArtifact(resolved.reviewed.Root, "inventory.yaml", yamlContents)
	if err != nil {
		return failArtifact(resolved.reviewed, "inventory", now(), err)
	}
	if err := runstate.AppendOperationEvent(resolved.reviewed, "inventory", "succeeded", now(), nil); err != nil {
		return output.Artifact{}, err
	}
	return artifactResult(resolved, request.RunID, "inventory", "inventory.yaml", digest), nil
}

func failArtifact(reviewed runstate.Reviewed, operation string, at time.Time, cause error) (output.Artifact, error) {
	eventErr := runstate.AppendOperationEvent(reviewed, operation, "failed", at, nil)
	return output.Artifact{}, errors.Join(cause, eventErr)
}

func resolveLockedArtifactContext(ctx context.Context, request ArtifactRequest, options PlanHostOptions) (artifactContext, func() error, error) {
	deployment, _, err := resolveArtifactApplicability(request, options)
	if err != nil {
		return artifactContext{}, nil, err
	}
	unlock, err := (reconcile.FileLocker{}).Lock(deployment.Target.Root, project.ManifestName)
	if err != nil {
		return artifactContext{}, nil, fmt.Errorf("acquire deployment operation lock: %w", err)
	}
	resolved, err := resolveArtifactContext(ctx, request, options)
	if err != nil {
		_ = unlock()
		return artifactContext{}, nil, fmt.Errorf("reverify run under operation lock: %w", err)
	}
	return resolved, unlock, nil
}

func resolveArtifactApplicability(request ArtifactRequest, options PlanHostOptions) (project.Deployment, template.Contract, error) {
	if request.RunID == "" {
		return project.Deployment{}, template.Contract{}, errors.New("--run requires an exact applied run ID")
	}
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return project.Deployment{}, template.Contract{}, errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return project.Deployment{}, template.Contract{}, fmt.Errorf("load deployment contract: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{ConfigPath: request.ConfigPath}, deployment.Target.Root, options)
	if err != nil {
		return project.Deployment{}, template.Contract{}, fmt.Errorf("resolve artifact configuration: %w", err)
	}
	runRoot, err := security.ResolveContained(settings.Paths.Runs, request.RunID)
	if err != nil {
		return project.Deployment{}, template.Contract{}, errors.New("retained run ID is invalid")
	}
	info, err := os.Lstat(runRoot)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return project.Deployment{}, template.Contract{}, errors.New("retained run is unavailable")
	}
	runRecord, err := security.ResolveContained(runRoot, "run.json")
	if err != nil || security.RequireRegular(runRecord) != nil {
		return project.Deployment{}, template.Contract{}, errors.New("retained run record is unavailable")
	}
	if err := runstate.RequireSuccessfulApply(runstate.Reviewed{Root: runRoot}); err != nil {
		return project.Deployment{}, template.Contract{}, err
	}
	lock, err := lockfile.Read(filepath.Join(deployment.Target.Root, lockfile.Filename))
	if err != nil {
		return project.Deployment{}, template.Contract{}, fmt.Errorf("read template lock: %w", err)
	}
	cachePath := filepath.Join(settings.Paths.Cache, "templates", "sha256", strings.TrimPrefix(lock.Template.Digest, "sha256:"))
	contract, err := template.LoadMaterialized(cachePath)
	if err != nil {
		return project.Deployment{}, template.Contract{}, fmt.Errorf("load locked template: %w", err)
	}
	return deployment, contract, nil
}

func resolveArtifactContext(ctx context.Context, request ArtifactRequest, options PlanHostOptions) (artifactContext, error) {
	if request.RunID == "" {
		return artifactContext{}, errors.New("--run requires an exact applied run ID")
	}
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return artifactContext{}, errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return artifactContext{}, fmt.Errorf("load deployment contract: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{ConfigPath: request.ConfigPath}, deployment.Target.Root, options)
	if err != nil {
		return artifactContext{}, fmt.Errorf("resolve artifact configuration: %w", err)
	}
	lock, err := lockfile.Read(filepath.Join(deployment.Target.Root, lockfile.Filename))
	if err != nil {
		return artifactContext{}, fmt.Errorf("read template lock: %w", err)
	}
	cachePath := filepath.Join(settings.Paths.Cache, "templates", "sha256", strings.TrimPrefix(lock.Template.Digest, "sha256:"))
	contract, err := template.LoadMaterialized(cachePath)
	if err != nil {
		return artifactContext{}, fmt.Errorf("load locked template: %w", err)
	}
	tofuPath := settings.Executables.Tofu
	if tofuPath == "" {
		tofuPath, err = exec.LookPath("tofu")
	}
	if err != nil {
		return artifactContext{}, fmt.Errorf("discover OpenTofu executable: %w", err)
	}
	executable, err := security.ResolveExecutable(tofuPath)
	if err != nil {
		return artifactContext{}, fmt.Errorf("validate OpenTofu executable: %w", err)
	}
	environment, err := security.BuildEnvironment(options.ParentEnvironment,
		[]string{"HOME", "PATH", "SSL_CERT_DIR", "SSL_CERT_FILE"}, map[string]string{"TF_IN_AUTOMATION": "1"})
	if err != nil {
		return artifactContext{}, err
	}
	adapter := tofu.Adapter{Executable: executable, Environment: environment}
	version, err := adapter.Version(ctx, deployment.Target.Root)
	if err != nil {
		return artifactContext{}, fmt.Errorf("read OpenTofu version: %w", err)
	}
	reviewed, err := runstate.LoadReviewed(runstate.ReviewOptions{ID: request.RunID,
		RunsRoot: settings.Paths.Runs, CacheRoot: settings.Paths.Cache, Deployment: deployment,
		Lock: lock, Executable: executable, EngineVersion: version})
	if err != nil {
		return artifactContext{}, fmt.Errorf("reverify applied run: %w", err)
	}
	if err := runstate.RequireSuccessfulApply(reviewed); err != nil {
		return artifactContext{}, err
	}
	return artifactContext{deployment: deployment, contract: contract, reviewed: reviewed, adapter: adapter, settings: settings}, nil
}

func readPrivateArtifact(root, name string, limit int) ([]byte, error) {
	directory, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	privateRoot, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	pathInfo, lstatErr := privateRoot.Lstat(name)
	if lstatErr != nil {
		_ = privateRoot.Close()
		return nil, lstatErr
	}
	if !pathInfo.Mode().IsRegular() || pathInfo.Mode().Perm()&0o077 != 0 {
		_ = privateRoot.Close()
		return nil, errors.New("artifact is not a private regular file")
	}
	handle, openErr := privateRoot.Open(name)
	if openErr != nil {
		_ = privateRoot.Close()
		return nil, openErr
	}
	information, statErr := handle.Stat()
	if statErr != nil || !information.Mode().IsRegular() ||
		information.Mode().Perm()&0o077 != 0 || !os.SameFile(pathInfo, information) {
		_ = handle.Close()
		_ = privateRoot.Close()
		return nil, errors.New("artifact is not a private regular file")
	}
	contents, readErr := io.ReadAll(io.LimitReader(handle, int64(limit)+1))
	handleCloseErr := handle.Close()
	closeErr := privateRoot.Close()
	if err := errors.Join(readErr, handleCloseErr, closeErr); err != nil {
		return nil, err
	}
	if len(contents) == 0 || len(contents) > limit {
		return nil, errors.New("artifact has invalid size")
	}
	return contents, nil
}

func artifactResult(resolved artifactContext, runID, kind, path, digest string) output.Artifact {
	evidence := output.Evidence{Kind: kind, Path: path, Sensitive: false}
	return output.Artifact{Deployment: output.Deployment{Name: resolved.deployment.Metadata.Name,
		Root: resolved.deployment.Target.Root}, RunID: runID, Applicability: "applicable",
		Artifact: &evidence, ContentDigest: digest}
}

func notApplicableArtifact(deployment project.Deployment, runID, reason string) output.Artifact {
	return output.Artifact{Deployment: output.Deployment{Name: deployment.Metadata.Name,
		Root: deployment.Target.Root}, RunID: runID, Applicability: "not-applicable", Reason: reason}
}
