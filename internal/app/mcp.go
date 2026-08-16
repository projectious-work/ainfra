package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
	runstate "github.com/projectious-work/ainfra/internal/run"
	"github.com/projectious-work/ainfra/internal/source"
	"github.com/projectious-work/ainfra/internal/template"
)

// MCPServeRequest contains server-start policy selected by the operator.
// Protocol-specific types belong to the MCP adapter and must not enter app.
type MCPServeRequest struct {
	ProjectPath        string
	Capabilities       []string
	AuthorizationTrust string
}

// MCPServeOptions supplies the closed host facts captured before serving.
type MCPServeOptions struct {
	Plan          PlanHostOptions
	Template      TemplateLockOptions
	Doctor        DoctorEnvironmentOptions
	Authorization MCPAuthorizationProvider
	Now           func() time.Time
}

// MCPServeSession contains the immutable, validated project identity exposed
// to protocol adapters. Request handlers cannot replace this root.
type MCPServeSession struct {
	Project           output.Deployment
	project           project.Deployment
	runsRoot          string
	cacheRoot         string
	settings          config.Settings
	planOptions       PlanHostOptions
	templateOptions   TemplateLockOptions
	environmentDoctor output.Doctor
	capabilities      map[string]struct{}
	authorization     MCPAuthorizationProvider
	now               func() time.Time
}

const (
	// MCPPlanningCapability exposes non-applying planning operations.
	MCPPlanningCapability = "planning"
	// MCPDeploymentCapability gates authorized lifecycle mutations.
	MCPDeploymentCapability = "deployment"
	// MCPDestructionCapability additionally gates destroy execution.
	MCPDestructionCapability = "destruction"
)

// MCPReconciliationAction is one stable, reviewable project-local repair.
type MCPReconciliationAction struct {
	Check              string `json:"check"`
	Path               string `json:"path"`
	Kind               string `json:"kind"`
	Mode               string `json:"mode"`
	RollbackLimitation string `json:"rollbackLimitation"`
}

// MCPReconciliationPlan is a non-applying preview for the startup project.
type MCPReconciliationPlan struct {
	Deployment output.Deployment         `json:"deployment"`
	Doctor     output.Doctor             `json:"doctor"`
	Actions    []MCPReconciliationAction `json:"actions"`
}

// MCPDeploymentContract is the sanitized semantic deployment contract fixed at
// server startup. Its paths are validated project-relative pointers only.
type MCPDeploymentContract struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Template    MCPDeploymentTemplate `json:"template"`
	Inputs      MCPDeploymentInputs   `json:"inputs"`
	SSH         MCPDeploymentSSH      `json:"ssh"`
}

type MCPDeploymentTemplate struct {
	Source string `json:"source"`
	Ref    string `json:"ref,omitempty"`
}

type MCPDeploymentInputs struct {
	TofuVariableFiles      []string `json:"tofuVariableFiles"`
	TofuBackendConfigFiles []string `json:"tofuBackendConfigFiles"`
	AnsibleVariableFiles   []string `json:"ansibleVariableFiles"`
}

type MCPDeploymentSSH struct {
	KnownHosts string `json:"knownHosts,omitempty"`
}

// MCPTemplateContract combines the verified immutable lock binding with the
// semantic template manifest. Cache and host filesystem paths are excluded.
type MCPTemplateContract struct {
	Source       string                    `json:"source"`
	RequestedRef string                    `json:"requestedRef,omitempty"`
	Resolved     string                    `json:"resolved"`
	Version      string                    `json:"version"`
	Digest       string                    `json:"digest"`
	Name         string                    `json:"name"`
	Tofu         MCPTemplateEngine         `json:"tofu"`
	Ansible      *MCPTemplateAnsibleEngine `json:"ansible,omitempty"`
	Inventory    string                    `json:"inventory"`
}

type MCPTemplateEngine struct {
	Directory string `json:"directory"`
	Version   string `json:"version"`
}

type MCPTemplateAnsibleEngine struct {
	MCPTemplateEngine
	Playbook string `json:"playbook"`
}

// MCPExecutionRequest carries only exact reviewed-plan and independent approval
// inputs. Project and host policy remain fixed by the serving session.
type MCPExecutionRequest struct {
	PlanID   string
	Caller   string
	Approval string
}

// MCPApplyRequest preserves the initial apply-adapter request name.
type MCPApplyRequest = MCPExecutionRequest

// MCPExecutionResult combines normal execution semantics with sanitized approval
// evidence. Opaque approval material is never retained or returned.
type MCPExecutionResult struct {
	Execution     output.Execution `json:"execution"`
	Authorization MCPAuthorization `json:"authorization"`
}

