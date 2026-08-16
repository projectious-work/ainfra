package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/template"
	"github.com/projectious-work/ainfra/internal/tofu"
)

// PlanOptions supplies already trusted policy outputs for saved planning.
type PlanOptions struct {
	Prepare       runstate.Options
	Template      template.Contract
	Adapter       tofu.Adapter
	EngineVersion string
	CreatedAt     time.Time
	Destroy       bool
}

// CreateApplyPlan prepares immutable inputs, invokes OpenTofu, sanitizes its
// structural summary, and publishes the reviewed plan record. It never applies.
func CreateApplyPlan(ctx context.Context, options PlanOptions) (runstate.PlanRecord, error) {
	prepared, err := runstate.Prepare(options.Prepare)
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	published := false
	defer func() {
		if !published {
			_ = runstate.Discard(prepared)
		}
	}()
	engineDirectory := filepath.Join("workspace", filepath.FromSlash(options.Template.Tofu.Directory))
	backend, variables, err := snapshotArguments(prepared, options.Prepare.Deployment, engineDirectory)
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	if _, err := options.Adapter.Init(ctx, prepared.Root, engineDirectory, backend); err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("initialize OpenTofu: %w", err)
	}
	planPath, err := filepath.Rel(filepath.Join(prepared.Root, engineDirectory), filepath.Join(prepared.Root, "plan.tfplan"))
	if err != nil {
		return runstate.PlanRecord{}, err
	}
	if _, err := options.Adapter.Plan(ctx, prepared.Root, engineDirectory, planPath, variables, options.Destroy); err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("create saved OpenTofu plan: %w", err)
	}
	summary, _, err := options.Adapter.ShowSummary(ctx, prepared.Root, engineDirectory, planPath)
	if err != nil {
		return runstate.PlanRecord{}, fmt.Errorf("summarize saved OpenTofu plan: %w", err)
	}
	intent := "apply"
	if options.Destroy {
		intent = "destroy"
	}
	record, err := runstate.PublishPlanIntent(prepared, options.Prepare.Deployment.Metadata.Name, options.EngineVersion, intent, options.CreatedAt, summary)
	if err == nil {
		published = true
	}
	return record, err
}

// PlanRequest selects a deployment and trusted configuration for planning.
type PlanRequest struct {
	Target, ProjectPath, ConfigPath string
	Destroy                         bool
}

// PlanHostOptions supplies host facts while keeping Plan testable.
type PlanHostOptions struct {
	WorkingDirectory, CacheDirectory, RunDirectory string
	HomeDirectory, XDGConfigHome, GOOS             string
	Environment                                    map[string]string
	ParentEnvironment                              []string
	Now                                            func() time.Time
	Random                                         func([]byte) (int, error)
}

// Plan resolves trusted inputs and creates a saved, never-applied plan.
func Plan(ctx context.Context, request PlanRequest, options PlanHostOptions) (output.Plan, error) {
	environmentPath := options.Environment["AINFRA_PROJECT"]
	if request.Target != "" && (request.ProjectPath != "" || environmentPath != "") {
		return output.Plan{}, errors.New("deployment TARGET conflicts with another project selection")
	}
	deployment, err := project.Load(project.ResolveOptions{WorkingDirectory: options.WorkingDirectory,
		ExplicitPath: request.Target, ProjectPath: request.ProjectPath, EnvironmentPath: environmentPath})
	if err != nil {
		return output.Plan{}, fmt.Errorf("load deployment contract: %w", err)
	}
	settings, err := resolvePlanConfiguration(request, deployment.Target.Root, options)
	if err != nil {
		return output.Plan{}, fmt.Errorf("resolve plan configuration: %w", err)
	}
	return planForDeployment(ctx, deployment, settings, request.Destroy, options)
}

