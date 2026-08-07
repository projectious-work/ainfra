// Package security enforces containment, file, environment, and output policy.
package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Refusal reports a security-policy failure safe for user presentation.
type Refusal struct {
	Policy string
	Path   string
	Reason string
}

// Error returns a safe description without including file contents.
func (refusal *Refusal) Error() string {
	if refusal.Path == "" {
		return fmt.Sprintf("%s policy refusal: %s", refusal.Policy, refusal.Reason)
	}
	return fmt.Sprintf("%s policy refusal for %q: %s", refusal.Policy, refusal.Path, refusal.Reason)
}

// ResolveContained resolves an existing candidate and proves it remains in root.
func ResolveContained(root, candidate string) (string, error) {
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", &Refusal{Policy: "path", Path: root, Reason: "resolve root"}
	}
	canonicalRoot, err = filepath.Abs(canonicalRoot)
	if err != nil {
		return "", fmt.Errorf("absolute root: %w", err)
	}

	target := candidate
	if !filepath.IsAbs(target) {
		target = filepath.Join(canonicalRoot, target)
	}
	if err := rejectSymlinkComponents(canonicalRoot, target); err != nil {
		return "", err
	}
	canonicalTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", &Refusal{Policy: "path", Path: candidate, Reason: "resolve target"}
	}
	canonicalTarget, err = filepath.Abs(canonicalTarget)
	if err != nil {
		return "", fmt.Errorf("absolute target: %w", err)
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalTarget)
	if err != nil {
		return "", &Refusal{Policy: "containment", Path: candidate, Reason: "compare root"}
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", &Refusal{Policy: "containment", Path: candidate, Reason: "outside allowed root"}
	}
	return canonicalTarget, nil
}

func rejectSymlinkComponents(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return &Refusal{Policy: "containment", Path: target, Reason: "outside allowed root"}
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				return &Refusal{Policy: "path", Path: target, Reason: "target does not exist"}
			}
			return fmt.Errorf("inspect path component: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return &Refusal{Policy: "symlink", Path: current, Reason: "symbolic links are not allowed"}
		}
	}
	return nil
}
