package contracts_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPProtocolDependencyIsConfinedToAdapter(t *testing.T) {
	t.Parallel()
	internalRoot := filepath.Join("..", "internal")
	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == filepath.Join(internalRoot, "mcpserver") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(contents), "github.com/modelcontextprotocol/go-sdk") {
			t.Errorf("MCP protocol dependency escaped adapter: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