// MCPApplyResult preserves the initial apply-adapter result name.
type MCPApplyResult = MCPExecutionResult

// Status reads sanitized retained lifecycle state for the fixed startup
// project. It does not rediscover a root or reload configuration.
func (session MCPServeSession) Status() (output.Status, error) {
	return statusForDeployment(session.project, session.runsRoot)
}

// InspectDeployment returns a detached copy of the validated startup contract.
func (session MCPServeSession) InspectDeployment() MCPDeploymentContract {
	return MCPDeploymentContract{Name: session.project.Metadata.Name,
		Description: session.project.Metadata.Description,
		Template: MCPDeploymentTemplate{Source: session.project.Template.Source,
			Ref: session.project.Template.Ref},
		Inputs: MCPDeploymentInputs{
			TofuVariableFiles:      append([]string(nil), session.project.Inputs.TofuVariableFiles...),
			TofuBackendConfigFiles: append([]string(nil), session.project.Inputs.TofuBackendConfigFiles...),
			AnsibleVariableFiles:   append([]string(nil), session.project.Inputs.AnsibleVariableFiles...),
		}, SSH: MCPDeploymentSSH{KnownHosts: session.project.SSH.KnownHosts}}
}

// InspectTemplate verifies the startup deployment-to-lock binding and the
// digest-addressed private cache before returning its sanitized contract.
func (session MCPServeSession) InspectTemplate() (MCPTemplateContract, error) {
	document, err := lockfile.Read(filepath.Join(session.project.Target.Root, lockfile.Filename))
	if err != nil {
		return MCPTemplateContract{}, fmt.Errorf("read template lock: %w", err)
	}
	reference, err := source.Parse(session.project.Template.Source, session.project.Template.Ref)
	if err != nil || document.Template.Source != reference.Display ||
		document.Template.RequestedRef != reference.RequestedRef {
		return MCPTemplateContract{}, errors.New("template lock does not match deployment source binding")
	}
	cachePath := filepath.Join(session.cacheRoot, "templates", "sha256",
		strings.TrimPrefix(document.Template.Digest, "sha256:"))
	observed, err := source.TreeDigest(cachePath)
	if err != nil || observed != document.Template.Digest {
		return MCPTemplateContract{}, errors.New("verified template cache does not match template lock")
	}
	contract, err := template.LoadMaterialized(cachePath)
	if err != nil {
		return MCPTemplateContract{}, fmt.Errorf("load verified template contract: %w", err)
	}
	result := MCPTemplateContract{Source: document.Template.Source,
		RequestedRef: document.Template.RequestedRef, Resolved: document.Template.Resolved,
		Version: document.Template.Version, Digest: document.Template.Digest,
		Name: contract.Name, Tofu: MCPTemplateEngine{Directory: contract.Tofu.Directory,
			Version: contract.Tofu.Version}, Inventory: contract.Inventory}
	if contract.Ansible != nil {
		result.Ansible = &MCPTemplateAnsibleEngine{MCPTemplateEngine: MCPTemplateEngine{
			Directory: contract.Ansible.Directory, Version: contract.Ansible.Version},
			Playbook: contract.Ansible.Playbook}
	}
	return result, nil
}

// DoctorDeployment runs the existing deployment checks against the fixed
// startup snapshot. Reconciliation is neither planned nor applied.
func (session MCPServeSession) DoctorDeployment() (output.Doctor, error) {
	result, _, err := diagnoseDeployment(session.project, session.cacheRoot,
		reconcile.Planner{}, false, false)
	return result, err
}

// DoctorRun validates the latest retained evidence layout beneath the fixed
// startup project without reading infrastructure state.
func (session MCPServeSession) DoctorRun() output.Doctor {
	return diagnoseLatestRun(session.project.Target.Root)
}

// DoctorTemplate diagnoses only the verified template bound to the startup
// project. Missing lock/cache evidence produces the same typed skip/failure
// findings as doctor all and never triggers source acquisition.
func (session MCPServeSession) DoctorTemplate() (output.Doctor, error) {
	deployment, _, err := diagnoseDeployment(session.project, session.cacheRoot,
		reconcile.Planner{}, false, false)
	if err != nil {
		return output.Doctor{}, err
	}
	findings := resolvedTemplateFindings(deployment.Findings)
	return output.Doctor{Scope: "template", Summary: summarizeDoctorFindings(findings),
		Findings: findings}, nil
}

// DoctorEnvironment returns the diagnostic snapshot captured before the MCP
// transport opened.
func (session MCPServeSession) DoctorEnvironment() output.Doctor {
	return session.environmentDoctor
}

