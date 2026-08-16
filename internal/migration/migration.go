// Package migration plans explicit, deterministic template-contract version
// transitions. It never owns provider, backend, state, or remote mutations.
package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectious-work/ainfra/internal/template"
	"go.yaml.in/yaml/v3"
)

const (
	// VersionV1 is the only template API currently supported by ainfra.
	VersionV1 = "v1"
	v1API     = "ainfra.projectious.work/v1"
)

// Change is one ordered, reviewable source transformation. No automatic v1
// transformations exist while v1 is the sole supported contract.
type Change struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Automatic   bool   `json:"automatic"`
}

// Plan is a deterministic migration analysis result.
type Plan struct {
	SourceVersion string   `json:"sourceVersion"`
	TargetVersion string   `json:"targetVersion"`
	Changes       []Change `json:"changes"`
}

type manifestIdentity struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// Analyze validates one local template working copy and selects an explicit
// registered source-to-target migrator. Unsupported transitions fail closed.
func Analyze(sourceRoot, targetVersion string) (Plan, error) {
	if targetVersion != VersionV1 {
		return Plan{}, fmt.Errorf("unsupported template migration target %q", targetVersion)
	}
	root, err := filepath.Abs(sourceRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve template migration source: %w", err)
	}
	information, err := os.Lstat(root)
	if err != nil || !information.IsDir() || information.Mode()&os.ModeSymlink != 0 {
		return Plan{}, errors.New("template migration source must be a non-symlink local directory")
	}
	contents, err := os.ReadFile(filepath.Join(root, "ainfra-template.yaml"))
	if err != nil {
		return Plan{}, fmt.Errorf("read template migration manifest: %w", err)
	}
	var identity manifestIdentity
	if err := yaml.Unmarshal(contents, &identity); err != nil {
		return Plan{}, fmt.Errorf("parse template migration manifest identity: %w", err)
	}
	if identity.APIVersion != v1API || identity.Kind != "Template" {
		return Plan{}, errors.New("unsupported or ambiguous template source contract version")
	}
	if _, err := template.Load(root); err != nil {
		return Plan{}, fmt.Errorf("validate template migration source: %w", err)
	}
	return Plan{SourceVersion: VersionV1, TargetVersion: VersionV1,
		Changes: []Change{}}, nil
}
