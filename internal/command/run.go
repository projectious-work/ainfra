package command

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/initialize"
	operational "github.com/projectious-work/ainfra/internal/logging"
	"github.com/projectious-work/ainfra/internal/output"
	"github.com/projectious-work/ainfra/internal/reconcile"
	"github.com/projectious-work/ainfra/internal/security"
)

// IO supplies explicit process streams and terminal capabilities.
type IO struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	IsTerminal bool
}

// Options supplies immutable composition facts to Run.
type Options struct {
	Build             app.Build
	IO                IO
	DoctorEnvironment func(app.DoctorEnvironmentRequest) (app.DoctorEnvironmentResponse, error)
	DoctorDeployment  func(app.DoctorDeploymentRequest) (app.DoctorEnvironmentResponse, error)
	DoctorTemplate    func(app.DoctorTemplateRequest) (app.DoctorEnvironmentResponse, error)
	DoctorRun         func(app.DoctorRunRequest) (app.DoctorEnvironmentResponse, error)
	DoctorAll         func(app.DoctorAllRequest) (app.DoctorEnvironmentResponse, error)
	TemplateLock      func(app.TemplateLockRequest) (output.Template, error)
	TemplateUpdate    func(app.TemplateLockRequest) (output.Template, error)
	Plan              func(app.PlanRequest) (output.Plan, error)
	Apply             func(app.ApplyRequest) (output.Execution, error)
	Destroy           func(app.DestroyRequest) (output.Execution, error)
	Logs              func(app.EvidenceRequest) (output.Logs, error)
	Status            func(app.EvidenceRequest) (output.Status, error)
	Output            func(app.ArtifactRequest) (output.Artifact, error)
	Inventory         func(app.ArtifactRequest) (output.Artifact, error)
	Configure         func(app.ConfigureRequest) (output.Execution, error)
	Deploy            func(app.DeployRequest) (output.Execution, error)
	MCPServe          func(context.Context, app.MCPServeRequest) error
	Initialize        func(string) (initialize.Result, error)
	Operational       *operational.Logger
	Now               func() time.Time
}

