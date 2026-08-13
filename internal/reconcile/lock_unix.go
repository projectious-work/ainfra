//go:build darwin || linux

package reconcile

import (
	"os"
	"syscall"
)

// FileLocker uses an exclusive advisory lock on the deployment manifest.
type FileLocker struct{}

// Lock acquires the operation lock and returns its idempotent release closure.
func (FileLocker) Lock(deploymentRoot, manifestName string) (func() error, error) {
	root, err := os.OpenRoot(deploymentRoot)
	if err != nil {
		return nil, err
	}
	file, err := root.Open(manifestName)
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		_ = root.Close()
		return nil, err
	}
	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		unlockErr := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		closeErr := file.Close()
		rootCloseErr := root.Close()
		if unlockErr != nil {
			return unlockErr
		}
		if closeErr != nil {
			return closeErr
		}
		return rootCloseErr
	}, nil
}
