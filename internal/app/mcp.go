package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/reconcile"
	runstate "github.com/projectious-work/ainfra/internal/run"
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
		planOptions: clonePlanHostOptions(planOptions)}, nil
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
