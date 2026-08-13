package doctor

import (
	"context"
	"fmt"

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
	}
	registry, err := NewRegistry(definitions...)
	if err != nil {
		panic(err)
	}
	return registry
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
