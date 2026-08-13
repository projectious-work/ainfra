package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/projectious-work/ainfra/internal/security"
)

// LocalResolution identifies a local source proven to be inside an approved
// deployment or repository root.
type LocalResolution struct {
	Path         string
	ApprovedRoot string
}

// ResolveLocal applies the v1 default local-source policy. A source may be
// contained by the deployment itself or by its nearest ancestor Git repository.
func ResolveLocal(reference Reference, deploymentRoot string) (LocalResolution, error) {
	if reference.Kind != KindLocal {
		return LocalResolution{}, errors.New("local source resolution requires a local reference")
	}
	canonicalDeployment, err := filepath.EvalSymlinks(deploymentRoot)
	if err != nil {
		return LocalResolution{}, fmt.Errorf("resolve deployment root: %w", err)
	}
	canonicalDeployment, err = filepath.Abs(canonicalDeployment)
	if err != nil {
		return LocalResolution{}, fmt.Errorf("resolve deployment root: %w", err)
	}
	candidate := filepath.Join(canonicalDeployment, filepath.FromSlash(reference.LocalPath))
	roots := []string{canonicalDeployment}
	if repositoryRoot, found, findErr := findRepositoryRoot(canonicalDeployment); findErr != nil {
		return LocalResolution{}, findErr
	} else if found && repositoryRoot != canonicalDeployment {
		roots = append(roots, repositoryRoot)
	}
	for _, root := range roots {
		resolved, resolveErr := security.ResolveContained(root, candidate)
		if resolveErr != nil {
			continue
		}
		info, statErr := os.Lstat(resolved)
		if statErr != nil {
			return LocalResolution{}, fmt.Errorf("inspect local template source: %w", statErr)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return LocalResolution{}, &security.Refusal{
				Policy: "local-source", Path: resolved,
				Reason: "non-symlink directory required",
			}
		}
		return LocalResolution{Path: resolved, ApprovedRoot: root}, nil
	}
	return LocalResolution{}, &security.Refusal{
		Policy: "local-source", Path: reference.LocalPath,
		Reason: "source is outside the deployment root and its parent repository",
	}
}

func findRepositoryRoot(start string) (string, bool, error) {
	for current := start; ; current = filepath.Dir(current) {
		marker := filepath.Join(current, ".git")
		if info, err := os.Lstat(marker); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
				return "", false, &security.Refusal{
					Policy: "repository-root", Path: marker,
					Reason: "Git marker must be a regular file or directory",
				}
			}
			return current, true, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, fmt.Errorf("inspect repository root: %w", err)
		}
		parent := filepath.Dir(current)
		if parent == current || strings.TrimSpace(parent) == "" {
			return "", false, nil
		}
	}
}