// Run parses one CLI invocation, renders its result, and returns its exit code.
func Run(arguments []string, options Options) (exit ExitCode) {
	if options.Operational != nil {
		now := time.Now
		if options.Now != nil {
			now = options.Now
		}
		commandName := "help"
		if len(arguments) > 0 {
			commandName = arguments[0]
		}
		if err := options.Operational.Write(operational.NewEvent(now(), "info", "command",
			"command started", commandName, "", nil)); err != nil {
			_, _ = fmt.Fprintf(options.IO.Stderr, "AINFRA-E0003: initialize operational logging: %s\n", err)
			return ExitOperationFailed
		}
		defer func() {
			level, message := "info", "command finished"
			if exit != ExitSuccess {
				level, message = "error", "command failed"
			}
			if err := options.Operational.Write(operational.NewEvent(now(), level, "command",
				message, commandName, "", nil)); err != nil {
				_, _ = fmt.Fprintf(options.IO.Stderr, "AINFRA-E0003: finalize operational logging: %s\n", err)
				exit = ExitOperationFailed
			}
		}()
	}
	renderArguments, controlArguments, positional, helpRequested := splitInvocation(arguments)
	renderOptions, err := parseRenderOptions(renderArguments, options.IO.IsTerminal)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}

	if len(arguments) == 0 || helpRequested || (len(positional) > 0 && positional[0] == "help") {
		topic, topicErr := normalizeHelpTopic(positional, helpRequested)
		if topicErr != nil {
			return failInvocation(arguments, topicErr.Error(), options.IO)
		}
		result, found := helpFor(topic)
		if !found {
			return failInvocation(arguments, fmt.Sprintf("unknown help topic %q", topic), options.IO)
		}
		envelope := output.Success(output.CommandHelp, result)
		if err := output.Render(options.IO.Stdout, envelope, renderOptions); err != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	if len(positional) == 2 && positional[0] == "doctor" && positional[1] == "environment" {
		return runDoctorEnvironment(
			arguments, renderArguments, controlArguments, renderOptions, options,
		)
	}
	if len(positional) == 2 && positional[0] == "mcp" && positional[1] == "serve" {
		return runMCPServe(arguments, controlArguments, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "init" {
		if len(controlArguments) != 0 {
			return failInvocation(arguments, "init does not accept configuration options", options.IO)
		}
		return runInit(positional, renderOptions, options)
	}
	if len(positional) >= 2 && len(positional) <= 3 &&
		positional[0] == "doctor" && positional[1] == "deployment" {
		return runDoctorDeployment(
			arguments, renderArguments, controlArguments, positional, renderOptions, options,
		)
	}
	if len(positional) >= 2 && len(positional) <= 3 && positional[0] == "template" &&
		(positional[1] == "lock" || positional[1] == "update") {
		return runTemplateLock(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "plan" {
		return runPlan(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "apply" {
		return runApply(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "destroy" {
		return runDestroy(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "logs" {
		return runLogs(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "status" {
		return runStatus(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 &&
		(positional[0] == "output" || positional[0] == "inventory") {
		return runArtifact(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "configure" {
		return runConfigure(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 1 && len(positional) <= 2 && positional[0] == "deploy" {
		return runDeploy(arguments, controlArguments, positional, renderOptions, options)
	}
	if len(positional) >= 2 && len(positional) <= 3 &&
		positional[0] == "doctor" && positional[1] == "template" {
		return runDoctorTemplate(
			arguments, renderArguments, controlArguments, positional, renderOptions, options,
		)
	}
	if len(positional) >= 2 && len(positional) <= 3 &&
		positional[0] == "doctor" && positional[1] == "run" {
		return runDoctorTarget(
			arguments, renderArguments, controlArguments, positional,
			renderOptions, options, output.CommandDoctorRun,
		)
	}
	if (len(positional) >= 1 && len(positional) <= 2 && positional[0] == "doctor") ||
		(len(positional) >= 2 && len(positional) <= 3 &&
			positional[0] == "doctor" && positional[1] == "all") {
		return runDoctorTarget(
			arguments, renderArguments, controlArguments, positional,
			renderOptions, options, output.CommandDoctorAll,
		)
	}

	if len(positional) != 1 || (positional[0] != "version" && positional[0] != "--version") {
		return failInvocation(arguments, "invalid command invocation", options.IO)
	}

	envelope := output.Success(output.CommandVersion, app.Version(options.Build))
	if err := output.Render(options.IO.Stdout, envelope, renderOptions); err != nil {
		if _, writeErr := fmt.Fprintf(options.IO.Stderr, "AINFRA-E0003: render result: %s\n", err); writeErr != nil {
			return ExitOperationFailed
		}
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runMCPServe(arguments, controlArguments []string, options Options) ExitCode {
	flags := flag.NewFlagSet("mcp serve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	stdio := flags.Bool("stdio", false, "serve MCP over standard input and output")
	projectPath := flags.String("project", "", "allowed project root")
	capabilities := make([]string, 0, 3)
	flags.Func("capability", "explicitly allowlisted capability group", func(value string) error {
		if value == "" {
			return errors.New("capability name is required")
		}
		capabilities = append(capabilities, value)
		return nil
	})
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if !*stdio {
		return failInvocation(arguments, "mcp serve requires --stdio", options.IO)
	}
	if options.MCPServe == nil {
		return failInvocation(arguments, "MCP server mode is unavailable", options.IO)
	}
	if err := options.MCPServe(context.Background(), app.MCPServeRequest{
		ProjectPath: *projectPath, Capabilities: capabilities,
	}); err != nil {
		_, _ = fmt.Fprintf(options.IO.Stderr, "AINFRA-E5001: MCP server failed: %s\n", err)
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runDeploy(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	planID := flags.String("plan", "", "reviewed plan ID")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *planID == "" {
		return failInvocation(arguments, "deploy requires --plan RUN_ID", options.IO)
	}
	if options.Deploy == nil {
		return failInvocation(arguments, "deploy is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Deploy(app.DeployRequest{ApplyRequest: app.ApplyRequest{Target: target,
		ProjectPath: *projectPath, ConfigPath: *configPath, PlanID: *planID}})
	if err == nil {
		if output.Render(options.IO.Stdout, output.Success(output.CommandDeploy, result), renderOptions) != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4301", Severity: diagnostic.SeverityError,
		Message: err.Error(), Component: "deploy", NextAction: "Inspect completed-stage evidence before resuming."}
	var failure *app.DeployFailure
	if errors.As(err, &failure) {
		result = failure.Result
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout, output.PartialFailure(output.CommandDeploy, result, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return ExitOperationFailed
	}
	if renderOptions.Format == output.FormatJSON {
		if output.Render(options.IO.Stdout, output.Failure(output.CommandDeploy, diagnosticValue), renderOptions) != nil {
			return ExitOperationFailed
		}
	} else {
		_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
	}
	return ExitOperationFailed
}

func runConfigure(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("configure", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	runID := flags.String("run", "", "applied run ID")
	check := flags.Bool("check", false, "verify convergence in check mode")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *runID == "" {
		return failInvocation(arguments, "configure requires --run RUN_ID", options.IO)
	}
	if options.Configure == nil {
		return failInvocation(arguments, "configure is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Configure(app.ConfigureRequest{ArtifactRequest: app.ArtifactRequest{
		Target: target, ProjectPath: *projectPath, ConfigPath: *configPath, RunID: *runID}, Check: *check})
	if err == nil {
		if output.Render(options.IO.Stdout, output.Success(output.CommandConfigure, result), renderOptions) != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4201", Severity: diagnostic.SeverityError,
		Message: err.Error(), Component: "configure", NextAction: "Inspect generated inventory and Ansible Runner evidence."}
	exit := ExitOperationFailed
	var failure *app.ConfigureFailure
	if errors.As(err, &failure) {
		result = failure.Result
		if result.ExecutionOutcome == "interrupted" {
			diagnosticValue.Code, exit = "AINFRA-E0006", ExitInterrupted
		}
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout, output.PartialFailure(output.CommandConfigure, result, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return exit
	}
	if renderOptions.Format == output.FormatJSON {
		if output.Render(options.IO.Stdout, output.Failure(output.CommandConfigure, diagnosticValue), renderOptions) != nil {
			return ExitOperationFailed
		}
	} else {
		_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
	}
	return exit
}

func runArtifact(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet(positional[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	runID := flags.String("run", "", "applied run ID")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *runID == "" {
		return failInvocation(arguments, positional[0]+" requires --run RUN_ID", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	request := app.ArtifactRequest{Target: target, ProjectPath: *projectPath, ConfigPath: *configPath, RunID: *runID}
	command := output.CommandOutput
	operation := options.Output
	if positional[0] == "inventory" {
		command, operation = output.CommandInventory, options.Inventory
	}
	if operation == nil {
		return failInvocation(arguments, positional[0]+" is unavailable", options.IO)
	}
	result, err := operation(request)
	if err != nil {
		diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4101", Severity: diagnostic.SeverityError,
			Message: err.Error(), Component: positional[0],
			NextAction: "Inspect the applied run binding and retained artifacts."}
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout, output.Failure(command, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return ExitOperationFailed
	}
	if err := output.Render(options.IO.Stdout, output.Success(command, result), renderOptions); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runApply(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	planID := flags.String("plan", "", "reviewed plan ID")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *planID == "" {
		return failInvocation(arguments, "apply requires --plan RUN_ID", options.IO)
	}
	if options.Apply == nil {
		return failInvocation(arguments, "apply is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Apply(app.ApplyRequest{Target: target, ProjectPath: *projectPath,
		ConfigPath: *configPath, PlanID: *planID})
	if err == nil {
		if renderErr := output.Render(options.IO.Stdout,
			output.Success(output.CommandApply, result), renderOptions); renderErr != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4002",
		Severity: diagnostic.SeverityError, Message: err.Error(), Component: "apply",
		NextAction: "Inspect the reviewed plan binding and retained run evidence."}
	exit := ExitStaleBinding
	var failure *app.ApplyFailure
	if errors.As(err, &failure) {
		result = failure.Result
		exit = ExitOperationFailed
		if result.ExecutionOutcome == "interrupted" {
			diagnosticValue.Code = "AINFRA-E0006"
			exit = ExitInterrupted
		}
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout,
				output.PartialFailure(output.CommandApply, result, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return exit
	}
	if renderOptions.Format == output.FormatJSON {
		if output.Render(options.IO.Stdout,
			output.Failure(output.CommandApply, diagnosticValue), renderOptions) != nil {
			return ExitOperationFailed
		}
	} else {
		_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
	}
	return exit
}

func runDestroy(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("destroy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	planID := flags.String("plan", "", "reviewed destroy plan ID")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *planID == "" {
		return failInvocation(arguments, "destroy requires --plan RUN_ID", options.IO)
	}
	if options.Destroy == nil {
		return failInvocation(arguments, "destroy is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Destroy(app.DestroyRequest{Target: target, ProjectPath: *projectPath,
		ConfigPath: *configPath, PlanID: *planID})
	if err == nil {
		if output.Render(options.IO.Stdout, output.Success(output.CommandDestroy, result), renderOptions) != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4402",
		Severity: diagnostic.SeverityError, Message: err.Error(), Component: "destroy",
		NextAction: "Inspect the reviewed destroy-plan binding and retained run evidence."}
	exit := ExitStaleBinding
	var failure *app.DestroyFailure
	if errors.As(err, &failure) {
		result = failure.Result
		exit = ExitOperationFailed
		if result.ExecutionOutcome == "interrupted" {
			diagnosticValue.Code, exit = "AINFRA-E0006", ExitInterrupted
		}
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout,
				output.PartialFailure(output.CommandDestroy, result, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return exit
	}
	if renderOptions.Format == output.FormatJSON {
		if output.Render(options.IO.Stdout,
			output.Failure(output.CommandDestroy, diagnosticValue), renderOptions) != nil {
			return ExitOperationFailed
		}
	} else {
		_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
	}
	return exit
}

func runLogs(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("logs", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	runID := flags.String("run", "", "retained run ID")
	source := flags.String("source", "", "evidence source")
	errorsOnly := flags.Bool("errors", false, "show attributed errors")
	raw := flags.Bool("raw", false, "show sensitive raw evidence")
	stream := flags.String("stream", "", "retained stream")
	nonInteractive := flags.Bool("non-interactive", false, "disable prompts")
	yes := flags.Bool("yes", false, "confirm raw access")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *runID == "" {
		return failInvocation(arguments, "logs requires --run RUN_ID", options.IO)
	}
	if options.Logs == nil {
		return failInvocation(arguments, "logs is unavailable", options.IO)
	}
	if *raw {
		if renderOptions.Format == output.FormatJSON {
			return failInvocation(arguments, "--raw and --format json are mutually exclusive", options.IO)
		}
		if *errorsOnly {
			return failInvocation(arguments, "--raw and --errors are mutually exclusive", options.IO)
		}
		if !confirmRawAccess(*nonInteractive, *yes, options.IO) {
			return failInvocation(arguments, "raw evidence access requires confirmation", options.IO)
		}
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Logs(app.EvidenceRequest{Target: target, ProjectPath: *projectPath,
		ConfigPath: *configPath, RunID: *runID, Source: *source, Errors: *errorsOnly,
		Raw: *raw, Stream: *stream})
	if err != nil {
		return failCommand(output.CommandLogs, "AINFRA-E4501", "logs", err, renderOptions, options.IO)
	}
	if *raw {
		_, _ = io.WriteString(options.IO.Stderr,
			"WARNING: writing sensitive raw engine evidence to stdout only.\n")
		for _, record := range result.Records {
			if _, err := io.WriteString(options.IO.Stdout, record); err != nil {
				return ExitOperationFailed
			}
		}
		return ExitSuccess
	}
	if output.Render(options.IO.Stdout, output.Success(output.CommandLogs, result), renderOptions) != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func confirmRawAccess(nonInteractive, yes bool, streams IO) bool {
	if nonInteractive {
		return yes
	}
	if yes || !streams.IsTerminal || streams.Stdin == nil {
		return false
	}
	_, _ = io.WriteString(streams.Stderr,
		"WARNING: raw engine evidence may contain credentials and secrets. Continue? [y/N] ")
	answer, err := bufio.NewReader(streams.Stdin).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func runStatus(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if options.Status == nil {
		return failInvocation(arguments, "status is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Status(app.EvidenceRequest{Target: target,
		ProjectPath: *projectPath, ConfigPath: *configPath})
	if err != nil {
		return failCommand(output.CommandStatus, "AINFRA-E4601", "status", err, renderOptions, options.IO)
	}
	if output.Render(options.IO.Stdout, output.Success(output.CommandStatus, result), renderOptions) != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func failCommand(command output.Command, code, component string, err error, renderOptions output.RenderOptions, streams IO) ExitCode {
	diagnosticValue := diagnostic.Diagnostic{Code: code, Severity: diagnostic.SeverityError,
		Message: err.Error(), Component: component, NextAction: "Inspect retained run evidence and deployment configuration."}
	if renderOptions.Format == output.FormatJSON {
		if output.Render(streams.Stdout, output.Failure(command, diagnosticValue), renderOptions) != nil {
			return ExitOperationFailed
		}
	} else {
		_, _ = fmt.Fprintf(streams.Stderr, "%s: %s\n", code, err)
	}
	return ExitInvalidInput
}

func runPlan(arguments, controlArguments, positional []string, renderOptions output.RenderOptions, options Options) ExitCode {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	destroy := flags.Bool("destroy", false, "create destroy plan")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if options.Plan == nil {
		return failInvocation(arguments, "plan is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Plan(app.PlanRequest{Target: target, ProjectPath: *projectPath,
		ConfigPath: *configPath, Destroy: *destroy})
	if err != nil {
		diagnosticValue := diagnostic.Diagnostic{Code: "AINFRA-E4001", Severity: diagnostic.SeverityError,
			Message: err.Error(), Component: "plan",
			NextAction: "Correct the deployment, lock, or OpenTofu configuration and rerun 'ainfra plan'."}
		if renderOptions.Format == output.FormatJSON {
			if output.Render(options.IO.Stdout, output.Failure(output.CommandPlan, diagnosticValue), renderOptions) != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return ExitOperationFailed
	}
	if err := output.Render(options.IO.Stdout, output.Success(output.CommandPlan, result), renderOptions); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runInit(
	positional []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	if options.Initialize == nil {
		return failInvocation(positional, "init is unavailable", options.IO)
	}
	target := "."
	if len(positional) == 2 {
		target = positional[1]
	}
	result, err := options.Initialize(target)
	if err != nil {
		diagnosticValue := diagnostic.Diagnostic{
			Code: "AINFRA-E1001", Severity: diagnostic.SeverityError,
			Message: err.Error(), Component: "initialization",
			NextAction: "Choose an empty deployment path and rerun 'ainfra init'.",
		}
		if renderOptions.Format == output.FormatJSON {
			if renderErr := output.Render(
				options.IO.Stdout, output.Failure(output.CommandInit, diagnosticValue),
				renderOptions,
			); renderErr != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", diagnosticValue.Code, err)
		}
		return ExitInvalidInput
	}
	envelope := output.Success(output.CommandInit, output.Init{
		Deployment:   output.Deployment{Name: result.Name, Root: result.Root},
		CreatedPaths: result.CreatedPaths,
	})
	if err := output.Render(options.IO.Stdout, envelope, renderOptions); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runTemplateLock(
	arguments, controlArguments, positional []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	flags := flag.NewFlagSet("template lock", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	projectPath := flags.String("project", "", "deployment project")
	configPath := flags.String("config", "", "configuration path")
	nonInteractive := flags.Bool("non-interactive", false, "disable prompts")
	yes := flags.Bool("yes", false, "confirm mutation")
	if err := flags.Parse(controlArguments); err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if *yes && !*nonInteractive {
		return failInvocation(arguments, "--yes requires --non-interactive", options.IO)
	}
	operation := positional[1]
	commandName := output.CommandTemplateLock
	mutation := options.TemplateLock
	if operation == "update" {
		commandName, mutation = output.CommandTemplateUpdate, options.TemplateUpdate
	}
	if mutation == nil {
		return failInvocation(arguments, "template "+operation+" is unavailable", options.IO)
	}
	target := ""
	if len(positional) == 3 {
		target = positional[2]
	}
	result, err := mutation(app.TemplateLockRequest{
		Target: target, ProjectPath: *projectPath, ConfigPath: *configPath,
	})
	if err != nil {
		exit, code := ExitInvalidInput, "AINFRA-E3001"
		var refusal *security.Refusal
		if errors.As(err, &refusal) {
			exit, code = ExitSecurity, "AINFRA-E3004"
		}
		diagnosticValue := diagnostic.Diagnostic{
			Code: code, Severity: diagnostic.SeverityError,
			Message: err.Error(), Component: "template-lock",
			NextAction: "Correct the source or lock state and rerun 'ainfra template " + operation + "'.",
		}
		if renderOptions.Format == output.FormatJSON {
			if renderErr := output.Render(
				options.IO.Stdout,
				output.Failure(commandName, diagnosticValue),
				renderOptions,
			); renderErr != nil {
				return ExitOperationFailed
			}
		} else {
			_, _ = fmt.Fprintf(options.IO.Stderr, "%s: %s\n", code, err)
		}
		return exit
	}
	if err := output.Render(
		options.IO.Stdout, output.Success(commandName, result), renderOptions,
	); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runDoctorTarget(
	arguments, renderArguments, controlArguments, positional []string,
	renderOptions output.RenderOptions,
	options Options,
	commandName output.Command,
) ExitCode {
	common, err := parseDoctorEnvironmentRequest(renderArguments, controlArguments)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	targetIndex := 2
	if commandName == output.CommandDoctorAll && len(positional) > 0 &&
		(len(positional) == 1 || positional[1] != "all") {
		targetIndex = 1
	}
	target := ""
	if len(positional) > targetIndex {
		target = positional[targetIndex]
	}
	request := app.DoctorDeploymentRequest{
		Target: target, ConfigPath: common.ConfigPath, ProjectPath: common.ProjectPath,
		Format: common.Format, OutputStyle: common.OutputStyle, Color: common.Color,
		Reconcile: common.Reconcile,
	}
	var response app.DoctorEnvironmentResponse
	if commandName == output.CommandDoctorRun {
		if options.DoctorRun == nil {
			return failDoctor(arguments, commandName, "run doctor is unavailable", options.IO)
		}
		response, err = options.DoctorRun(request)
	} else {
		if options.DoctorAll == nil {
			return failDoctor(arguments, commandName, "all doctor is unavailable", options.IO)
		}
		response, err = options.DoctorAll(request)
	}
	if err != nil {
		exit, code := ExitInvalidInput, "AINFRA-E2500"
		if commandName == output.CommandDoctorAll {
			code = "AINFRA-E2600"
		}
		var doctorError *app.DoctorError
		if errors.As(err, &doctorError) && doctorError.Kind == "security" {
			exit, code = ExitSecurity, "AINFRA-E2504"
			if commandName == output.CommandDoctorAll {
				code = "AINFRA-E2604"
			}
		}
		return failDoctorWithExit(arguments, commandName, code, err.Error(), options.IO, exit)
	}
	if common.Reconcile && len(response.ReconciliationPlan) > 0 {
		if !confirmReconciliation(response.ReconciliationPlan, common, options.IO) {
			return failDoctorWithExit(
				arguments, commandName, "AINFRA-E2002",
				"reconciliation requires confirmation", options.IO, ExitInvalidInput,
			)
		}
		request.ApplyReconciliation = true
		if commandName == output.CommandDoctorAll {
			response, err = options.DoctorAll(request)
		} else {
			response, err = options.DoctorRun(request)
		}
		if err != nil {
			return failDoctorWithExit(
				arguments, commandName, "AINFRA-E2003", err.Error(),
				options.IO, ExitOperationFailed,
			)
		}
	}
	return renderDoctorResponse(commandName, response, renderOptions, options.IO)
}

func confirmReconciliation(
	plan []reconcile.Action,
	request app.DoctorEnvironmentRequest,
	streams IO,
) bool {
	for _, action := range plan {
		_, _ = fmt.Fprintf(
			streams.Stderr, "reconcile %s: %s (%04o)\n  rollback: %s\n",
			action.CheckID, action.Path, action.Mode.Perm(), action.RollbackLimitation,
		)
	}
	if request.NonInteractive {
		return request.Yes
	}
	if !streams.IsTerminal || streams.Stdin == nil {
		return false
	}
	_, _ = io.WriteString(streams.Stderr, "Apply reconciliation plan? [y/N] ")
	answer, err := bufio.NewReader(streams.Stdin).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func renderDoctorResponse(
	commandName output.Command,
	response app.DoctorEnvironmentResponse,
	renderOptions output.RenderOptions,
	streams IO,
) ExitCode {
	renderOptions.Format = output.Format(response.Format)
	renderOptions.Style = output.Style(response.OutputStyle)
	renderOptions.Color = output.ColorMode(response.Color)
	envelope := output.Success(commandName, response.Result)
	exit := ExitSuccess
	if response.Result.Summary.Fail > 0 {
		failed := make([]diagnostic.Diagnostic, 0, response.Result.Summary.Fail)
		for _, finding := range response.Result.Findings {
			if finding.Status == "fail" {
				failed = append(failed, finding)
			}
		}
		envelope = output.PartialFailure(commandName, response.Result, failed...)
		exit = ExitDependency
	}
	if err := output.Render(streams.Stdout, envelope, renderOptions); err != nil {
		return ExitOperationFailed
	}
	return exit
}

func runDoctorTemplate(
	arguments, renderArguments, controlArguments, positional []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	if options.DoctorTemplate == nil {
		return failDoctor(
			arguments, output.CommandDoctorTemplate,
			"template doctor is unavailable", options.IO,
		)
	}
	common, err := parseDoctorEnvironmentRequest(renderArguments, controlArguments)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	if common.ProjectPath != "" {
		return failInvocation(arguments, "doctor template does not accept --project", options.IO)
	}
	target := ""
	if len(positional) == 3 {
		target = positional[2]
	}
	response, err := options.DoctorTemplate(app.DoctorTemplateRequest{
		Target: target, ConfigPath: common.ConfigPath, Format: common.Format,
		OutputStyle: common.OutputStyle, Color: common.Color,
	})
	if err != nil {
		exit := ExitInvalidInput
		code := "AINFRA-E2400"
		var doctorError *app.DoctorError
		if errors.As(err, &doctorError) && doctorError.Kind == "security" {
			exit = ExitSecurity
			code = "AINFRA-E2405"
		}
		return failDoctorWithExit(
			arguments, output.CommandDoctorTemplate, code, err.Error(), options.IO, exit,
		)
	}
	if err := output.Render(
		options.IO.Stdout,
		output.Success(output.CommandDoctorTemplate, response.Result),
		output.RenderOptions{
			Format: output.Format(response.Format), Style: output.Style(response.OutputStyle),
			Color: output.ColorMode(response.Color), IsTerminal: renderOptions.IsTerminal,
		},
	); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func runDoctorDeployment(
	arguments, renderArguments, controlArguments, positional []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	if options.DoctorDeployment == nil {
		return failDoctor(
			arguments, output.CommandDoctorDeployment,
			"deployment doctor is unavailable", options.IO,
		)
	}
	common, err := parseDoctorEnvironmentRequest(renderArguments, controlArguments)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	target := ""
	if len(positional) == 3 {
		target = positional[2]
	}
	response, err := options.DoctorDeployment(app.DoctorDeploymentRequest{
		Target: target, ConfigPath: common.ConfigPath, ProjectPath: common.ProjectPath,
		Format: common.Format, OutputStyle: common.OutputStyle, Color: common.Color,
		Reconcile: common.Reconcile,
	})
	if err != nil {
		exit := ExitInvalidInput
		code := "AINFRA-E2300"
		var doctorError *app.DoctorError
		if errors.As(err, &doctorError) && doctorError.Kind == "security" {
			exit = ExitSecurity
			code = "AINFRA-E2304"
		}
		return failDoctorWithExit(
			arguments, output.CommandDoctorDeployment, code, err.Error(), options.IO, exit,
		)
	}
	if common.Reconcile && len(response.ReconciliationPlan) > 0 {
		if !confirmReconciliation(response.ReconciliationPlan, common, options.IO) {
			return failDoctorWithExit(
				arguments, output.CommandDoctorDeployment, "AINFRA-E2002",
				"reconciliation requires confirmation", options.IO, ExitInvalidInput,
			)
		}
		response, err = options.DoctorDeployment(app.DoctorDeploymentRequest{
			Target: target, ConfigPath: common.ConfigPath, ProjectPath: common.ProjectPath,
			Format: common.Format, OutputStyle: common.OutputStyle, Color: common.Color,
			Reconcile: true, ApplyReconciliation: true,
		})
		if err != nil {
			return failDoctorWithExit(
				arguments, output.CommandDoctorDeployment, "AINFRA-E2306",
				err.Error(), options.IO, ExitOperationFailed,
			)
		}
	}
	if err := output.Render(
		options.IO.Stdout,
		output.Success(output.CommandDoctorDeployment, response.Result),
		output.RenderOptions{
			Format: output.Format(response.Format), Style: output.Style(response.OutputStyle),
			Color: output.ColorMode(response.Color), IsTerminal: renderOptions.IsTerminal,
		},
	); err != nil {
		return ExitOperationFailed
	}
	return ExitSuccess
}

func splitInvocation(arguments []string) (
	renderArguments, controlArguments, positional []string,
	help bool,
) {
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			positional = append(positional, arguments[index+1:]...)
			break
		}
		if argument == "--help" || argument == "-h" {
			help = true
			continue
		}
		if argument == "--format" || argument == "--output-style" || argument == "--color" {
			renderArguments = append(renderArguments, argument)
			if index+1 < len(arguments) {
				index++
				renderArguments = append(renderArguments, arguments[index])
			}
			continue
		}
		if argument == "--log-level" || argument == "--log-format" || argument == "--log-file" {
			if index+1 < len(arguments) {
				index++
			}
			continue
		}
		if argument == "--syslog" || argument == "-v" || argument == "-vv" || argument == "-vvv" {
			continue
		}
		if stringsHasRenderPrefix(argument) {
			renderArguments = append(renderArguments, argument)
			continue
		}
		if argument == "--config" || argument == "--project" || argument == "--plan" ||
			argument == "--run" || argument == "--source" || argument == "--stream" ||
			argument == "--capability" {
			controlArguments = append(controlArguments, argument)
			if index+1 < len(arguments) {
				index++
				controlArguments = append(controlArguments, arguments[index])
			}
			continue
		}
		if argument == "--reconcile" || argument == "--non-interactive" ||
			argument == "--yes" || argument == "--destroy" || argument == "--check" ||
			argument == "--errors" || argument == "--raw" || argument == "--stdio" {
			controlArguments = append(controlArguments, argument)
			continue
		}
		if stringsHasControlPrefix(argument) {
			controlArguments = append(controlArguments, argument)
			continue
		}
		positional = append(positional, argument)
	}
	return renderArguments, controlArguments, positional, help
}

func stringsHasControlPrefix(argument string) bool {
	return strings.HasPrefix(argument, "--config=") ||
		strings.HasPrefix(argument, "--project=") ||
		strings.HasPrefix(argument, "--capability=") ||
		strings.HasPrefix(argument, "--plan=") ||
		strings.HasPrefix(argument, "--run=")
}

func runDoctorEnvironment(
	arguments, renderArguments, controlArguments []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	if options.DoctorEnvironment == nil {
		return failDoctor(arguments, output.CommandDoctorEnvironment, "environment doctor is unavailable", options.IO)
	}
	request, err := parseDoctorEnvironmentRequest(renderArguments, controlArguments)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	response, err := options.DoctorEnvironment(request)
	if err != nil {
		return failDoctor(arguments, output.CommandDoctorEnvironment, err.Error(), options.IO)
	}
	renderOptions.Format = output.Format(response.Format)
	renderOptions.Style = output.Style(response.OutputStyle)
	renderOptions.Color = output.ColorMode(response.Color)
	envelope := output.Success(output.CommandDoctorEnvironment, response.Result)
	exitCode := ExitSuccess
	if response.Result.Summary.Fail > 0 {
		failed := make([]diagnostic.Diagnostic, 0, response.Result.Summary.Fail)
		for _, finding := range response.Result.Findings {
			if finding.Status == "fail" {
				failed = append(failed, finding)
			}
		}
		envelope = output.PartialFailure(
			output.CommandDoctorEnvironment, response.Result, failed...,
		)
		exitCode = ExitDependency
	}
	if err := output.Render(options.IO.Stdout, envelope, renderOptions); err != nil {
		return ExitOperationFailed
	}
	return exitCode
}

func parseDoctorEnvironmentRequest(
	renderArguments, controlArguments []string,
) (app.DoctorEnvironmentRequest, error) {
	flags := flag.NewFlagSet("doctor environment", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	configPath := flags.String("config", "", "explicit configuration file")
	projectPath := flags.String("project", "", "deployment target")
	reconcileRequested := flags.Bool("reconcile", false, "apply safe local reconciliation")
	nonInteractive := flags.Bool("non-interactive", false, "disable prompts")
	yes := flags.Bool("yes", false, "confirm in non-interactive mode")
	if err := flags.Parse(controlArguments); err != nil {
		return app.DoctorEnvironmentRequest{}, err
	}
	request := app.DoctorEnvironmentRequest{
		ConfigPath: *configPath, ProjectPath: *projectPath,
		Reconcile: *reconcileRequested, NonInteractive: *nonInteractive, Yes: *yes,
	}
	if request.Yes && !request.NonInteractive {
		return app.DoctorEnvironmentRequest{}, errors.New(
			"--yes requires --non-interactive",
		)
	}
	if request.Yes && !request.Reconcile {
		return app.DoctorEnvironmentRequest{}, errors.New(
			"--yes requires --reconcile",
		)
	}
	request.Format = explicitOption(renderArguments, "--format")
	request.OutputStyle = explicitOption(renderArguments, "--output-style")
	request.Color = explicitOption(renderArguments, "--color")
	return request, nil
}

func explicitOption(arguments []string, name string) *string {
	for index, argument := range arguments {
		if argument == name && index+1 < len(arguments) {
			value := arguments[index+1]
			return &value
		}
		if value, found := strings.CutPrefix(argument, name+"="); found {
			return &value
		}
	}
	return nil
}

func failDoctor(arguments []string, command output.Command, message string, streams IO) ExitCode {
	return failDoctorWithExit(
		arguments, command, "AINFRA-E2001", message, streams, ExitInvalidInput,
	)
}

func failDoctorWithExit(
	arguments []string,
	command output.Command,
	code, message string,
	streams IO,
	exit ExitCode,
) ExitCode {
	scope := "environment"
	check := "environment.configuration"
	if command == output.CommandDoctorDeployment {
		scope = "deployment"
		check = "deployment.contract"
	}
	if command == output.CommandDoctorTemplate {
		scope = "template"
		check = "template.contract"
	}
	if command == output.CommandDoctorRun {
		scope = "run"
		check = "run.contract"
	}
	if command == output.CommandDoctorAll {
		scope = "all"
		check = "doctor.all"
	}
	diagnosticValue := diagnostic.Diagnostic{
		Code: code, Severity: diagnostic.SeverityError,
		Check: check, Scope: scope, Status: "fail",
		Reconciliation: "not_available", Component: "configuration",
		Message: message, NextAction: "Correct the configuration and rerun 'ainfra doctor environment'.",
	}
	if requestedJSON(arguments) {
		if err := output.Render(
			streams.Stdout,
			output.Failure(command, diagnosticValue),
			output.RenderOptions{Format: output.FormatJSON},
		); err != nil {
			return ExitOperationFailed
		}
		return exit
	}
	if _, err := fmt.Fprintf(streams.Stderr, "%s: %s\n", diagnosticValue.Code, message); err != nil {
		return ExitOperationFailed
	}
	return exit
}

func stringsHasRenderPrefix(argument string) bool {
	return len(argument) > 9 && (argument[:9] == "--format=" ||
		(len(argument) > 15 && argument[:15] == "--output-style=") ||
		(len(argument) > 8 && argument[:8] == "--color="))
}

func normalizeHelpTopic(positional []string, helpFlag bool) (string, error) {
	if len(positional) == 0 {
		return "ainfra", nil
	}
	if positional[0] == "help" {
		positional = positional[1:]
	}
	if len(positional) == 0 {
		return "ainfra", nil
	}
	if !helpFlag && len(positional) > 2 {
		return "", errors.New("help accepts at most one group and one command topic")
	}
	if len(positional) > 2 {
		return "", errors.New("invalid help topic")
	}
	if len(positional) == 2 {
		return positional[0] + "." + positional[1], nil
	}
	return positional[0], nil
}

func requestedJSON(arguments []string) bool {
	for index, argument := range arguments {
		if argument == "--" {
			return false
		}
		if argument == "--format=json" {
			return true
		}
		if argument == "--format" && index+1 < len(arguments) && arguments[index+1] == "json" {
			return true
		}
	}
	return false
}

func failInvocation(arguments []string, message string, streams IO) ExitCode {
	if requestedJSON(arguments) {
		envelope := output.Failure(output.CommandInvocation, diagnostic.Diagnostic{
			Code:       "AINFRA-E0001",
			Severity:   diagnostic.SeverityError,
			Message:    message,
			Component:  "command",
			NextAction: "Run 'ainfra help' to list valid commands and options.",
		})
		if err := output.Render(streams.Stdout, envelope, output.RenderOptions{Format: output.FormatJSON}); err != nil {
			return ExitOperationFailed
		}
		return ExitInvalidInput
	}
	if _, err := fmt.Fprintf(streams.Stderr, "AINFRA-E0001: %s\nUsage: ainfra <command> [options]\n", message); err != nil {
		return ExitOperationFailed
	}
	return ExitInvalidInput
}

func parseRenderOptions(arguments []string, terminal bool) (output.RenderOptions, error) {
	flags := flag.NewFlagSet("version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	format := flags.String("format", string(output.FormatText), "text or json")
	style := flags.String("output-style", string(output.StyleAuto), "auto, rich, or plain")
	color := flags.String("color", string(output.ColorAuto), "auto, always, or never")
	if err := flags.Parse(arguments); err != nil {
		return output.RenderOptions{}, err
	}
	if flags.NArg() != 0 {
		return output.RenderOptions{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	options := output.RenderOptions{
		Format:     output.Format(*format),
		Style:      output.Style(*style),
		Color:      output.ColorMode(*color),
		IsTerminal: terminal,
	}
	if options.Format != output.FormatText && options.Format != output.FormatJSON {
		return output.RenderOptions{}, fmt.Errorf("invalid --format %q", options.Format)
	}
	if options.Style != output.StyleAuto && options.Style != output.StyleRich && options.Style != output.StylePlain {
		return output.RenderOptions{}, fmt.Errorf("invalid --output-style %q", options.Style)
	}
	if options.Color != output.ColorAuto && options.Color != output.ColorAlways && options.Color != output.ColorNever {
		return output.RenderOptions{}, fmt.Errorf("invalid --color %q", options.Color)
	}
	if options.Format == output.FormatJSON && (options.Style != output.StyleAuto || options.Color != output.ColorAuto) {
		return output.RenderOptions{}, errors.New("--output-style and --color cannot be used with --format json")
	}
	return options, nil
}
