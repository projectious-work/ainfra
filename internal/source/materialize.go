package source

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Materialized identifies one verified immutable local cache entry.
type Materialized struct {
	Path   string
	Digest string
	Reused bool
}

// MaterializeLocal copies a validated local template tree into a private,
// digest-addressed cache entry. The source root and cache root are explicit
// policy outputs; this function never consults ambient configuration.
func MaterializeLocal(sourceRoot, cacheRoot string) (Materialized, error) {
	if err := requireSeparateRoots(sourceRoot, cacheRoot); err != nil {
		return Materialized{}, err
	}
	digest, err := TreeDigest(sourceRoot)
	if err != nil {
		return Materialized{}, fmt.Errorf("digest local template source: %w", err)
	}
	key := strings.TrimPrefix(digest, "sha256:")
	entriesRoot := filepath.Join(cacheRoot, "templates", "sha256")
	if err := os.MkdirAll(entriesRoot, 0o700); err != nil {
		return Materialized{}, fmt.Errorf("create template cache root: %w", err)
	}
	if err := os.Chmod(entriesRoot, 0o700); err != nil { // #nosec G302 -- owner-only directory requires execute permission.
		return Materialized{}, fmt.Errorf("secure template cache root: %w", err)
	}
	destination := filepath.Join(entriesRoot, key)
	if info, statErr := os.Lstat(destination); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return Materialized{}, errors.New("template cache entry is not a real directory")
		}
		observed, digestErr := TreeDigest(destination)
		if digestErr != nil || observed != digest {
			return Materialized{}, errors.New("template cache entry failed digest verification")
		}
		return Materialized{Path: destination, Digest: digest, Reused: true}, nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return Materialized{}, fmt.Errorf("inspect template cache entry: %w", statErr)
	}

	staging, err := os.MkdirTemp(entriesRoot, ".staging-")
	if err != nil {
		return Materialized{}, fmt.Errorf("create template cache staging directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(staging)
		}
	}()
	if err := os.Chmod(staging, 0o700); err != nil { // #nosec G302 -- owner-only directory requires execute permission.
		return Materialized{}, fmt.Errorf("secure template cache staging directory: %w", err)
	}
	if err := copyTemplateTree(sourceRoot, staging); err != nil {
		return Materialized{}, err
	}
	observed, err := TreeDigest(staging)
	if err != nil || observed != digest {
		return Materialized{}, errors.New("materialized template digest does not match source")
	}
	if err := os.Rename(staging, destination); err != nil {
		if _, statErr := os.Stat(destination); statErr == nil {
			observed, digestErr := TreeDigest(destination)
			if digestErr == nil && observed == digest {
				return Materialized{Path: destination, Digest: digest, Reused: true}, nil
			}
		}
		return Materialized{}, fmt.Errorf("publish template cache entry: %w", err)
	}
	cleanup = false
	return Materialized{Path: destination, Digest: digest}, nil
}

func requireSeparateRoots(sourceRoot, cacheRoot string) error {
	canonicalSource, err := filepath.Abs(sourceRoot)
	if err != nil {
		return fmt.Errorf("resolve local source root: %w", err)
	}
	canonicalCache, err := filepath.Abs(cacheRoot)
	if err != nil {
		return fmt.Errorf("resolve template cache root: %w", err)
	}
	if pathContains(canonicalSource, canonicalCache) || pathContains(canonicalCache, canonicalSource) {
		return errors.New("local source and template cache roots must not overlap")
	}
	return nil
}

func pathContains(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func copyTemplateTree(sourceRoot, destinationRoot string) error {
	canonicalSource, err := filepath.Abs(sourceRoot)
	if err != nil {
		return fmt.Errorf("resolve local source root: %w", err)
	}
	sourceDirectory, err := os.OpenRoot(canonicalSource)
	if err != nil {
		return fmt.Errorf("open local source root: %w", err)
	}
	defer func() { _ = sourceDirectory.Close() }()
	destinationDirectory, err := os.OpenRoot(destinationRoot)
	if err != nil {
		return fmt.Errorf("open template staging root: %w", err)
	}
	defer func() { _ = destinationDirectory.Close() }()
	return filepath.WalkDir(canonicalSource, func(candidate string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if candidate == canonicalSource {
			return nil
		}
		relative, err := filepath.Rel(canonicalSource, candidate)
		if err != nil {
			return fmt.Errorf("relativize local source path: %w", err)
		}
		if firstPathSegment(filepath.ToSlash(relative)) == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect local source path: %w", err)
		}
		if info.IsDir() {
			return destinationDirectory.Mkdir(relative, 0o700)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || fileLinkCount(info) > 1 {
			return fmt.Errorf("unsafe local source entry: %q", filepath.ToSlash(relative))
		}
		mode := os.FileMode(0o600)
		if info.Mode().Perm()&0o111 != 0 {
			mode = 0o700
		}
		return copyRegularFile(sourceDirectory, destinationDirectory, relative, mode, info.Size())
	})
}

func copyRegularFile(
	sourceRoot, destinationRoot *os.Root,
	relative string,
	mode os.FileMode,
	expectedSize int64,
) error {
	sourceFile, err := sourceRoot.Open(relative)
	if err != nil {
		return fmt.Errorf("open local source file: %w", err)
	}
	destinationFile, err := destinationRoot.OpenFile(
		relative, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode,
	)
	if err != nil {
		_ = sourceFile.Close()
		return fmt.Errorf("create materialized file: %w", err)
	}
	written, copyErr := io.Copy(destinationFile, sourceFile)
	closeDestinationErr := destinationFile.Close()
	closeSourceErr := sourceFile.Close()
	if copyErr != nil || closeDestinationErr != nil || closeSourceErr != nil || written != expectedSize {
		return errors.New("local source changed or failed while materializing")
	}
	return nil
}
