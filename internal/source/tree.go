package source

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const treeDigestDomain = "ainfra-template-tree-v1"

// TreeDigest computes the normative ainfra v1 template-tree digest. The root
// must be an existing real directory. Git metadata is excluded; runtime state,
// symbolic links, and special files are rejected.
func TreeDigest(root string) (string, error) {
	return treeDigest(root, "")
}

// EngineWorkspaceDigest hashes template-controlled files while excluding only
// OpenTofu's declared initialization artifacts.
func EngineWorkspaceDigest(root, templateRoot string) (string, error) {
	return treeDigest(root, templateRoot)
}

func treeDigest(root, engineTemplateRoot string) (string, error) {
	isEngineWorkspace := engineTemplateRoot != ""
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return "", fmt.Errorf("inspect template root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return "", errors.New("template root must be a real directory")
	}
	canonicalRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve template root: %w", err)
	}

	files := make([]treeFile, 0)
	err = filepath.WalkDir(canonicalRoot, func(candidate string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if candidate == canonicalRoot {
			return nil
		}
		relative, err := filepath.Rel(canonicalRoot, candidate)
		if err != nil {
			return fmt.Errorf("relativize template path: %w", err)
		}
		normalizedPath := filepath.ToSlash(relative)
		if err := validateDigestPath(normalizedPath); err != nil {
			return err
		}
		if firstPathSegment(normalizedPath) == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if isEngineWorkspace && hasPathSegment(normalizedPath, ".terraform") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if isEngineWorkspace && filepath.Base(normalizedPath) == ".terraform.lock.hcl" {
			if _, statErr := os.Stat(filepath.Join(engineTemplateRoot, relative)); errors.Is(statErr, os.ErrNotExist) {
				return nil
			} else if statErr != nil {
				return fmt.Errorf("inspect template lock file %q: %w", normalizedPath, statErr)
			}
		}
		if isRuntimeArtifact(normalizedPath) {
			return fmt.Errorf("prohibited runtime artifact in template tree: %q", normalizedPath)
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect template path %q: %w", normalizedPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic link in template tree: %q", normalizedPath)
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("special file in template tree: %q", normalizedPath)
		}
		if fileLinkCount(info) > 1 {
			return fmt.Errorf("hard-linked file in template tree: %q", normalizedPath)
		}
		if info.Size() < 0 {
			return fmt.Errorf("negative file size in template tree: %q", normalizedPath)
		}
		files = append(files, treeFile{
			path: normalizedPath, nativePath: candidate,
			executable: info.Mode().Perm()&0o111 != 0, size: info.Size(),
		})
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(files, func(left, right int) bool {
		return files[left].path < files[right].path
	})

	digest := sha256.New()
	_, _ = io.WriteString(digest, treeDigestDomain)
	_, _ = digest.Write([]byte{0})
	var integer [8]byte
	for _, file := range files {
		binary.BigEndian.PutUint64(integer[:], uint64(len([]byte(file.path))))
		_, _ = digest.Write(integer[:])
		_, _ = io.WriteString(digest, file.path)
		if file.executable {
			_, _ = digest.Write([]byte{1})
		} else {
			_, _ = digest.Write([]byte{0})
		}
		if err := binary.Write(digest, binary.BigEndian, file.size); err != nil {
			return "", fmt.Errorf("frame template file size: %w", err)
		}
		if err := hashFile(digest, file); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("sha256:%x", digest.Sum(nil)), nil
}

func hasPathSegment(value, wanted string) bool {
	for _, segment := range strings.Split(value, "/") {
		if segment == wanted {
			return true
		}
	}
	return false
}

type treeFile struct {
	path       string
	nativePath string
	executable bool
	size       int64
}

func validateDigestPath(value string) error {
	if value == "" || value == "." || strings.HasPrefix(value, "/") ||
		!utf8.ValidString(value) || !norm.NFC.IsNormalString(value) {
		return fmt.Errorf("invalid non-NFC template path: %q", value)
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid template path segment in %q", value)
		}
	}
	return nil
}

func firstPathSegment(value string) string {
	if index := strings.IndexByte(value, '/'); index >= 0 {
		return value[:index]
	}
	return value
}

func isRuntimeArtifact(value string) bool {
	for _, segment := range strings.Split(value, "/") {
		if segment == ".ainfra" || segment == ".terraform" ||
			strings.HasSuffix(segment, ".tfstate") || strings.Contains(segment, ".tfstate.") {
			return true
		}
	}
	return false
}

func hashFile(destination io.Writer, file treeFile) error {
	stream, err := os.Open(file.nativePath)
	if err != nil {
		return fmt.Errorf("open template file %q: %w", file.path, err)
	}
	written, copyErr := io.Copy(destination, stream)
	closeErr := stream.Close()
	if copyErr != nil {
		return fmt.Errorf("hash template file %q: %w", file.path, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close template file %q: %w", file.path, closeErr)
	}
	if written != file.size {
		return fmt.Errorf("template file changed while hashing: %q", file.path)
	}
	return nil
}
