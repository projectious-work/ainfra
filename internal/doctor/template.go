package doctor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

// TemplateInput contains only facts from a validated local template contract.
type TemplateInput struct {
	Name       string
	Version    string
	Root       string
	HasAnsible bool
	Inventory  string
}

// TemplateRegistry returns checks for an already-resolved local template.
func TemplateRegistry(input TemplateInput) Registry {
	definitions := []Definition{
		templateFact("template.identity", "AINFRA-E2401", input.Root,
			fmt.Sprintf("template %q version %s matches its directory", input.Name, input.Version)),
		templateFact("template.layout", "AINFRA-E2402", input.Root,
			"required documentation, native engine paths, and minimal example are present"),
		templateFact("template.engine-ownership", "AINFRA-E2403", input.Root,
			"native engine files and dependency declarations remain engine-owned"),
		templateFact("template.inventory-applicability", "AINFRA-E2404", input.Root,
			fmt.Sprintf("inventory %q and Ansible applicability are consistent", input.Inventory)),
		templateAuthoringLayout(input.Root),
		templateVariableReference(input.Root),
		templateAuthoringReadme(input.Root),
		templateChildTools(),
	}
	registry, err := NewRegistry(definitions...)
	if err != nil {
		panic(err)
	}
	return registry
}

func templateAuthoringLayout(root string) Definition {
	return Definition{ID: "template.authoring-layout", Scope: ScopeTemplate,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			required := []string{
				"tofu/variables.tf", "tofu/outputs.tf", "tofu/versions.tf",
				"tests/README.md", "tests/validate.sh", "tests/fixtures/output.json",
			}
			for _, path := range required {
				if info, err := templateStat(root, path); err != nil || info.IsDir() {
					return templateFailure("AINFRA-E2406", root,
						fmt.Sprintf("template authoring layout requires %s", path))
				}
			}
			return templatePass("AINFRA-E2406", root,
				"native variable/output declarations and clean-room conformance fixtures are present")
		}}
}

func templateVariableReference(root string) Definition {
	return Definition{ID: "template.variable-reference", Scope: ScopeTemplate,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			contents, err := templateRead(root, "docs/variables.md")
			if err != nil {
				return templateFailure("AINFRA-E2407", root, "read standard variable reference")
			}
			sections := []string{
				"## OpenTofu variables", "## Ansible variables",
				"## Cross-variable rules", "## Examples",
			}
			last := -1
			for _, section := range sections {
				position := bytes.Index(contents, []byte(section))
				if position < 0 || position < last {
					return templateFailure("AINFRA-E2407", root,
						"standard variable reference sections are missing or out of order")
				}
				last = position
			}
			for _, column := range []string{
				"Name", "Type or shape", "Required", "Default",
				"Valid values and constraints", "Sensitive", "Description",
			} {
				if !bytes.Contains(contents, []byte(column)) {
					return templateFailure("AINFRA-E2407", root,
						"standard variable reference table columns are incomplete")
				}
			}
			return templatePass("AINFRA-E2407", root,
				"standard native-variable reference structure is present")
		}}
}

func templateAuthoringReadme(root string) Definition {
	return Definition{ID: "template.authoring-readme", Scope: ScopeTemplate,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			contents, err := templateRead(root, "README.md")
			if err != nil {
				return templateFailure("AINFRA-E2408", root, "read template README")
			}
			for _, section := range []string{
				"Prerequisites", "Architecture", "Variables", "Network", "Cost",
				"Failure", "Teardown", "Compatibility", "Validation",
			} {
				if !bytes.Contains(bytes.ToLower(contents), []byte(strings.ToLower(section))) {
					return templateFailure("AINFRA-E2408", root,
						"template README is missing required authoring guidance: "+section)
				}
			}
			return templatePass("AINFRA-E2408", root,
				"template README covers required authoring and lifecycle guidance")
		}}
}

func templateChildTools() Definition {
	return Definition{ID: "template.native-tool-validation", Scope: ScopeTemplate,
		ChildToolNeeds: []string{"tofu", "ansible-playbook"},
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			return diagnostic.Diagnostic{Code: "AINFRA-E2409", Severity: diagnostic.SeverityInfo,
				Status: "skip", Component: "template",
				Message:    "native OpenTofu and Ansible validation is delegated to the template's clean-room test script",
				NextAction: "Run tests/validate.sh with compatible OpenTofu and Ansible tools."}
		}}
}

func templatePass(code, path, message string) diagnostic.Diagnostic {
	return diagnostic.Diagnostic{Code: code, Severity: diagnostic.SeverityInfo, Status: "pass",
		Component: "template", Path: path, Message: message}
}

func templateFailure(code, path, message string) diagnostic.Diagnostic {
	return diagnostic.Diagnostic{Code: code, Severity: diagnostic.SeverityError, Status: "fail",
		Component: "template", Path: path, Message: message,
		NextAction: "Follow the published template-authoring guide and rerun ainfra doctor template."}
}

func templateRead(root, path string) ([]byte, error) {
	rootDirectory, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rootDirectory.Close() }()
	return rootDirectory.ReadFile(path)
}

func templateStat(root, path string) (os.FileInfo, error) {
	rootDirectory, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rootDirectory.Close() }()
	return rootDirectory.Stat(path)
}

func templateFact(id, code, path, message string) Definition {
	return Definition{
		ID: id, Scope: ScopeTemplate,
		Run: func(context.Context, Input, Capabilities) diagnostic.Diagnostic {
			return diagnostic.Diagnostic{
				Code: code, Severity: diagnostic.SeverityInfo, Status: "pass",
				Component: "template", Path: path, Message: message,
			}
		},
	}
}
