// Package template validates already-resolved local template contracts.
package template

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
	"go.yaml.in/yaml/v3"
)

const manifestName = "ainfra-template.yaml"

var (
	namePattern    = regexp.MustCompile(`^[a-z][a-z0-9-]{2,62}$`)
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
)

// Contract is an immutable validated view of a local template.
type Contract struct {
	Root         string
	ManifestPath string
	Name         string
	Version      string
	Tofu         Engine
	Ansible      *AnsibleEngine
	Inventory    string
}

// Engine declares one contained engine working directory and version range.
type Engine struct {
	Directory string
	Version   string
}

// AnsibleEngine adds the contained native playbook to an engine declaration.
type AnsibleEngine struct {
	Engine
	Playbook string
}

type document struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   metadata `yaml:"metadata"`
	Spec       spec     `yaml:"spec"`
}

type metadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

type spec struct {
	Engines engines `yaml:"engines"`
	Outputs outputs `yaml:"outputs"`
}

type engines struct {
	Tofu    engine         `yaml:"tofu"`
	Ansible *ansibleEngine `yaml:"ansible,omitempty"`
}

type engine struct {
	Directory string `yaml:"directory"`
	Version   string `yaml:"version"`
}

type ansibleEngine struct {
	Directory string `yaml:"directory"`
	Playbook  string `yaml:"playbook"`
	Version   string `yaml:"version"`
}

type outputs struct {
	Inventory string `yaml:"inventory"`
}

// Load strictly validates one already-resolved local template directory.
func Load(path string) (Contract, error) {
	root, err := canonicalDirectory(path)
	if err != nil {
		return Contract{}, err
	}
	manifestPath, err := requiredFile(root, manifestName)
	if err != nil {
		return Contract{}, err
	}
	contents, err := readRootFile(root, manifestName)
	if err != nil {
		return Contract{}, fmt.Errorf("read template manifest: %w", err)
	}
	document, err := decode(contents)
	if err != nil {
		return Contract{}, err
	}
	if err := validateDocument(root, document); err != nil {
		return Contract{}, err
	}
	tofuDirectory, err := requiredDirectory(root, document.Spec.Engines.Tofu.Directory)
	if err != nil {
		return Contract{}, fmt.Errorf("validate OpenTofu directory: %w", err)
	}
	_ = tofuDirectory
	for _, path := range []string{"README.md", "docs/variables.md", "examples/minimal/ainfra.yaml"} {
		if _, err := requiredFile(root, path); err != nil {
			return Contract{}, err
		}
	}
	if _, err := project.Load(project.ResolveOptions{
		ExplicitPath: filepath.Join(root, "examples", "minimal"),
	}); err != nil {
		return Contract{}, fmt.Errorf("validate minimal deployment example: %w", err)
	}
	readme, err := readRootFile(root, "README.md")
	if err != nil {
		return Contract{}, fmt.Errorf("read template README: %w", err)
	}
	if !bytes.Contains(readme, []byte("tofu ")) {
		return Contract{}, errors.New("template README must document direct OpenTofu commands")
	}
	contract := Contract{
		Root: root, ManifestPath: manifestPath, Name: document.Metadata.Name,
		Version: document.Metadata.Version,
		Tofu: Engine{
			Directory: document.Spec.Engines.Tofu.Directory,
			Version:   document.Spec.Engines.Tofu.Version,
		},
		Inventory: document.Spec.Outputs.Inventory,
	}
	if document.Spec.Engines.Ansible != nil {
		declaration := document.Spec.Engines.Ansible
		ansibleDirectory, dirErr := requiredDirectory(root, declaration.Directory)
		if dirErr != nil {
			return Contract{}, fmt.Errorf("validate Ansible directory: %w", dirErr)
		}
		if _, fileErr := requiredFile(ansibleDirectory, declaration.Playbook); fileErr != nil {
			return Contract{}, fmt.Errorf("validate Ansible playbook: %w", fileErr)
		}
		if _, fileErr := requiredFile(ansibleDirectory, "requirements.yml"); fileErr != nil {
			return Contract{}, fmt.Errorf("validate Ansible requirements: %w", fileErr)
		}
		if !bytes.Contains(readme, []byte("ansible-playbook ")) {
			return Contract{}, errors.New("template README must document direct Ansible commands")
		}
		contract.Ansible = &AnsibleEngine{
			Engine:   Engine{Directory: declaration.Directory, Version: declaration.Version},
			Playbook: declaration.Playbook,
		}
	}
	if err := rejectRuntimeState(root); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func readRootFile(root, relative string) ([]byte, error) {
	rootDirectory, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open template root: %w", err)
	}
	contents, readErr := rootDirectory.ReadFile(relative)
	closeErr := rootDirectory.Close()
	if readErr != nil {
		return nil, errors.Join(fmt.Errorf("read contained file %q: %w", relative, readErr), closeErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close template root: %w", closeErr)
	}
	return contents, nil
}

func canonicalDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve template root: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return "", fmt.Errorf("inspect template root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", &security.Refusal{
			Policy: "template-root", Path: absolute,
			Reason: "non-symlink directory required",
		}
	}
	return filepath.Clean(absolute), nil
}

