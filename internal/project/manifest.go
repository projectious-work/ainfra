package project

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/projectious-work/ainfra/internal/security"
	"go.yaml.in/yaml/v3"
)

const documentAPIVersion = "ainfra.projectious.work/v1"

var deploymentNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

// Deployment is an immutable, validated view of an ainfra deployment.
type Deployment struct {
	Target   Target
	Metadata Metadata
	Template TemplateReference
	Inputs   Inputs
	SSH      SSH
}

// Metadata identifies a deployment without containing engine variables.
type Metadata struct {
	Name        string
	Description string
}

// TemplateReference points to a template without resolving or acquiring it.
type TemplateReference struct {
	Source string
	Ref    string
}

// Inputs retains ordered pointers to opaque native engine input files.
type Inputs struct {
	TofuVariableFiles      []string
	TofuBackendConfigFiles []string
	AnsibleVariableFiles   []string
}

// SSH retains the optional contained known-hosts path.
type SSH struct {
	KnownHosts string
}

type manifestDocument struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   manifestMetadata `yaml:"metadata"`
	Spec       manifestSpec     `yaml:"spec"`
}

type manifestMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type manifestSpec struct {
	Template manifestTemplate `yaml:"template"`
	Inputs   manifestInputs   `yaml:"inputs,omitempty"`
	SSH      manifestSSH      `yaml:"ssh,omitempty"`
}

type manifestTemplate struct {
	Source string `yaml:"source"`
	Ref    string `yaml:"ref,omitempty"`
}

type manifestInputs struct {
	Tofu    manifestTofu    `yaml:"tofu,omitempty"`
	Ansible manifestAnsible `yaml:"ansible,omitempty"`
}

type manifestTofu struct {
	VariableFiles      []string `yaml:"variableFiles,omitempty"`
	BackendConfigFiles []string `yaml:"backendConfigFiles,omitempty"`
}

type manifestAnsible struct {
	VariableFiles []string `yaml:"variableFiles,omitempty"`
}

type manifestSSH struct {
	KnownHosts string `yaml:"knownHosts,omitempty"`
}

// Load resolves a target and strictly validates its ainfra-owned manifest and
// every declared native-file pointer. Native file contents remain unread.
func Load(options ResolveOptions) (Deployment, error) {
	target, err := Resolve(options)
	if err != nil {
		return Deployment{}, err
	}
	contents, err := os.ReadFile(target.ManifestPath)
	if err != nil {
		return Deployment{}, fmt.Errorf("read deployment manifest: %w", err)
	}
	document, err := decodeManifest(contents)
	if err != nil {
		return Deployment{}, err
	}
	if err := validateManifest(target, document); err != nil {
		return Deployment{}, err
	}
	return Deployment{
		Target: target,
		Metadata: Metadata{
			Name: document.Metadata.Name, Description: document.Metadata.Description,
		},
		Template: TemplateReference{
			Source: document.Spec.Template.Source, Ref: document.Spec.Template.Ref,
		},
		Inputs: Inputs{
			TofuVariableFiles:      clone(document.Spec.Inputs.Tofu.VariableFiles),
			TofuBackendConfigFiles: clone(document.Spec.Inputs.Tofu.BackendConfigFiles),
			AnsibleVariableFiles:   clone(document.Spec.Inputs.Ansible.VariableFiles),
		},
		SSH: SSH{KnownHosts: document.Spec.SSH.KnownHosts},
	}, nil
}

func decodeManifest(contents []byte) (manifestDocument, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var document manifestDocument
	if err := decoder.Decode(&document); err != nil {
		return manifestDocument{}, fmt.Errorf("parse deployment manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return manifestDocument{}, errors.New("deployment manifest must contain exactly one document")
		}
		return manifestDocument{}, fmt.Errorf("parse trailing deployment document: %w", err)
	}
	return document, nil
}

func validateManifest(target Target, document manifestDocument) error {
	if document.APIVersion != documentAPIVersion {
		return fmt.Errorf("unsupported deployment apiVersion %q", document.APIVersion)
	}
	if document.Kind != "Deployment" {
		return fmt.Errorf("deployment kind must be %q", "Deployment")
	}
	if !deploymentNamePattern.MatchString(document.Metadata.Name) {
		return fmt.Errorf("invalid deployment metadata.name %q", document.Metadata.Name)
	}
	if document.Metadata.Description != "" &&
		strings.TrimSpace(document.Metadata.Description) == "" {
		return errors.New("deployment metadata.description must not be blank")
	}
	if strings.TrimSpace(document.Spec.Template.Source) == "" {
		return errors.New("deployment spec.template.source is required")
	}
	groups := []struct {
		name  string
		paths []string
	}{
		{name: "spec.inputs.tofu.variableFiles", paths: document.Spec.Inputs.Tofu.VariableFiles},
		{name: "spec.inputs.tofu.backendConfigFiles", paths: document.Spec.Inputs.Tofu.BackendConfigFiles},
		{name: "spec.inputs.ansible.variableFiles", paths: document.Spec.Inputs.Ansible.VariableFiles},
	}
	for _, group := range groups {
		if err := validateNativePaths(target.Root, group.name, group.paths); err != nil {
			return err
		}
	}
	if document.Spec.SSH.KnownHosts != "" {
		if err := validateRelativeFile(target.Root, "spec.ssh.knownHosts", document.Spec.SSH.KnownHosts); err != nil {
			return err
		}
	}
	return nil
}

func validateNativePaths(root, field string, paths []string) error {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, exists := seen[path]; exists {
			return fmt.Errorf("%s contains duplicate path %q", field, path)
		}
		seen[path] = struct{}{}
		if err := validateRelativeFile(root, field, path); err != nil {
			return err
		}
	}
	return nil
}

func validateRelativeFile(root, field, path string) error {
	if path == "" || filepath.IsAbs(path) {
		return fmt.Errorf("%s path %q must be non-empty and relative", field, path)
	}
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s path %q escapes the deployment root", field, path)
	}
	resolved, err := security.ResolveContained(root, path)
	if err != nil {
		return fmt.Errorf("validate %s path %q: %w", field, path, err)
	}
	if err := security.RequireRegular(resolved); err != nil {
		return fmt.Errorf("validate %s path %q: %w", field, path, err)
	}
	return nil
}

func clone(values []string) []string {
	return append([]string(nil), values...)
}
