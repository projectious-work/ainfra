package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/output"
)

func TestEnvelopeConstructorsEnforceSchemaInvariants(t *testing.T) {
	tests := []struct {
		name string
		call func()
	}{
		{name: "nil success", call: func() { output.Success(output.CommandVersion, nil) }},
		{name: "empty failure", call: func() { output.Failure(output.CommandInvocation) }},
		{name: "nil partial failure", call: func() {
			output.PartialFailure(output.CommandDoctorEnvironment, nil, diagnostic.Diagnostic{})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("constructor accepted an invalid envelope")
				}
			}()
			test.call()
		})
	}
}

func versionEnvelope() output.Envelope {
	return output.Success(output.CommandVersion, output.Version{
		Version:   "1.0.0-alpha.1",
		Commit:    "0123456789abcdef",
		BuiltAt:   "2026-08-07T00:00:00Z",
		GoVersion: "go1.26.5",
		Platform:  "linux/arm64",
		SupportedContractVersions: output.SupportedContractVersions{
			DocumentAPIVersions:          []string{"ainfra.projectious.work/v1"},
			ResultAPIVersions:            []string{"ainfra.result/v1"},
			StandardOutputSchemaVersions: []string{"1"},
		},
	})
}

func TestRenderJSON(t *testing.T) {
	t.Parallel()
	var rendered bytes.Buffer
	err := output.Render(&rendered, versionEnvelope(), output.RenderOptions{
		Format: output.FormatJSON,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(rendered.Bytes(), &envelope); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if envelope["apiVersion"] != output.APIVersion {
		t.Errorf("apiVersion = %v", envelope["apiVersion"])
	}
	if strings.Contains(rendered.String(), "\x1b[") {
		t.Error("JSON contains an ANSI escape")
	}
}

func TestRenderPlainVersion(t *testing.T) {
	t.Parallel()
	var rendered bytes.Buffer
	err := output.Render(&rendered, versionEnvelope(), output.RenderOptions{
		Format: output.FormatText,
		Style:  output.StylePlain,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := "ainfra 1.0.0-alpha.1\ncommit 0123456789abcdef\nbuilt 2026-08-07T00:00:00Z\ngo go1.26.5\nplatform linux/arm64\ncontracts documents ainfra.projectious.work/v1; results ainfra.result/v1; standard output 1\n"
	if rendered.String() != want {
		t.Errorf("plain output:\n%s\nwant:\n%s", rendered.String(), want)
	}
}
