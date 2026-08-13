// Package initialize creates the minimal local deployment contract.
package initialize

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	manifestName = "ainfra.yaml"
	ignoreName   = ".gitignore"
	ignoreBlock  = "# ainfra BEGIN\n.ainfra/\n# ainfra END\n"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

// Result identifies the initialized deployment and every created path.
type Result struct {
	Name         string
	Root         string
	CreatedPaths []string
}

// Create initializes a minimal deployment after detecting every known
// conflict. Existing ainfra-owned files are never overwritten.
func Create(path string) (Result, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return Result{}, fmt.Errorf("resolve deployment path: %w", err)
	}
	name := filepath.Base(filepath.Clean(root))
	if !namePattern.MatchString(name) {
		return Result{}, fmt.Errorf("deployment directory name %q is invalid", name)
	}
	if err := inspectTarget(root); err != nil {
		return Result{}, err
	}
	created := []string{}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return Result{}, fmt.Errorf("create deployment directory: %w", err)
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return Result{}, fmt.Errorf("open deployment root: %w", err)
	}
	defer func() { _ = rootHandle.Close() }()
	manifest := fmt.Sprintf(`apiVersion: ainfra.projectious.work/v1
kind: Deployment
metadata:
  name: %s
spec:
  template:
    source: local:../template
`, name)
	if err := writeExclusive(rootHandle, manifestName, []byte(manifest), 0o644); err != nil {
		return Result{}, err
	}
	created = append(created, filepath.Join(root, manifestName))
	ignoreCreated, err := updateIgnore(rootHandle)
	if err != nil {
		return Result{}, err
	}
	if ignoreCreated {
		created = append(created, filepath.Join(root, ignoreName))
	}
	return Result{Name: name, Root: root, CreatedPaths: created}, nil
}

func inspectTarget(root string) error {
	information, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect deployment target: %w", err)
	}
	if information.Mode()&os.ModeSymlink != 0 || !information.IsDir() {
		return errors.New("deployment target must be a non-symlink directory")
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("open deployment target: %w", err)
	}
	defer func() { _ = rootHandle.Close() }()
	if _, err := rootHandle.Lstat(manifestName); err == nil {
		return errors.New("deployment manifest already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect deployment manifest: %w", err)
	}
	contents, err := rootHandle.ReadFile(ignoreName)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect .gitignore: %w", err)
	}
	begin := strings.Contains(string(contents), "# ainfra BEGIN")
	end := strings.Contains(string(contents), "# ainfra END")
	if begin != end {
		return errors.New("existing .gitignore contains an incomplete ainfra block")
	}
	return nil
}

func writeExclusive(root *os.Root, name string, contents []byte, mode os.FileMode) error {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", name, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", name, err)
	}
	return nil
}

func updateIgnore(root *os.Root) (bool, error) {
	file, err := root.OpenFile(ignoreName, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return false, fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = file.Close() }()
	contents, err := io.ReadAll(file)
	if err != nil {
		return false, fmt.Errorf("read .gitignore: %w", err)
	}
	if strings.Contains(string(contents), ignoreBlock) {
		return false, nil
	}
	prefix := ""
	if len(contents) > 0 && contents[len(contents)-1] != '\n' {
		prefix = "\n"
	}
	if _, err := file.WriteAt([]byte(prefix+ignoreBlock), int64(len(contents))); err != nil {
		return false, fmt.Errorf("append .gitignore: %w", err)
	}
	return len(contents) == 0, nil
}
