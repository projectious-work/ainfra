// Package mcpserver adapts typed ainfra application results to MCP.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	operational "github.com/projectious-work/ainfra/internal/logging"
	"github.com/projectious-work/ainfra/internal/output"
	contracts "github.com/projectious-work/ainfra/spec"
)

const protocolVersion = "2026-07-28"

// Options contains composition-root dependencies for one stdio server.
type Options struct {
	Build       app.Build
	Prepare     func(context.Context, app.MCPServeRequest) (app.MCPServeSession, error)
	Stdin       io.ReadCloser
	Stdout      io.WriteCloser
	Stderr      io.Writer
	Operational *operational.Logger
	Now         func() time.Time
}

// VersionInput is the closed input contract for ainfra.version.
type VersionInput struct{}

// VersionResult is the versioned result contract for ainfra.version.
type VersionResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      output.Version          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// ProjectInput is the closed input contract for ainfra.project.inspect.
type ProjectInput struct{}

// ProjectResult is the versioned result contract for project inspection.
type ProjectResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      output.Deployment       `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// DeploymentInspectInput is closed because the deployment is startup-fixed.
type DeploymentInspectInput struct{}

// DeploymentInspectResult is the versioned sanitized deployment contract.
type DeploymentInspectResult struct {
	APIVersion  string                    `json:"apiVersion"`
	Tool        string                    `json:"tool"`
	OK          bool                      `json:"ok"`
	Result      app.MCPDeploymentContract `json:"result"`
	Diagnostics []diagnostic.Diagnostic   `json:"diagnostics"`
}

// TemplateInspectInput is closed because only the startup-bound lock and cache
// may be inspected.
type TemplateInspectInput struct{}

// TemplateInspectResult is the versioned verified template contract.
type TemplateInspectResult struct {
	APIVersion  string                   `json:"apiVersion"`
	Tool        string                   `json:"tool"`
	OK          bool                     `json:"ok"`
	Result      *app.MCPTemplateContract `json:"result"`
	Diagnostics []diagnostic.Diagnostic  `json:"diagnostics"`
}

// StatusInput is the closed input contract for ainfra.status. The project is
// deliberately absent because it is fixed at server startup.
type StatusInput struct{}

// StatusResult is the versioned result contract for retained status.
type StatusResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Status          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// DoctorDeploymentInput is closed because the project and reconciliation
// policy are fixed at server startup.
type DoctorDeploymentInput struct{}

// DoctorDeploymentResult is the versioned deployment-diagnostic result.
type DoctorDeploymentResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Doctor          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// DoctorRunInput is closed because the project is fixed at server startup.
type DoctorRunInput struct{}

// DoctorRunResult is the versioned retained-run diagnostic result.
type DoctorRunResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Doctor          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// DoctorTemplateInput is closed because only the template bound to the
// startup project may be diagnosed.
type DoctorTemplateInput struct{}

// DoctorTemplateResult is the versioned resolved-template diagnostic result.
type DoctorTemplateResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Doctor          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// DoctorEnvironmentInput is closed because configuration and executable
// selection are captured before the server opens its transport.
type DoctorEnvironmentInput struct{}

// DoctorEnvironmentResult is the versioned environment diagnostic snapshot.
type DoctorEnvironmentResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Doctor          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// RetainedArtifactInput selects one retained run beneath the fixed runs root.
// Filesystem paths and project selectors are deliberately absent.
type RetainedArtifactInput struct {
	RunID string `json:"runId" jsonschema:"retained run ID"`
}

// OutputResult is the versioned retained standardized-output result.
type OutputResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *app.MCPRetainedOutput  `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// InventoryResult is the versioned retained inventory result.
type InventoryResult struct {
	APIVersion  string                    `json:"apiVersion"`
	Tool        string                    `json:"tool"`
	OK          bool                      `json:"ok"`
	Result      *app.MCPRetainedInventory `json:"result"`
	Diagnostics []diagnostic.Diagnostic   `json:"diagnostics"`
}

// ReconciliationPlanInput is closed because the project and planner registry
// are fixed at server startup.
type ReconciliationPlanInput struct{}

