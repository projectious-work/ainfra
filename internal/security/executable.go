package security

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Executable is a validated immutable child executable descriptor.
type Executable struct {
	path   string
	digest [sha256.Size]byte
}

// ResolveExecutable validates an absolute regular executable and fingerprints it.
func ResolveExecutable(path string) (Executable, error) {
	if !filepath.IsAbs(path) {
		return Executable{}, &Refusal{Policy: "executable", Path: path, Reason: "absolute path required"}
	}
	if err := RequireRegular(path); err != nil {
		return Executable{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Executable{}, fmt.Errorf("inspect executable: %w", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return Executable{}, &Refusal{Policy: "executable", Path: path, Reason: "file is not executable"}
	}
	// #nosec G304 -- the absolute path passed regular-file and symlink policy
	// immediately above; VerifyUnchanged repeats this check before execution.
	file, err := os.Open(path)
	if err != nil {
		return Executable{}, fmt.Errorf("open executable: %w", err)
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return Executable{}, fmt.Errorf("hash executable: %w", copyErr)
	}
	if closeErr != nil {
		return Executable{}, fmt.Errorf("close executable: %w", closeErr)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return Executable{}, fmt.Errorf("canonicalize executable: %w", err)
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return Executable{path: canonical, digest: digest}, nil
}

// Path returns the canonical executable path.
func (executable Executable) Path() string {
	return executable.path
}

// VerifyUnchanged refuses an executable whose bytes changed after resolution.
func (executable Executable) VerifyUnchanged() error {
	current, err := ResolveExecutable(executable.path)
	if err != nil {
		return err
	}
	if current.digest != executable.digest {
		return &Refusal{Policy: "executable", Path: executable.path, Reason: "binding changed"}
	}
	return nil
}