func decode(contents []byte) (document, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var parsed document
	if err := decoder.Decode(&parsed); err != nil {
		return document{}, fmt.Errorf("parse template manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return document{}, errors.New("template manifest must contain exactly one document")
		}
		return document{}, fmt.Errorf("parse trailing template document: %w", err)
	}
	return parsed, nil
}

func validateDocument(root string, document document) error {
	if document.APIVersion != "ainfra.projectious.work/v1" || document.Kind != "Template" {
		return errors.New("unsupported template manifest contract")
	}
	if !namePattern.MatchString(document.Metadata.Name) ||
		filepath.Base(root) != document.Metadata.Name {
		return fmt.Errorf("template name %q must match directory basename", document.Metadata.Name)
	}
	if !versionPattern.MatchString(document.Metadata.Version) {
		return fmt.Errorf("invalid template version %q", document.Metadata.Version)
	}
	if strings.TrimSpace(document.Spec.Engines.Tofu.Directory) == "" ||
		strings.TrimSpace(document.Spec.Engines.Tofu.Version) == "" {
		return errors.New("opentofu directory and version are required")
	}
	hasAnsible := document.Spec.Engines.Ansible != nil
	if document.Spec.Outputs.Inventory == "none" && hasAnsible {
		return errors.New("ansible must be omitted when inventory is none")
	}
	if document.Spec.Outputs.Inventory != "none" && !hasAnsible {
		return errors.New("ansible is required when inventory names an output")
	}
	if document.Spec.Outputs.Inventory == "" {
		return errors.New("inventory declaration is required")
	}
	return nil
}

func requiredDirectory(root, relative string) (string, error) {
	if err := requireRelative(relative); err != nil {
		return "", err
	}
	path, err := security.ResolveContained(root, relative)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", &security.Refusal{Policy: "file-type", Path: path, Reason: "directory required"}
	}
	return path, nil
}

func requiredFile(root, relative string) (string, error) {
	if err := requireRelative(relative); err != nil {
		return "", err
	}
	path, err := security.ResolveContained(root, relative)
	if err != nil {
		return "", fmt.Errorf("resolve required template file %q: %w", relative, err)
	}
	if err := security.RequireRegular(path); err != nil {
		return "", fmt.Errorf("validate required template file %q: %w", relative, err)
	}
	return path, nil
}

func requireRelative(path string) error {
	if path == "" || filepath.IsAbs(path) {
		return &security.Refusal{
			Policy: "template-path", Path: path,
			Reason: "non-empty relative path required",
		}
	}
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return &security.Refusal{
			Policy: "template-path", Path: path,
			Reason: "path escapes template root",
		}
	}
	return nil
}

func rejectRuntimeState(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if name == ".ainfra" || name == ".terraform" ||
			strings.HasSuffix(name, ".tfstate") || strings.Contains(name, ".tfstate.") {
			return &security.Refusal{
				Policy: "template-state", Path: path,
				Reason: "runtime or infrastructure state is prohibited in template source",
			}
		}
		return nil
	})
}
