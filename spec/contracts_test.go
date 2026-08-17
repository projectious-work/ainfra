package contracts_test

import (
	"encoding/json"
	"testing"

	contracts "github.com/projectious-work/ainfra/spec"
)

func TestPublishedSchemaRegistryIsClosedAndValid(t *testing.T) {
	t.Parallel()
	names := contracts.Schemas()
	if len(names) != 10 {
		t.Fatalf("schema count = %d", len(names))
	}
	for _, name := range names {
		contents, ok := contracts.Schema(name)
		if !ok || !json.Valid(contents) {
			t.Fatalf("invalid embedded schema %q", name)
		}
	}
	if _, ok := contracts.Schema("../go.mod"); ok {
		t.Fatal("unregistered schema path succeeded")
	}
}