// CapabilityEnabled reports whether an optional startup capability was
// explicitly allowlisted.
func (session MCPServeSession) CapabilityEnabled(capability string) bool {
	_, enabled := session.capabilities[capability]
	return enabled
}

// PlanReconciliation computes registered local repairs without applying them.
func (session MCPServeSession) PlanReconciliation() (MCPReconciliationPlan, error) {
	doctorResult, actions, err := diagnoseDeployment(session.project,
		session.cacheRoot, reconcile.Planner{}, false, true)
	if err != nil {
		return MCPReconciliationPlan{}, err
	}
	result := MCPReconciliationPlan{Deployment: session.Project, Doctor: doctorResult,
		Actions: make([]MCPReconciliationAction, len(actions))}
	for index, action := range actions {
		relative, relativeErr := filepath.Rel(session.Project.Root, action.Path)
		if relativeErr != nil || relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			return MCPReconciliationPlan{}, errors.New("reconciliation action escaped project root")
		}
		result.Actions[index] = MCPReconciliationAction{Check: action.CheckID,
			Path: filepath.ToSlash(relative), Kind: string(action.Kind),
			Mode:               fmt.Sprintf("%04o", action.Mode.Perm()),
			RollbackLimitation: action.RollbackLimitation}
	}
	return result, nil
}

// CreatePlan creates a reviewed saved plan through the normal planning core
// while retaining the project and configuration fixed at server startup.
func (session MCPServeSession) CreatePlan(ctx context.Context, intent string) (output.Plan, error) {
	destroy := false
	switch intent {
	case "apply":
	case "destroy":
		destroy = true
	default:
		return output.Plan{}, errors.New("plan intent must be apply or destroy")
	}
	return planForDeployment(ctx, session.project, session.settings, destroy, session.planOptions)
}

// PlanTemplateLock resolves and validates the candidate lock binding without
// publishing ainfra.lock. Cache materialization is the only permitted write.
func (session MCPServeSession) PlanTemplateLock(ctx context.Context,
	operation string,
) (output.Template, error) {
	allowUpdate := false
	switch operation {
	case "lock":
	case "update":
		allowUpdate = true
	default:
		return output.Template{}, errors.New("template operation must be lock or update")
	}
	options := session.templateOptions
	options.CacheDirectory = session.cacheRoot
	return resolveTemplateLockMutation(ctx, session.project, options, allowUpdate, false)
}

// ApplyAuthorized independently authorizes and then executes one exact saved
// apply plan through the normal application core.
func (session MCPServeSession) ApplyAuthorized(ctx context.Context,
	request MCPApplyRequest,
) (MCPApplyResult, error) {
	return session.executeAuthorizedPlan(ctx, request, "apply", "apply",
		func(beforeExecute func() error) (output.Execution, error) {
			return applyForDeploymentWithHook(ctx, session.project, session.settings,
				request.PlanID, session.planOptions, beforeExecute)
		})
}

// DestroyAuthorized independently authorizes and executes one exact saved
// destroy plan. It is exposed only by the separate destruction capability.
func (session MCPServeSession) DestroyAuthorized(ctx context.Context,
	request MCPExecutionRequest,
) (MCPExecutionResult, error) {
	return session.executeAuthorizedPlan(ctx, request, "destroy", "destroy",
		func(beforeExecute func() error) (output.Execution, error) {
			return destroyForDeploymentWithHook(ctx, session.project, session.settings,
				request.PlanID, session.planOptions, beforeExecute)
		})
}

func (session MCPServeSession) executeAuthorizedPlan(ctx context.Context,
	request MCPExecutionRequest, operation, intent string,
	execute func(func() error) (output.Execution, error),
) (MCPExecutionResult, error) {
	binding, err := runstate.LoadAuthorizationBinding(session.runsRoot,
		request.PlanID, session.Project.Name, intent)
	if err != nil {
		return MCPExecutionResult{}, err
	}
	authorization, err := session.AuthorizeMutation(ctx, MCPMutationAuthorizationRequest{
		Approval: request.Approval, Operation: operation, PlanID: binding.PlanID,
		PlanDigest: binding.PlanDigest, Intent: binding.Intent, Caller: request.Caller,
	})
	if err != nil {
		return MCPExecutionResult{}, err
	}
	now := session.now
	if now == nil {
		now = time.Now
	}
	recorded := false
	execution, err := execute(func() error {
		expiresAt, parseErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
		current := now().UTC()
		if parseErr != nil || !expiresAt.After(current) {
			return errors.New("MCP authorization expired before execution")
		}
		record := runstate.NewAuthorizationRecord(authorization.AuthorizationID, operation,
			binding.PlanID, binding.PlanDigest, authorization.Caller, authorization.Issuer,
			authorization.ExpiresAt, current)
		if recordErr := runstate.RecordAuthorization(session.runsRoot, record); recordErr != nil {
			return recordErr
		}
		recorded = true
		return nil
	})
	if recorded {
		execution.Evidence = append(execution.Evidence, output.Evidence{
			Kind: "authorization", Path: "authorization.json", Sensitive: false})
	}
	result := MCPExecutionResult{Execution: execution, Authorization: authorization}
	return result, err
}

