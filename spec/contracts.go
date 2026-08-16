// Package contracts embeds the published ainfra contracts shipped with the
// binary. Callers can read only the explicitly registered schema names.
package contracts

import "embed"

//go:embed schemas/v1/*.json
var files embed.FS

var schemaNames = []string{
	"ainfra.schema.json",
	"config.schema.json",
	"engine-evidence-profile.schema.json",
	"lock.schema.json",
	"machine-output.schema.json",
	"plan-record.schema.json",
	"roadmap.schema.json",
	"run-record.schema.json",
	"standard-output.schema.json",
	"template-manifest.schema.json",
}

// Schemas returns the closed set of published v1 JSON schema names.
func Schemas() []string { return append([]string(nil), schemaNames...) }

// Schema returns one embedded published schema. Unknown names are refused.
func Schema(name string) ([]byte, bool) {
	known := false
	for _, candidate := range schemaNames {
		if name == candidate {
			known = true
			break
		}
	}
	if !known {
		return nil, false
	}
	contents, err := files.ReadFile("schemas/v1/" + name)
	if err != nil {
		return nil, false
	}
	return contents, true
}
