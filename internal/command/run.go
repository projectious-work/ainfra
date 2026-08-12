package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/output"
)

// IO supplies explicit process streams and terminal capabilities.
type IO struct {
	Stdout     io.Writer
	Stderr     io.Writer
	IsTerminal bool
}

// Options supplies immutable composition facts to Run.
type Options struct {
	Build             app.Build
	IO                IO
	DoctorEnvironment func(app.DoctorEnvironmentRequest) (app.DoctorEnvironmentResponse, error)
}

// Run parses one CLI invocation, renders its result, and returns its exit code.
func Run(arguments []string, options Options) ExitCode {
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
		if stringsHasRenderPrefix(argument) {
			renderArguments = append(renderArguments, argument)
			continue
		}
		if argument == "--config" || argument == "--project" {
			controlArguments = append(controlArguments, argument)
			if index+1 < len(arguments) {
				index++
				controlArguments = append(controlArguments, arguments[index])
			}
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
	return len(argument) > 9 && (argument[:9] == "--config=" ||
		(len(argument) > 10 && argument[:10] == "--project="))
}

func runDoctorEnvironment(
	arguments, renderArguments, controlArguments []string,
	renderOptions output.RenderOptions,
	options Options,
) ExitCode {
	if options.DoctorEnvironment == nil {
		return failDoctorEnvironment(arguments, "environment doctor is unavailable", options.IO)
	}
	request, err := parseDoctorEnvironmentRequest(renderArguments, controlArguments)
	if err != nil {
		return failInvocation(arguments, err.Error(), options.IO)
	}
	response, err := options.DoctorEnvironment(request)
	if err != nil {
		return failDoctorEnvironment(arguments, err.Error(), options.IO)
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
	if err := flags.Parse(controlArguments); err != nil {
		return app.DoctorEnvironmentRequest{}, err
	}
	request := app.DoctorEnvironmentRequest{
		ConfigPath: *configPath, ProjectPath: *projectPath,
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

func failDoctorEnvironment(arguments []string, message string, streams IO) ExitCode {
	diagnosticValue := diagnostic.Diagnostic{
		Code: "AINFRA-E2001", Severity: diagnostic.SeverityError,
		Check: "environment.configuration", Scope: "environment", Status: "fail",
		Reconciliation: "not_available", Component: "configuration",
		Message: message, NextAction: "Correct the configuration and rerun 'ainfra doctor environment'.",
	}
	if requestedJSON(arguments) {
		if err := output.Render(
			streams.Stdout,
			output.Failure(output.CommandDoctorEnvironment, diagnosticValue),
			output.RenderOptions{Format: output.FormatJSON},
		); err != nil {
			return ExitOperationFailed
		}
		return ExitInvalidInput
	}
	if _, err := fmt.Fprintf(streams.Stderr, "%s: %s\n", diagnosticValue.Code, message); err != nil {
		return ExitOperationFailed
	}
	return ExitInvalidInput
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
