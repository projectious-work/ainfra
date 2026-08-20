package contracts_test

import (
	"os"
	"strings"
	"testing"
)

func TestAITemplateAuthoringEntryPointIsSelfContained(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../docs/content/docs/template-authoring-ai.md")
	if err != nil {
		t.Fatal(err)
	}
	guide := string(contents)
	for _, required := range []string{
		"apiVersion: ainfra.projectious.work/v1",
		"kind: Template",
		"## Native variables and dependencies",
		"## Standard output",
		"## README and security",
		"## Local validation",
		"## Disposable live acceptance",
		"ainfra doctor template . --format json",
		"./tests/validate.sh",
		"explicit, immediate approval",
		"independent teardown verification",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("AI authoring guide is missing %q", required)
		}
	}
	if strings.Contains(strings.ToLower(guide), "and similar") {
		t.Error("AI authoring guide contains an ambiguous supported-value phrase")
	}
	for _, path := range []string{
		"schemas/v1/template-manifest.schema.json",
		"schemas/v1/standard-output.schema.json",
		"examples/v1/template-example/ainfra-template.yaml",
		"examples/v1/template-clean-room/ainfra-template.yaml",
		"examples/v1/template-clean-room/tests/validate.sh",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("referenced authoring resource %q: %v", path, err)
		}
	}
}
