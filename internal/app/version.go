// Package app coordinates ainfra use cases and transaction boundaries.
package app

import (
	"runtime"

	"github.com/projectious-work/ainfra/internal/output"
)

// Build describes deterministic build facts supplied by the composition root.
type Build struct {
	Version string
	Commit  string
	BuiltAt string
}

// Version returns the semantic version result for this binary.
func Version(build Build) output.Version {
	return output.Version{
		Version:   build.Version,
		Commit:    build.Commit,
		BuiltAt:   build.BuiltAt,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		SupportedContractVersions: output.SupportedContractVersions{
			DocumentAPIVersions:          []string{"ainfra.projectious.work/v1"},
			ResultAPIVersions:            []string{output.APIVersion},
			StandardOutputSchemaVersions: []string{"1"},
		},
	}
}
