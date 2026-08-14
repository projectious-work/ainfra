package run

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/projectious-work/ainfra/internal/reconcile"
)

func TestConfigurationStartIsExclusiveUnderDeploymentLock(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ainfra.yaml"), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	reviewed := Reviewed{Root: root}
	if err := AppendExecutionEvent(reviewed, "started", time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	if err := AppendExecutionEvent(reviewed, "succeeded", time.Now(), nil); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			ready.Wait()
			unlock, err := (reconcile.FileLocker{}).Lock(root, "ainfra.yaml")
			if err == nil {
				err = BeginOperation(reviewed, "configure", time.Now())
				_ = unlock()
			}
			results <- err
		}()
	}
	succeeded, refused := 0, 0
	for range 2 {
		if err := <-results; err == nil {
			succeeded++
		} else {
			refused++
		}
	}
	if succeeded != 1 || refused != 1 {
		t.Fatalf("succeeded=%d refused=%d", succeeded, refused)
	}
}