// ReconciliationPlanResult is the versioned non-applying planning result.
type ReconciliationPlanResult struct {
	APIVersion  string                     `json:"apiVersion"`
	Tool        string                     `json:"tool"`
	OK          bool                       `json:"ok"`
	Result      *app.MCPReconciliationPlan `json:"result"`
	Diagnostics []diagnostic.Diagnostic    `json:"diagnostics"`
}

// CreatePlanInput selects only the reviewed plan intent. Project,
// configuration, executable, cache, and run roots remain startup-fixed.
type CreatePlanInput struct {
	Intent string `json:"intent" jsonschema:"reviewed plan intent: apply or destroy"`
}

// CreatePlanResult is the versioned saved-plan result.
type CreatePlanResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Plan            `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// TemplatePlanInput selects a non-publishing lock or update preview.
type TemplatePlanInput struct {
	Operation string `json:"operation" jsonschema:"template operation: lock or update"`
}

// TemplatePlanResult is the versioned candidate template binding.
type TemplatePlanResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Template        `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// ApplyInput binds an independently approved caller to one exact saved plan.
// Approval is opaque and is never returned, retained, or logged.
type ApplyInput struct {
	PlanID   string `json:"planId" jsonschema:"exact reviewed plan ID"`
	Caller   string `json:"caller" jsonschema:"authenticated caller ID"`
	Approval string `json:"approval" jsonschema:"opaque independent approval artifact"`
}

// ApplyResult is the versioned authorized execution result.
type ApplyResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *app.MCPApplyResult     `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// ConfigureResult shares the authorized execution result contract.
type ConfigureResult = ApplyResult

// DeployResult shares the authorized execution result contract.
type DeployResult = ApplyResult

// ReconciliationExecuteResult is the versioned authorized local-repair result.
type ReconciliationExecuteResult struct {
	APIVersion  string                       `json:"apiVersion"`
	Tool        string                       `json:"tool"`
	OK          bool                         `json:"ok"`
	Result      *app.MCPReconciliationResult `json:"result"`
	Diagnostics []diagnostic.Diagnostic      `json:"diagnostics"`
}

// DestroyResult is the versioned independently authorized destruction result.
type DestroyResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *app.MCPExecutionResult `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// Serve runs one MCP session until stdin closes or the context is cancelled.
func Serve(ctx context.Context, request app.MCPServeRequest, options Options) error {
	if options.Prepare == nil {
		return errors.New("MCP project preparation is unavailable")
	}
	session, err := options.Prepare(ctx, request)
	if err != nil {
		return err
	}
	server := New(session, options)
	reader := newFrameLimitReadCloser(options.Stdin, maxStdioFrameBytes)
	return server.Run(ctx, &mcp.IOTransport{Reader: reader, Writer: options.Stdout})
}

