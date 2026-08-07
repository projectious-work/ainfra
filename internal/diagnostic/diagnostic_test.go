package diagnostic_test

import (
	"testing"

	"github.com/projectious-work/ainfra/internal/diagnostic"
)

func TestValidCode(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"AINFRA-E0001": true,
		"AINFRA-E9999": true,
		"AINFRA-E001":  false,
		"ainfra-E0001": false,
		"AINFRA-W0001": false,
	}
	for code, expected := range tests {
		if actual := diagnostic.ValidCode(code); actual != expected {
			t.Errorf("ValidCode(%q) = %v, want %v", code, actual, expected)
		}
	}
}