// AuthorizeMutation verifies an exact mutation binding against the provider
// fixed at server startup. The project root is always replaced by the
// canonical startup root and cannot come from a protocol request.
func (session MCPServeSession) AuthorizeMutation(ctx context.Context,
	request MCPMutationAuthorizationRequest,
) (MCPAuthorization, error) {
	request.ProjectRoot = session.Project.Root
	now := session.now
	if now == nil {
		now = time.Now
	}
	return VerifyMCPAuthorization(ctx, session.authorization, request, now())
}

// PrepareMCPServe resolves one project and validates its normal configuration
// before a protocol transport starts accepting requests.
func PrepareMCPServe(ctx context.Context, request MCPServeRequest,
	options MCPServeOptions,
) (MCPServeSession, error) {
	capabilities, err := validateMCPCapabilities(request.Capabilities)
	if err != nil {
		return MCPServeSession{}, err
	}
	planOptions := options.Plan
	environmentPath := planOptions.Environment["AINFRA_PROJECT"]
	if request.ProjectPath != "" && environmentPath != "" &&
		request.ProjectPath != environmentPath {
		return MCPServeSession{}, errors.New("--project conflicts with AINFRA_PROJECT")
	}
	deployment, err := project.Load(project.ResolveOptions{
		WorkingDirectory: planOptions.WorkingDirectory,
		ProjectPath:      request.ProjectPath,
		EnvironmentPath:  environmentPath,
	})
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("load MCP project: %w", err)
	}
	settings, err := resolvePlanConfiguration(PlanRequest{}, deployment.Target.Root, planOptions)
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("resolve MCP configuration: %w", err)
	}
	doctorOptions := options.Doctor
	doctorOptions.Environment = cloneEnvironment(doctorOptions.Environment)
	delete(doctorOptions.Environment, "AINFRA_PROJECT")
	environment, err := DoctorEnvironmentContext(ctx, DoctorEnvironmentRequest{
		ProjectPath: deployment.Target.Root, NonInteractive: true,
	}, doctorOptions)
	if err != nil {
		return MCPServeSession{}, fmt.Errorf("diagnose MCP environment: %w", err)
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return MCPServeSession{Project: output.Deployment{
		Name: deployment.Metadata.Name,
		Root: deployment.Target.Root,
	}, project: deployment, runsRoot: settings.Paths.Runs, settings: settings,
		cacheRoot: settings.Paths.Cache, environmentDoctor: environment.Result,
		capabilities: capabilities, authorization: options.Authorization, now: now,
		planOptions:     clonePlanHostOptions(planOptions),
		templateOptions: cloneTemplateLockOptions(options.Template)}, nil
}

func cloneTemplateLockOptions(options TemplateLockOptions) TemplateLockOptions {
	options.Environment = cloneEnvironment(options.Environment)
	return options
}

func clonePlanHostOptions(options PlanHostOptions) PlanHostOptions {
	options.Environment = cloneEnvironment(options.Environment)
	options.ParentEnvironment = append([]string(nil), options.ParentEnvironment...)
	return options
}

func validateMCPCapabilities(requested []string) (map[string]struct{}, error) {
	capabilities := make(map[string]struct{}, len(requested))
	for _, capability := range requested {
		switch capability {
		case MCPPlanningCapability, MCPDeploymentCapability, MCPDestructionCapability:
		default:
			return nil, fmt.Errorf("unsupported MCP capability %q", capability)
		}
		if _, duplicate := capabilities[capability]; duplicate {
			return nil, fmt.Errorf("duplicate MCP capability %q", capability)
		}
		capabilities[capability] = struct{}{}
	}
	if _, destruction := capabilities[MCPDestructionCapability]; destruction {
		if _, deployment := capabilities[MCPDeploymentCapability]; !deployment {
			return nil, errors.New("MCP destruction capability requires deployment capability")
		}
	}
	return capabilities, nil
}

func cloneEnvironment(source map[string]string) map[string]string {
	cloned := make(map[string]string, len(source))
	for name, value := range source {
		cloned[name] = value
	}
	return cloned
}