// New constructs the default-deny server registry.
func New(session app.MCPServeSession, options Options) *mcp.Server {
	logger := slog.New(slog.NewTextHandler(options.Stderr, &slog.HandlerOptions{}))
	server := mcp.NewServer(&mcp.Implementation{Name: "ainfra", Version: options.Build.Version},
		&mcp.ServerOptions{Instructions: "Read-only ainfra tools are exposed by default.", Logger: logger})
	limiter := newRequestLimiter(maxConcurrentRequests)
	limiter.audit = newRequestAudit(options.Operational, options.Now, session.Project.Name)
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.version",
		Description: "Return the ainfra build and supported contract versions.",
		Annotations: &mcp.ToolAnnotations{Title: "ainfra version", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ VersionInput) (*mcp.CallToolResult,
			VersionResult, error) {
			return nil, VersionResult{APIVersion: output.APIVersion, Tool: "ainfra.version",
				OK: true, Result: app.Version(options.Build), Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.project.inspect",
		Description: "Return the immutable project identity selected at server startup.",
		Annotations: &mcp.ToolAnnotations{Title: "inspect ainfra project", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ ProjectInput) (*mcp.CallToolResult,
			ProjectResult, error) {
			return nil, ProjectResult{APIVersion: output.APIVersion,
				Tool: "ainfra.project.inspect", OK: true, Result: session.Project,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.deployment.inspect",
		Description: "Return the validated semantic deployment contract fixed at startup.",
		Annotations: &mcp.ToolAnnotations{Title: "inspect ainfra deployment contract",
			ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ DeploymentInspectInput) (*mcp.CallToolResult,
			DeploymentInspectResult, error) {
			return nil, DeploymentInspectResult{APIVersion: output.APIVersion,
				Tool: "ainfra.deployment.inspect", OK: true, Result: session.InspectDeployment(),
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.template.inspect",
		Description: "Return the verified template lock and semantic contract bound at startup.",
		Annotations: &mcp.ToolAnnotations{Title: "inspect bound ainfra template contract",
			ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ TemplateInspectInput) (*mcp.CallToolResult,
			TemplateInspectResult, error) {
			contract, err := session.InspectTemplate()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, TemplateInspectResult{
					APIVersion: output.APIVersion, Tool: "ainfra.template.inspect", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E2400",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "template",
						NextAction: "Inspect the deployment binding, lock, and verified template cache."}},
				}, nil
			}
			return nil, TemplateInspectResult{APIVersion: output.APIVersion,
				Tool: "ainfra.template.inspect", OK: true, Result: &contract,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.status",
		Description: "Return sanitized retained lifecycle status for the startup project.",
		Annotations: &mcp.ToolAnnotations{Title: "ainfra retained status", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ StatusInput) (*mcp.CallToolResult,
			StatusResult, error) {
			status, err := session.Status()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, StatusResult{
					APIVersion: output.APIVersion, Tool: "ainfra.status", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4601",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "status",
						NextAction: "Inspect retained run evidence and deployment configuration."}},
				}, nil
			}
			return nil, StatusResult{APIVersion: output.APIVersion,
				Tool: "ainfra.status", OK: true, Result: &status,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.doctor.deployment",
		Description: "Diagnose the startup project without reconciliation or mutation.",
		Annotations: &mcp.ToolAnnotations{Title: "diagnose ainfra project", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ DoctorDeploymentInput) (*mcp.CallToolResult,
			DoctorDeploymentResult, error) {
			result, err := session.DoctorDeployment()
			if err != nil {
				code := "AINFRA-E2300"
				var doctorError *app.DoctorError
				if errors.As(err, &doctorError) && doctorError.Kind == "security" {
					code = "AINFRA-E2304"
				}
				return &mcp.CallToolResult{IsError: true}, DoctorDeploymentResult{
					APIVersion: output.APIVersion, Tool: "ainfra.doctor.deployment", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: code,
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "doctor"}},
				}, nil
			}
			failed := make([]diagnostic.Diagnostic, 0, result.Summary.Fail)
			for _, finding := range result.Findings {
				if finding.Status == "fail" {
					failed = append(failed, finding)
				}
			}
			return &mcp.CallToolResult{IsError: len(failed) > 0}, DoctorDeploymentResult{
				APIVersion: output.APIVersion, Tool: "ainfra.doctor.deployment",
				OK: len(failed) == 0, Result: &result, Diagnostics: failed,
			}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.doctor.run",
		Description: "Validate latest retained run evidence for the startup project.",
		Annotations: &mcp.ToolAnnotations{Title: "diagnose latest ainfra run", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ DoctorRunInput) (*mcp.CallToolResult,
			DoctorRunResult, error) {
			result := session.DoctorRun()
			failed := make([]diagnostic.Diagnostic, 0, result.Summary.Fail)
			for _, finding := range result.Findings {
				if finding.Status == "fail" {
					failed = append(failed, finding)
				}
			}
			return &mcp.CallToolResult{IsError: len(failed) > 0}, DoctorRunResult{
				APIVersion: output.APIVersion, Tool: "ainfra.doctor.run",
				OK: len(failed) == 0, Result: &result, Diagnostics: failed,
			}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.doctor.template",
		Description: "Diagnose the verified template bound to the startup project.",
		Annotations: &mcp.ToolAnnotations{Title: "diagnose bound ainfra template", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ DoctorTemplateInput) (*mcp.CallToolResult,
			DoctorTemplateResult, error) {
			result, err := session.DoctorTemplate()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, DoctorTemplateResult{
					APIVersion: output.APIVersion, Tool: "ainfra.doctor.template", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E2400",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "doctor"}},
				}, nil
			}
			failed := make([]diagnostic.Diagnostic, 0, result.Summary.Fail)
			for _, finding := range result.Findings {
				if finding.Status == "fail" {
					failed = append(failed, finding)
				}
			}
			return &mcp.CallToolResult{IsError: len(failed) > 0}, DoctorTemplateResult{
				APIVersion: output.APIVersion, Tool: "ainfra.doctor.template",
				OK: len(failed) == 0, Result: &result, Diagnostics: failed,
			}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.doctor.environment",
		Description: "Return the environment diagnostic snapshot captured at startup.",
		Annotations: &mcp.ToolAnnotations{Title: "diagnose ainfra environment", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ DoctorEnvironmentInput) (*mcp.CallToolResult,
			DoctorEnvironmentResult, error) {
			result := session.DoctorEnvironment()
			failed := make([]diagnostic.Diagnostic, 0, result.Summary.Fail)
			for _, finding := range result.Findings {
				if finding.Status == "fail" {
					failed = append(failed, finding)
				}
			}
			return &mcp.CallToolResult{IsError: len(failed) > 0}, DoctorEnvironmentResult{
				APIVersion: output.APIVersion, Tool: "ainfra.doctor.environment",
				OK: len(failed) == 0, Result: &result, Diagnostics: failed,
			}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.output.read",
		Description: "Read sanitized standardized output retained by a bound run.",
		Annotations: &mcp.ToolAnnotations{Title: "read retained ainfra output", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, input RetainedArtifactInput) (*mcp.CallToolResult,
			OutputResult, error) {
			result, err := session.ReadOutput(input.RunID)
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, OutputResult{
					APIVersion: output.APIVersion, Tool: "ainfra.output.read", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4101",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "output",
						NextAction: "Inspect the applied run binding and retained artifacts."}},
				}, nil
			}
			return nil, OutputResult{APIVersion: output.APIVersion, Tool: "ainfra.output.read",
				OK: true, Result: &result, Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.inventory.read",
		Description: "Read verified deterministic inventory retained by a bound run.",
		Annotations: &mcp.ToolAnnotations{Title: "read retained ainfra inventory", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, input RetainedArtifactInput) (*mcp.CallToolResult,
			InventoryResult, error) {
			result, err := session.ReadInventory(input.RunID)
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, InventoryResult{
					APIVersion: output.APIVersion, Tool: "ainfra.inventory.read", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4101",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "inventory",
						NextAction: "Inspect the applied run binding and retained artifacts."}},
				}, nil
			}
			return nil, InventoryResult{APIVersion: output.APIVersion,
				Tool: "ainfra.inventory.read", OK: true, Result: &result,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	if session.CapabilityEnabled(app.MCPPlanningCapability) {
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.reconciliation.plan",
			Description: "Preview registered local repairs without applying them.",
			Annotations: &mcp.ToolAnnotations{Title: "plan ainfra reconciliation",
				ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
			func(_ context.Context, _ *mcp.CallToolRequest, _ ReconciliationPlanInput) (*mcp.CallToolResult,
				ReconciliationPlanResult, error) {
				result, err := session.PlanReconciliation()
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, ReconciliationPlanResult{
						APIVersion: output.APIVersion, Tool: "ainfra.reconciliation.plan", OK: false,
						Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E2300",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "doctor"}},
					}, nil
				}
				return nil, ReconciliationPlanResult{APIVersion: output.APIVersion,
					Tool: "ainfra.reconciliation.plan", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.template.plan",
			Description: "Preview a validated template lock or update without publishing the lock.",
			Annotations: &mcp.ToolAnnotations{Title: "plan ainfra template binding",
				ReadOnlyHint: false, DestructiveHint: boolPointer(false),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input TemplatePlanInput) (*mcp.CallToolResult,
				TemplatePlanResult, error) {
				result, err := session.PlanTemplateLock(ctx, input.Operation)
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, TemplatePlanResult{
						APIVersion: output.APIVersion, Tool: "ainfra.template.plan", OK: false,
						Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E3001",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "template",
							NextAction: "Correct the template source, lock state, or requested operation."}},
					}, nil
				}
				return nil, TemplatePlanResult{APIVersion: output.APIVersion,
					Tool: "ainfra.template.plan", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.plan.create",
			Description: "Create a reviewed saved apply or destroy plan for the startup project.",
			Annotations: &mcp.ToolAnnotations{Title: "create reviewed ainfra plan",
				ReadOnlyHint: false, DestructiveHint: boolPointer(false),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input CreatePlanInput) (*mcp.CallToolResult,
				CreatePlanResult, error) {
				result, err := session.CreatePlan(ctx, input.Intent)
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, CreatePlanResult{
						APIVersion: output.APIVersion, Tool: "ainfra.plan.create", OK: false,
						Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4001",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "plan",
							NextAction: "Correct the deployment, lock, or OpenTofu configuration."}},
					}, nil
				}
				return nil, CreatePlanResult{APIVersion: output.APIVersion,
					Tool: "ainfra.plan.create", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
	}
	if session.CapabilityEnabled(app.MCPDeploymentCapability) {
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.reconciliation.execute",
			Description: "Apply one exact independently authorized reconciliation plan.",
			Annotations: &mcp.ToolAnnotations{Title: "execute authorized ainfra reconciliation",
				ReadOnlyHint: false, DestructiveHint: boolPointer(true),
				IdempotentHint: false, OpenWorldHint: boolPointer(false)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input ApplyInput) (*mcp.CallToolResult,
				ReconciliationExecuteResult, error) {
				result, err := session.ReconcileAuthorized(ctx, app.MCPExecutionRequest{
					PlanID: input.PlanID, Caller: input.Caller, Approval: input.Approval})
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, ReconciliationExecuteResult{
						APIVersion: output.APIVersion, Tool: "ainfra.reconciliation.execute", OK: false,
						Result: &result, Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E2300",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "doctor",
							NextAction: "Review a current reconciliation plan and obtain fresh independent approval."}},
					}, nil
				}
				return nil, ReconciliationExecuteResult{APIVersion: output.APIVersion,
					Tool: "ainfra.reconciliation.execute", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.configure.execute",
			Description: "Configure one exact independently authorized applied run.",
			Annotations: &mcp.ToolAnnotations{Title: "execute authorized ainfra configuration",
				ReadOnlyHint: false, DestructiveHint: boolPointer(true),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input ApplyInput) (*mcp.CallToolResult,
				ConfigureResult, error) {
				result, err := session.ConfigureAuthorized(ctx, app.MCPExecutionRequest{
					PlanID: input.PlanID, Caller: input.Caller, Approval: input.Approval})
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, ConfigureResult{
						APIVersion: output.APIVersion, Tool: "ainfra.configure.execute", OK: false,
						Result: &result, Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4301",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "configure",
							NextAction: "Reverify retained artifacts and obtain fresh configuration approval."}},
					}, nil
				}
				return nil, ConfigureResult{APIVersion: output.APIVersion,
					Tool: "ainfra.configure.execute", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.deploy.execute",
			Description: "Execute one exact independently authorized composed deployment.",
			Annotations: &mcp.ToolAnnotations{Title: "execute authorized ainfra deployment",
				ReadOnlyHint: false, DestructiveHint: boolPointer(true),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input ApplyInput) (*mcp.CallToolResult,
				DeployResult, error) {
				result, err := session.DeployAuthorized(ctx, app.MCPExecutionRequest{
					PlanID: input.PlanID, Caller: input.Caller, Approval: input.Approval})
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, DeployResult{
						APIVersion: output.APIVersion, Tool: "ainfra.deploy.execute", OK: false,
						Result: &result, Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4501",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "deploy",
							NextAction: "Inspect retained stage evidence and obtain fresh deployment approval."}},
					}, nil
				}
				return nil, DeployResult{APIVersion: output.APIVersion,
					Tool: "ainfra.deploy.execute", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.apply.execute",
			Description: "Execute one exact independently authorized reviewed apply plan.",
			Annotations: &mcp.ToolAnnotations{Title: "execute authorized ainfra apply",
				ReadOnlyHint: false, DestructiveHint: boolPointer(true),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input ApplyInput) (*mcp.CallToolResult,
				ApplyResult, error) {
				result, err := session.ApplyAuthorized(ctx, app.MCPApplyRequest{
					PlanID: input.PlanID, Caller: input.Caller, Approval: input.Approval})
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, ApplyResult{
						APIVersion: output.APIVersion, Tool: "ainfra.apply.execute", OK: false,
						Result: &result, Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4002",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "apply",
							NextAction: "Reverify the saved plan and obtain fresh independent approval."}},
					}, nil
				}
				return nil, ApplyResult{APIVersion: output.APIVersion,
					Tool: "ainfra.apply.execute", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
	}
	if session.CapabilityEnabled(app.MCPDestructionCapability) {
		addBoundedTool(server, limiter, &mcp.Tool{Name: "ainfra.destroy.execute",
			Description: "Execute one exact independently authorized reviewed destroy plan.",
			Annotations: &mcp.ToolAnnotations{Title: "execute authorized ainfra destruction",
				ReadOnlyHint: false, DestructiveHint: boolPointer(true),
				IdempotentHint: false, OpenWorldHint: boolPointer(true)}},
			func(ctx context.Context, _ *mcp.CallToolRequest, input ApplyInput) (*mcp.CallToolResult,
				DestroyResult, error) {
				result, err := session.DestroyAuthorized(ctx, app.MCPExecutionRequest{
					PlanID: input.PlanID, Caller: input.Caller, Approval: input.Approval})
				if err != nil {
					return &mcp.CallToolResult{IsError: true}, DestroyResult{
						APIVersion: output.APIVersion, Tool: "ainfra.destroy.execute", OK: false,
						Result: &result, Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4402",
							Severity: diagnostic.SeverityError, Message: err.Error(), Component: "destroy",
							NextAction: "Reverify the destroy plan and obtain fresh destroy approval."}},
					}, nil
				}
				return nil, DestroyResult{APIVersion: output.APIVersion,
					Tool: "ainfra.destroy.execute", OK: true, Result: &result,
					Diagnostics: []diagnostic.Diagnostic{}}, nil
			})
	}
	addContractResources(server, limiter)
	return server
}

