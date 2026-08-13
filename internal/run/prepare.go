// Package run creates and binds private lifecycle run workspaces.
package run

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	lockfile "github.com/projectious-work/ainfra/internal/lock"
	"github.com/projectious-work/ainfra/internal/project"
	"github.com/projectious-work/ainfra/internal/security"
	"github.com/projectious-work/ainfra/internal/source"
)

// Binding identifies one immutable input to a reviewed plan.
type Binding struct {
	Path         string
	SnapshotPath string
	Digest       string
}

// Prepared is a private workspace and its complete pre-engine binding set.
type Prepared struct {
	ID               string
	Root             string
	Workspace        string
	Deployment       Binding
	NativeInputs     []Binding
	TemplateSource   string
	TemplateResolved string
	TemplateDigest   string
	ExecutablePath   string
	ExecutableDigest string
}

// Options contains trusted paths and identities selected by the application.
type Options struct {
	ID         string
	RunsRoot   string
	CacheRoot  string
	Deployment project.Deployment
	Lock       lockfile.Document
	Executable security.Executable
}

// Prepare creates an exclusive private run directory, verifies the lock-bound
// cache entry, copies it into workspace, and hashes all Phase 4 inputs.
func Prepare(options Options) (prepared Prepared, err error) {
	if !validID(options.ID) {
		return Prepared{}, errors.New("run ID must be a collision-resistant safe directory name")
	}
	if options.Lock.Template.Digest == "" || options.Lock.Template.Source == "" {
		return Prepared{}, errors.New("complete template lock required")
	}
	if err := options.Executable.VerifyUnchanged(); err != nil {
		return Prepared{}, fmt.Errorf("verify OpenTofu executable binding: %w", err)
	}
	if err := security.EnsurePrivateDir(options.RunsRoot); err != nil {
		return Prepared{}, fmt.Errorf("secure runs root: %w", err)
	}
	runRoot := filepath.Join(options.RunsRoot, options.ID)
	if err := os.Mkdir(runRoot, 0o700); err != nil {
		return Prepared{}, fmt.Errorf("create exclusive run directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(runRoot)
		}
	}()
	cachePath := filepath.Join(options.CacheRoot, "templates", "sha256", strings.TrimPrefix(options.Lock.Template.Digest, "sha256:"))
	workspace := filepath.Join(runRoot, "workspace")
	if err := source.CopyVerifiedTemplate(cachePath, workspace, options.Lock.Template.Digest); err != nil {
		return Prepared{}, fmt.Errorf("materialize run workspace: %w", err)
	}
	manifestDigest, err := digestFile(options.Deployment.Target.ManifestPath)
	if err != nil {
		return Prepared{}, fmt.Errorf("bind deployment manifest: %w", err)
	}
	inputPaths := append([]string(nil), options.Deployment.Inputs.TofuBackendConfigFiles...)
	inputPaths = append(inputPaths, options.Deployment.Inputs.TofuVariableFiles...)
	inputsRoot := filepath.Join(runRoot, "inputs")
	if len(inputPaths) > 0 {
		if err := os.Mkdir(inputsRoot, 0o700); err != nil {
			return Prepared{}, fmt.Errorf("create native input snapshot root: %w", err)
		}
	}
	nativeInputs := make([]Binding, 0, len(inputPaths))
	for _, relative := range inputPaths {
		path, resolveErr := security.ResolveContained(options.Deployment.Target.Root, relative)
		if resolveErr != nil {
			return Prepared{}, fmt.Errorf("bind native input %q: %w", relative, resolveErr)
		}
		digest, digestErr := digestFile(path)
		if digestErr != nil {
			return Prepared{}, fmt.Errorf("bind native input %q: %w", relative, digestErr)
		}
		snapshotRelative := filepath.Join("inputs", relative)
		snapshotPath := filepath.Join(runRoot, snapshotRelative)
		if mkdirErr := os.MkdirAll(filepath.Dir(snapshotPath), 0o700); mkdirErr != nil {
			return Prepared{}, fmt.Errorf("create native input snapshot parent: %w", mkdirErr)
		}
		if copyErr := copyBoundFile(path, runRoot, snapshotRelative, digest); copyErr != nil {
			return Prepared{}, fmt.Errorf("snapshot native input %q: %w", relative, copyErr)
		}
		nativeInputs = append(nativeInputs, Binding{Path: filepath.ToSlash(relative), SnapshotPath: filepath.ToSlash(snapshotRelative), Digest: digest})
	}
	prepared = Prepared{
		ID: options.ID, Root: runRoot, Workspace: workspace,
		Deployment: Binding{Path: "ainfra.yaml", Digest: manifestDigest}, NativeInputs: nativeInputs,
		TemplateSource: options.Lock.Template.Source, TemplateResolved: options.Lock.Template.Resolved,
		TemplateDigest: options.Lock.Template.Digest, ExecutablePath: options.Executable.Path(),
		ExecutableDigest: options.Executable.Digest(),
	}
	cleanup = false
	return prepared, nil
}

// Discard removes an unpublished or failed run after validating its identity.
func Discard(prepared Prepared) error {
	if !validID(prepared.ID) || filepath.Base(prepared.Root) != prepared.ID {
		return errors.New("refuse to discard invalid run path")
	}
	return os.RemoveAll(prepared.Root)
}

func copyBoundFile(sourcePath, destinationRoot, destinationRelative, expectedDigest string) error {
	sourceFile, err := os.Open(sourcePath) // #nosec G304 -- source passed contained regular-file policy.
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(destinationRoot)
	if err != nil {
		_ = sourceFile.Close()
		return err
	}
	destinationFile, err := root.OpenFile(destinationRelative, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	rootCloseErr := root.Close()
	if err != nil {
		_ = sourceFile.Close()
		return errors.Join(err, rootCloseErr)
	}
	if rootCloseErr != nil {
		return errors.Join(rootCloseErr, destinationFile.Close(), sourceFile.Close())
	}
	_, copyErr := io.Copy(destinationFile, sourceFile)
	err = errors.Join(copyErr, destinationFile.Sync(), destinationFile.Close(), sourceFile.Close())
	if err != nil {
		return err
	}
	observed, err := digestFile(filepath.Join(destinationRoot, destinationRelative))
	if err != nil || observed != expectedDigest {
		return errors.New("native input changed while snapshotting")
	}
	return nil
}

func validID(value string) bool {
	if len(value) < 16 || len(value) > 128 || value == "." || value == ".." {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._-", character) {
			continue
		}
		return false
	}
	return true
}

func digestFile(path string) (string, error) {
	if err := security.RequireRegular(path); err != nil {
		return "", err
	}
	file, err := os.Open(path) // #nosec G304 -- path passed contained regular-file policy.
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
