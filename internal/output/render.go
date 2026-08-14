package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Style selects a human-readable presentation style.
type Style string

const (
	// StyleAuto selects rich output only for a capable terminal.
	StyleAuto Style = "auto"
	// StyleRich selects decorated terminal output.
	StyleRich Style = "rich"
	// StylePlain selects stable undecorated output.
	StylePlain Style = "plain"
)

// Format selects human-oriented text or the machine-result protocol.
type Format string

const (
	// FormatText selects human-readable output.
	FormatText Format = "text"
	// FormatJSON selects the versioned machine-result envelope.
	FormatJSON Format = "json"
)

// ColorMode controls whether terminal color may be used.
type ColorMode string

const (
	// ColorAuto enables color only when terminal capabilities permit it.
	ColorAuto ColorMode = "auto"
	// ColorAlways requests color output.
	ColorAlways ColorMode = "always"
	// ColorNever disables color output.
	ColorNever ColorMode = "never"
)

// RenderOptions describe the selected presentation without changing results.
type RenderOptions struct {
	Format     Format
	Style      Style
	Color      ColorMode
	IsTerminal bool
}

// Render writes one semantic result in the selected presentation.
func Render(writer io.Writer, envelope Envelope, options RenderOptions) error {
	if options.Format == FormatJSON {
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(envelope)
	}

	style := options.Style
	if style == StyleAuto {
		if options.IsTerminal {
			style = StyleRich
		} else {
			style = StylePlain
		}
	}

	switch result := envelope.Result.(type) {
	case Plan:
		_, err := fmt.Fprintf(
			writer, "saved %s plan %s for %s\nplan digest %s\nnext: %s\n",
			result.Intent, result.RunID, result.Deployment.Name, result.PlanDigest,
			strings.Join(result.NextCommands, "\nnext: "),
		)
		return err
	case Execution:
		_, err := fmt.Fprintf(writer, "%s %s for %s: %s\n",
			result.Operation, result.RunID, result.Deployment.Name, result.ExecutionOutcome)
		return err
	case Artifact:
		if result.Applicability == "not-applicable" {
			_, err := fmt.Fprintf(writer, "%s for %s: not applicable (%s)\n",
				envelope.Command, result.Deployment.Name, result.Reason)
			return err
		}
		_, err := fmt.Fprintf(writer, "%s %s for %s\ndigest %s\n",
			result.Artifact.Kind, result.Artifact.Path, result.Deployment.Name,
			result.ContentDigest)
		return err
	case Template:
		state := "unchanged"
		if result.Changed {
			state = "written"
		}
		operation := "lock"
		if envelope.Command == CommandTemplateUpdate {
			operation = "update"
		}
		_, err := fmt.Fprintf(
			writer, "template %s %s\nsource %s\nresolved %s\ndigest %s\n",
			operation, state, result.Source, result.ResolvedRevision, result.ContentDigest,
		)
		return err
	case Init:
		_, err := fmt.Fprintf(
			writer, "initialized deployment %s at %s\n", result.Deployment.Name,
			result.Deployment.Root,
		)
		if err != nil {
			return err
		}
		for _, path := range result.CreatedPaths {
			if _, err = fmt.Fprintf(writer, "created %s\n", path); err != nil {
				return err
			}
		}
		return nil
	case Doctor:
		_, err := fmt.Fprintf(
			writer,
			"doctor %s: %d pass, %d skip, %d warning, %d fail\n",
			result.Scope, result.Summary.Pass, result.Summary.Skip,
			result.Summary.Warning, result.Summary.Fail,
		)
		if err != nil {
			return err
		}
		for _, finding := range result.Findings {
			if _, err = fmt.Fprintf(writer, "%s %s: %s\n", finding.Status, finding.Code, finding.Message); err != nil {
				return err
			}
			if finding.NextAction != "" {
				if _, err = fmt.Fprintf(writer, "  next: %s\n", finding.NextAction); err != nil {
					return err
				}
			}
		}
		if result.EffectiveConfiguration != nil {
			keys := make([]string, 0, len(result.EffectiveConfiguration.Values))
			for key := range result.EffectiveConfiguration.Values {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				value := result.EffectiveConfiguration.Values[key]
				if _, err = fmt.Fprintf(
					writer, "%s = %s (%s)\n", key, value.DisplayValue, value.Source,
				); err != nil {
					return err
				}
			}
		}
		return nil
	case Help:
		_, err := fmt.Fprintf(writer, "%s\n\nUsage:\n  %s\n", result.Summary, result.Usage)
		if err != nil {
			return err
		}
		if len(result.Subcommands) > 0 {
			if _, err = io.WriteString(writer, "\nCommands:\n"); err != nil {
				return err
			}
			for _, command := range result.Subcommands {
				if _, err = fmt.Fprintf(writer, "  %-18s %s\n", command.Name, command.Summary); err != nil {
					return err
				}
			}
		}
		if len(result.Arguments) > 0 {
			if _, err = io.WriteString(writer, "\nArguments:\n"); err != nil {
				return err
			}
			for _, argument := range result.Arguments {
				if _, err = fmt.Fprintf(writer, "  %-18s %s\n", argument.Name, argument.Summary); err != nil {
					return err
				}
			}
		}
		if len(result.Options) > 0 {
			if _, err = io.WriteString(writer, "\nOptions:\n"); err != nil {
				return err
			}
			for _, option := range result.Options {
				names := strings.Join(option.Names, ", ")
				if option.ValueName != "" {
					names += " " + option.ValueName
				}
				if _, err = fmt.Fprintf(writer, "  %-18s %s\n", names, option.Summary); err != nil {
					return err
				}
			}
		}
		return nil
	case Version:
		contracts := fmt.Sprintf("documents %s; results %s; standard output %s",
			strings.Join(result.SupportedContractVersions.DocumentAPIVersions, ", "),
			strings.Join(result.SupportedContractVersions.ResultAPIVersions, ", "),
			strings.Join(result.SupportedContractVersions.StandardOutputSchemaVersions, ", "))
		if style == StyleRich {
			_, err := fmt.Fprintf(writer, "ainfra %s\n  commit: %s\n  built: %s\n  go: %s\n  platform: %s\n  contracts: %s\n", result.Version, result.Commit, result.BuiltAt, result.GoVersion, result.Platform, contracts)
			return err
		}
		_, err := fmt.Fprintf(writer, "ainfra %s\ncommit %s\nbuilt %s\ngo %s\nplatform %s\ncontracts %s\n", result.Version, result.Commit, result.BuiltAt, result.GoVersion, result.Platform, contracts)
		return err
	default:
		return fmt.Errorf("render unsupported result for %s", envelope.Command)
	}
}
