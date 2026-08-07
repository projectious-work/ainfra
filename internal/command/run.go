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

var errHelp = errors.New("help requested")

// IO supplies explicit process streams and terminal capabilities.
type IO struct {
	Stdout     io.Writer
	Stderr     io.Writer
	IsTerminal bool
}

// Options supplies immutable composition facts to Run.
type Options struct {
	Build app.Build
	IO    IO
}

// Run parses one CLI invocation, renders its result, and returns its exit code.
func Run(arguments []string, options Options) ExitCode {
	if len(arguments) == 0 {
		if err := writeHelp(options.IO.Stdout); err != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	if arguments[0] == "help" || arguments[0] == "--help" || arguments[0] == "-h" {
		if err := writeHelp(options.IO.Stdout); err != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	versionArguments, isVersion := normalizeVersionInvocation(arguments)
	if !isVersion {
		if _, err := fmt.Fprintf(options.IO.Stderr, "AINFRA-E0001: unknown command %q\n", arguments[0]); err != nil {
			return ExitOperationFailed
		}
		return ExitInvalidInput
	}

	renderOptions, err := parseRenderOptions(versionArguments, options.IO.IsTerminal)
	if errors.Is(err, errHelp) {
		if writeErr := writeVersionHelp(options.IO.Stdout); writeErr != nil {
			return ExitOperationFailed
		}
		return ExitSuccess
	}
	if err != nil {
		if requestedJSON(versionArguments) {
			envelope := output.Failure(output.CommandVersion, diagnostic.Diagnostic{
				Code:       "AINFRA-E0002",
				Severity:   diagnostic.SeverityError,
				Message:    err.Error(),
				Component:  "command",
				NextAction: "Run 'ainfra version --help' for valid syntax.",
			})
			if renderErr := output.Render(options.IO.Stdout, envelope, output.RenderOptions{Format: output.FormatJSON}); renderErr != nil {
				return ExitOperationFailed
			}
			return ExitInvalidInput
		}
		if _, writeErr := fmt.Fprintf(options.IO.Stderr, "AINFRA-E0002: %s\n", err); writeErr != nil {
			return ExitOperationFailed
		}
		return ExitInvalidInput
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

func normalizeVersionInvocation(arguments []string) ([]string, bool) {
	for index, argument := range arguments {
		if argument != "version" && argument != "--version" {
			continue
		}
		options := make([]string, 0, len(arguments)-1)
		options = append(options, arguments[:index]...)
		options = append(options, arguments[index+1:]...)
		return options, true
	}
	return nil, false
}

func requestedJSON(arguments []string) bool {
	for index, argument := range arguments {
		if argument == "--format=json" {
			return true
		}
		if argument == "--format" && index+1 < len(arguments) && arguments[index+1] == "json" {
			return true
		}
	}
	return false
}

func parseRenderOptions(arguments []string, terminal bool) (output.RenderOptions, error) {
	flags := flag.NewFlagSet("version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	format := flags.String("format", string(output.FormatText), "text or json")
	style := flags.String("output-style", string(output.StyleAuto), "auto, rich, or plain")
	color := flags.String("color", string(output.ColorAuto), "auto, always, or never")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return output.RenderOptions{}, errHelp
		}
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

func writeHelp(writer io.Writer) error {
	_, err := io.WriteString(writer, strings.TrimSpace(`ainfra manages native OpenTofu and Ansible infrastructure templates.

Usage:
  ainfra help
  ainfra version [--format text|json] [--output-style auto|rich|plain]
                 [--color auto|always|never]

Commands:
  help       Show command help
  version    Show version and build information
`)+"\n")
	return err
}

func writeVersionHelp(writer io.Writer) error {
	_, err := io.WriteString(writer, "Usage: ainfra version [--format text|json] [--output-style auto|rich|plain] [--color auto|always|never]\n")
	return err
}
