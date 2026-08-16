package command

import "github.com/projectious-work/ainfra/internal/output"

func helpFor(topic string) (output.Help, bool) {
	options := func(extra ...output.HelpOption) []output.HelpOption {
		return append(extra,
			output.HelpOption{Names: []string{"-v", "-vv", "-vvv"}, Summary: "Increase operational log verbosity."},
			output.HelpOption{Names: []string{"--log-level"}, ValueName: "LEVEL", Summary: "Set operational log level."},
			output.HelpOption{Names: []string{"--log-format"}, ValueName: "text|json", Summary: "Set operational sink format."},
			output.HelpOption{Names: []string{"--log-file"}, ValueName: "PATH", Summary: "Add a rotating file sink."},
			output.HelpOption{Names: []string{"--syslog"}, Summary: "Enable the local syslog sink."},
			output.HelpOption{Names: []string{"--format"}, ValueName: "text|json", Summary: "Select output format."},
			output.HelpOption{Names: []string{"--output-style"}, ValueName: "auto|rich|plain", Summary: "Select text presentation."},
			output.HelpOption{Names: []string{"--color"}, ValueName: "auto|always|never", Summary: "Control terminal color."},
			output.HelpOption{Names: []string{"--help", "-h"}, Summary: "Show static help."},
		)
	}
	argument := func(name, summary string, required bool) []output.HelpArgument {
		return []output.HelpArgument{{Name: name, Required: required, Summary: summary}}
	}
	command := func(topic, usage, summary string, arguments []output.HelpArgument, extra ...output.HelpOption) output.Help {
		if arguments == nil {
			arguments = []output.HelpArgument{}
		}
		return output.Help{Topic: topic, Usage: usage, Summary: summary,
			Subcommands: []output.HelpNamedItem{}, Arguments: arguments, Options: options(extra...)}
	}

	topics := map[string]output.Help{
		"ainfra": {Topic: "ainfra", Usage: "ainfra <command> [options]",
			Summary: "Manage native OpenTofu and Ansible infrastructure templates.",
			Subcommands: []output.HelpNamedItem{
				{Name: "help", Summary: "Show static command help."},
				{Name: "init", Summary: "Create a minimal deployment definition."},
				{Name: "doctor", Summary: "Diagnose ainfra contracts and prerequisites."},
				{Name: "template", Summary: "Manage immutable template sources."},
				{Name: "plan", Summary: "Create a saved apply or destroy plan."},
				{Name: "apply", Summary: "Apply an exact reviewed plan."},
				{Name: "configure", Summary: "Run Ansible configuration."},
				{Name: "deploy", Summary: "Run the applicable deployment stages."},
				{Name: "output", Summary: "Show standardized template output."},
				{Name: "inventory", Summary: "Generate deterministic inventory."},
				{Name: "logs", Summary: "Show retained run evidence."},
				{Name: "status", Summary: "Show deployment and run status."},
				{Name: "mcp", Summary: "Serve guarded MCP capabilities."},
				{Name: "destroy", Summary: "Apply an exact reviewed destroy plan."},
				{Name: "version", Summary: "Show version and build information."},
			}, Arguments: []output.HelpArgument{}, Options: options()},
		"doctor": {Topic: "doctor", Usage: "ainfra doctor [--reconcile]",
			Summary: "Diagnose ainfra contracts and execution prerequisites.",
			Subcommands: []output.HelpNamedItem{
				{Name: "all", Summary: "Run all applicable diagnostics."},
				{Name: "deployment", Summary: "Diagnose a deployment."},
				{Name: "template", Summary: "Diagnose a resolved template."},
				{Name: "run", Summary: "Diagnose retained run evidence."},
				{Name: "environment", Summary: "Diagnose the execution environment."},
			}, Arguments: []output.HelpArgument{}, Options: options(reconcileOptions()...)},
		"template": {Topic: "template", Usage: "ainfra template <command> [options]",
			Summary: "Manage immutable template sources.",
			Subcommands: []output.HelpNamedItem{
				{Name: "lock", Summary: "Lock a template source."},
				{Name: "update", Summary: "Update a template lock."},
				{Name: "migrate", Summary: "Preview a contract migration."},
			}, Arguments: []output.HelpArgument{}, Options: options()},
		"mcp": {Topic: "mcp", Usage: "ainfra mcp <command> [options]",
			Summary: "Serve guarded Model Context Protocol capabilities.",
			Subcommands: []output.HelpNamedItem{
				{Name: "serve", Summary: "Serve the allowlisted registry."},
			}, Arguments: []output.HelpArgument{}, Options: options()},
		"mcp.serve": command("mcp.serve", "ainfra mcp serve --stdio [--project PATH]",
			"Serve read-only MCP tools over standard input and output.", nil,
			output.HelpOption{Names: []string{"--stdio"}, Summary: "Use newline-delimited stdio transport."},
			output.HelpOption{Names: []string{"--project"}, ValueName: "PATH", Summary: "Select the allowed project root."}),
		"init":               command("init", "ainfra init [DEPLOYMENT] [options]", "Create a minimal deployment definition.", argument("DEPLOYMENT", "Deployment directory.", false)),
		"doctor.all":         command("doctor.all", "ainfra doctor all [TARGET] [--reconcile]", "Run all applicable diagnostics.", argument("TARGET", "Deployment or project target.", false), reconcileOptions()...),
		"doctor.deployment":  command("doctor.deployment", "ainfra doctor deployment [TARGET] [--reconcile]", "Diagnose a deployment.", argument("TARGET", "Deployment or project target.", false), reconcileOptions()...),
		"doctor.template":    command("doctor.template", "ainfra doctor template [TARGET] [--reconcile]", "Diagnose a resolved template.", argument("TARGET", "Deployment or project target.", false), reconcileOptions()...),
		"doctor.run":         command("doctor.run", "ainfra doctor run [TARGET] [--reconcile]", "Diagnose retained run evidence.", argument("TARGET", "Deployment or project target.", false), reconcileOptions()...),
		"doctor.environment": command("doctor.environment", "ainfra doctor environment [--reconcile]", "Diagnose the execution environment.", nil, reconcileOptions()...),
		"template.lock":      command("template.lock", "ainfra template lock [DEPLOYMENT] [options]", "Lock a template source.", argument("DEPLOYMENT", "Deployment directory or manifest.", false)),
		"template.update":    command("template.update", "ainfra template update [DEPLOYMENT] [options]", "Update a template lock.", argument("DEPLOYMENT", "Deployment directory or manifest.", false)),
		"template.migrate":   command("template.migrate", "ainfra template migrate SOURCE --to VERSION [--write]", "Preview a template contract migration.", argument("SOURCE", "Local template working copy.", true), output.HelpOption{Names: []string{"--to"}, ValueName: "VERSION", Summary: "Select the target contract version."}, output.HelpOption{Names: []string{"--write"}, Summary: "Apply safe deterministic changes."}),
		"plan":               command("plan", "ainfra plan [DEPLOYMENT] [--destroy] [options]", "Create a saved apply or destroy plan.", argument("DEPLOYMENT", "Deployment directory or ainfra.yaml path.", false), output.HelpOption{Names: []string{"--destroy"}, Summary: "Create a destroy plan."}),
		"apply":              command("apply", "ainfra apply [DEPLOYMENT] --plan RUN_ID [options]", "Apply an exact reviewed plan.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--plan"}, ValueName: "RUN_ID", Summary: "Select the reviewed plan."}),
		"configure":          command("configure", "ainfra configure [DEPLOYMENT] --run RUN_ID [--check]", "Run Ansible configuration.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--run"}, ValueName: "RUN_ID", Summary: "Select the applied run."}, output.HelpOption{Names: []string{"--check"}, Summary: "Run Ansible check mode."}),
		"deploy":             command("deploy", "ainfra deploy [DEPLOYMENT] --plan RUN_ID [options]", "Run applicable deployment stages.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--plan"}, ValueName: "RUN_ID", Summary: "Select the reviewed plan."}),
		"output":             command("output", "ainfra output [DEPLOYMENT] --run RUN_ID", "Show standardized template output.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--run"}, ValueName: "RUN_ID", Summary: "Select the retained run."}),
		"inventory":          command("inventory", "ainfra inventory [DEPLOYMENT] --run RUN_ID", "Generate deterministic inventory.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--run"}, ValueName: "RUN_ID", Summary: "Select the retained run."}),
		"logs":               command("logs", "ainfra logs [DEPLOYMENT] --run RUN_ID [--source SOURCE] [--errors|--raw --stream STREAM]", "Show retained run evidence.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--run"}, ValueName: "RUN_ID", Summary: "Select the retained run."}, output.HelpOption{Names: []string{"--source"}, ValueName: "SOURCE", Summary: "Select an evidence source."}, output.HelpOption{Names: []string{"--errors"}, Summary: "Show attributed errors."}, output.HelpOption{Names: []string{"--raw"}, Summary: "Show sensitive raw evidence."}, output.HelpOption{Names: []string{"--stream"}, ValueName: "STREAM", Summary: "Select a retained stream."}, output.HelpOption{Names: []string{"--non-interactive"}, Summary: "Disable raw-access prompts."}, output.HelpOption{Names: []string{"--yes"}, Summary: "Confirm raw access with --non-interactive."}),
		"status":             command("status", "ainfra status [DEPLOYMENT] [options]", "Show deployment and run status.", argument("DEPLOYMENT", "Deployment directory or manifest.", false)),
		"destroy":            command("destroy", "ainfra destroy [DEPLOYMENT] --plan RUN_ID [options]", "Apply an exact reviewed destroy plan.", argument("DEPLOYMENT", "Deployment directory or manifest.", false), output.HelpOption{Names: []string{"--plan"}, ValueName: "RUN_ID", Summary: "Select the reviewed destroy plan."}),
		"version":            command("version", "ainfra version [options]", "Show version and build information.", nil),
	}
	result, found := topics[topic]
	return result, found
}

func reconcileOptions() []output.HelpOption {
	return []output.HelpOption{
		{Names: []string{"--reconcile"}, Summary: "Apply safe local reconciliation."},
		{Names: []string{"--non-interactive"}, Summary: "Disable confirmation prompts."},
		{Names: []string{"--yes"}, Summary: "Approve reconciliation with --non-interactive."},
	}
}