func addContractResources(server *mcp.Server, limiter *requestLimiter) {
	for _, name := range contracts.Schemas() {
		name := name
		contents, ok := contracts.Schema(name)
		if !ok {
			panic("registered schema is not embedded: " + name)
		}
		uri := "ainfra://schemas/v1/" + name
		addBoundedResource(server, limiter, &mcp.Resource{Name: name, Title: "ainfra v1 schema: " + name,
			Description: "Published, immutable ainfra v1 JSON Schema.",
			MIMEType:    "application/schema+json", URI: uri, Size: int64(len(contents))},
			func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
					URI: uri, MIMEType: "application/schema+json", Text: string(contents),
				}}}, nil
			})
	}
	catalog := contractCatalog()
	const catalogURI = "ainfra://contracts/v1"
	addBoundedResource(server, limiter, &mcp.Resource{Name: "ainfra-contracts-v1",
		Title:       "ainfra v1 contracts and documentation",
		Description: "Supported contract versions, schema resources, and documentation references.",
		MIMEType:    "application/json", URI: catalogURI, Size: int64(len(catalog))},
		func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
				URI: catalogURI, MIMEType: "application/json", Text: catalog,
			}}}, nil
		})
}

func contractCatalog() string {
	type reference struct {
		Name string `json:"name"`
		URI  string `json:"uri"`
	}
	schemaNames := contracts.Schemas()
	schemas := make([]reference, len(schemaNames))
	for index, name := range schemaNames {
		schemas[index] = reference{Name: name, URI: "ainfra://schemas/v1/" + name}
	}
	value := struct {
		APIVersion                string                           `json:"apiVersion"`
		SupportedContractVersions output.SupportedContractVersions `json:"supportedContractVersions"`
		Schemas                   []reference                      `json:"schemas"`
		Documentation             []reference                      `json:"documentation"`
	}{APIVersion: "ainfra.contracts/v1",
		SupportedContractVersions: app.Version(app.Build{}).SupportedContractVersions,
		Schemas:                   schemas,
		Documentation: []reference{
			{Name: "documentation", URI: "https://projectious-work.github.io/ainfra/docs/"},
			{Name: "roadmap", URI: "https://projectious-work.github.io/ainfra/docs/roadmap/"},
			{Name: "v1 specification", URI: "https://github.com/projectious-work/ainfra/tree/v1.x-dev/spec/doc/v1"},
		},
	}
	contents, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

func boolPointer(value bool) *bool { return &value }

// ProtocolVersion identifies the MCP specification selected for Phase 7.
func ProtocolVersion() string { return protocolVersion }