func planForDeployment(ctx context.Context, deployment project.Deployment,
	settings config.Settings, destroy bool, options PlanHostOptions,
) (output.Plan, error) {
	lock, err := lockfile.Read(filepath.Join(deployment.Target.Root, lockfile.Filename))
	if err != nil {
		return output.Plan{}, fmt.Errorf("read template lock: %w", err)
	}
	cachePath := filepath.Join(settings.Paths.Cache, "templates", "sha256",
		strings.TrimPrefix(lock.Template.Digest, "sha256:"))
	contract, err := template.LoadMaterialized(cachePath)
	if err != nil {
		return output.Plan{}, fmt.Errorf("load locked template: %w", err)
	}
	tofuPath := settings.Executables.Tofu
	if tofuPath == "" {
		tofuPath, err = exec.LookPath("tofu")
		if err != nil {
			return output.Plan{}, fmt.Errorf("discover OpenTofu executable: %w", err)
		}
	}
	executable, err := security.ResolveExecutable(tofuPath)
	if err != nil {
		return output.Plan{}, fmt.Errorf("validate OpenTofu executable: %w", err)
	}
	environment, err := security.BuildEnvironment(options.ParentEnvironment,
		[]string{"HOME", "PATH", "SSL_CERT_DIR", "SSL_CERT_FILE"},
		map[string]string{"TF_IN_AUTOMATION": "1"})
	if err != nil {
		return output.Plan{}, fmt.Errorf("build OpenTofu environment: %w", err)
	}
	adapter := tofu.Adapter{Executable: executable, Environment: environment}
	version, err := adapter.Version(ctx, deployment.Target.Root)
	if err != nil {
		return output.Plan{}, fmt.Errorf("read OpenTofu version: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	random := rand.Read
	if options.Random != nil {
		random = options.Random
	}
	id, err := newRunID(now(), random)
	if err != nil {
		return output.Plan{}, fmt.Errorf("create run ID: %w", err)
	}
	record, err := CreateApplyPlan(ctx, PlanOptions{Prepare: runstate.Options{ID: id,
		RunsRoot: settings.Paths.Runs, CacheRoot: settings.Paths.Cache,
		Deployment: deployment, Lock: lock, Executable: executable}, Template: contract,
		Adapter: adapter, EngineVersion: version, CreatedAt: now(), Destroy: destroy})
	if err != nil {
		return output.Plan{}, err
	}
	intent := "apply"
	if destroy {
		intent = "destroy"
	}
	nextCommand := "ainfra apply "
	if destroy {
		nextCommand = "ainfra destroy "
	}
	successExitCode := 0
	return output.Plan{Deployment: output.Deployment{Name: deployment.Metadata.Name, Root: deployment.Target.Root},
		RunID: record.RunID, Intent: intent, PlanDigest: record.Plan.Digest,
		TemplateDigest: record.Template.Digest, InputDigest: planInputDigest(record),
		EngineReport: output.EngineReport{Engine: "opentofu", Status: "succeeded", ExitCode: &successExitCode,
			Protocol: output.Protocol{Name: "opentofu-json-ui", Version: "1.2"}},
		Evidence:     []output.Evidence{{Kind: "plan-summary", Path: "plan.json", Sensitive: false}},
		NextCommands: []string{nextCommand + deployment.Metadata.Name + " --plan " + record.RunID}}, nil
}

func planInputDigest(record runstate.PlanRecord) string {
	contents, _ := json.Marshal(struct {
		Deployment runstate.NamedBinding   `json:"deployment"`
		Inputs     []runstate.InputBinding `json:"inputs"`
	}{record.Deployment, record.Inputs})
	digest := sha256.Sum256(contents)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func resolvePlanConfiguration(request PlanRequest, root string, options PlanHostOptions) (config.Settings, error) {
	explicit := request.ConfigPath
	if explicit == "" {
		explicit = options.Environment["AINFRA_CONFIG"]
	}
	if explicit != "" && !filepath.IsAbs(explicit) {
		explicit = filepath.Join(options.WorkingDirectory, explicit)
	}
	files, err := config.Files(config.LocationOptions{GOOS: options.GOOS, HomeDirectory: options.HomeDirectory,
		XDGConfigHome: options.XDGConfigHome, DeploymentRoot: root, ExplicitPath: explicit})
	if err != nil {
		return config.Settings{}, err
	}
	effective, err := config.Resolve(config.ResolveOptions{Defaults: config.Defaults(options.CacheDirectory,
		options.RunDirectory), Files: files, Environment: options.Environment})
	if err != nil {
		return config.Settings{}, err
	}
	return effective.Settings, nil
}

func newRunID(now time.Time, read func([]byte) (int, error)) (string, error) {
	bytes := make([]byte, 16)
	if n, err := read(bytes); err != nil || n != len(bytes) {
		return "", errors.New("read complete random run identifier")
	}
	return now.UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(bytes), nil
}

// HostPlanOptions reads the closed host facts needed by local planning.
func HostPlanOptions() (PlanHostOptions, error) {
	working, err := os.Getwd()
	if err != nil {
		return PlanHostOptions{}, err
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return PlanHostOptions{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return PlanHostOptions{}, err
	}
	return PlanHostOptions{WorkingDirectory: working, CacheDirectory: filepath.Join(cache, "ainfra"),
		RunDirectory: filepath.Join(home, ".local", "state", "ainfra", "runs"), HomeDirectory: home,
		XDGConfigHome: os.Getenv("XDG_CONFIG_HOME"), GOOS: runtime.GOOS,
		Environment: supportedEnvironment(), ParentEnvironment: os.Environ()}, nil
}

func snapshotArguments(prepared runstate.Prepared, deployment project.Deployment, engineDirectory string) ([]string, []string, error) {
	lookup := make(map[string]string, len(prepared.NativeInputs))
	for _, binding := range prepared.NativeInputs {
		lookup[binding.Path] = binding.SnapshotPath
	}
	convert := func(values []string) ([]string, error) {
		arguments := make([]string, len(values))
		for index, value := range values {
			snapshot, ok := lookup[filepath.ToSlash(value)]
			if !ok {
				return nil, fmt.Errorf("native input snapshot missing for %q", value)
			}
			relative, err := filepath.Rel(filepath.Join(prepared.Root, engineDirectory), filepath.Join(prepared.Root, filepath.FromSlash(snapshot)))
			if err != nil {
				return nil, err
			}
			arguments[index] = relative
		}
		return arguments, nil
	}
	backend, err := convert(deployment.Inputs.TofuBackendConfigFiles)
	if err != nil {
		return nil, nil, err
	}
	variables, err := convert(deployment.Inputs.TofuVariableFiles)
	return backend, variables, err
}
