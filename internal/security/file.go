package security

import (
	"errors"
	"fmt"
	"os"
)

// RequireRegular verifies that path names a non-symlink regular file.
func RequireRegular(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect regular file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return &Refusal{Policy: "symlink", Path: path, Reason: "symbolic links are not allowed"}
	}
	if !info.Mode().IsRegular() {
		return &Refusal{Policy: "file-type", Path: path, Reason: "regular file required"}
	}
	return nil
}

// EnsurePrivateDir creates a directory and enforces owner-only permissions.
func EnsurePrivateDir(path string) error {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create private directory: %w", err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect private directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return &Refusal{Policy: "file-type", Path: path, Reason: "private directory required"}
	}
	if info.Mode().Perm()&0o077 != 0 {
		return &Refusal{Policy: "permissions", Path: path, Reason: "group or other access is prohibited"}
	}
	return nil
}

// CreatePrivateFile exclusively creates a new owner-only regular file in root.
// The caller owns the returned file and must close it.
func CreatePrivateFile(root, relative string) (*os.File, error) {
	privateRoot, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open private root: %w", err)
	}
	file, err := privateRoot.OpenFile(relative, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	rootCloseErr := privateRoot.Close()
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, errors.Join(&Refusal{Policy: "file", Path: relative, Reason: "refusing to overwrite existing path"}, rootCloseErr)
		}
		return nil, errors.Join(fmt.Errorf("create private file: %w", err), rootCloseErr)
	}
	if rootCloseErr != nil {
		fileCloseErr := file.Close()
		return nil, errors.Join(fmt.Errorf("close private root: %w", rootCloseErr), fileCloseErr)
	}
	info, statErr := file.Stat()
	if statErr != nil {
		closeErr := file.Close()
		return nil, errors.Join(fmt.Errorf("inspect private file: %w", statErr), closeErr)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		closeErr := file.Close()
		return nil, errors.Join(&Refusal{Policy: "permissions", Path: relative, Reason: "owner-only regular file required"}, closeErr)
	}
	return file, nil
}
