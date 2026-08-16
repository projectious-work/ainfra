// Package mcpserver adapts typed ainfra application results to MCP.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/output"
	contracts "github.com/projectious-work/ainfra/spec"
)

const protocolVersion = "2026-07-28"

// Options contains composition-root dependencies for one stdio server.
type Options struct {
	Build   app.Build
	Prepare func(app.MCPServeRequest) (app.MCPServeSession, error)
	Stdin   io.ReadCloser
	Stdout  io.WriteCloser
	Stderr  io.Writer
}

// VersionInput is the closed input contract for ainfra.version.
type VersionInput struct{}

// VersionResult is the versioned result contract for ainfra.version.
type VersionResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      output.Version          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// ProjectInput is the closed input contract for ainfra.project.inspect.
type ProjectInput struct{}

// ProjectResult is the versioned result contract for project inspection.
type ProjectResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      output.Deployment       `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// StatusInput is the closed input contract for ainfra.status. The project is
// deliberately absent because it is fixed at server startup.
type StatusInput struct{}

// StatusResult is the versioned result contract for retained status.
type StatusResult struct {
	APIVersion  string                  `json:"apiVersion"`
	Tool        string                  `json:"tool"`
	OK          bool                    `json:"ok"`
	Result      *output.Status          `json:"result"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// Serve runs one MCP session until stdin closes or the context is cancelled.
func Serve(ctx context.Context, request app.MCPServeRequest, options Options) error {
	if options.Prepare == nil {
		return errors.New("MCP project preparation is unavailable")
	}
	session, err := options.Prepare(request)
	if err != nil {
		return err
	}
	server := New(session, options)
	return server.Run(ctx, &mcp.IOTransport{Reader: options.Stdin, Writer: options.Stdout})
}

// New constructs the default-deny server registry.
func New(session app.MCPServeSession, options Options) *mcp.Server {
	logger := slog.New(slog.NewTextHandler(options.Stderr, &slog.HandlerOptions{}))
	server := mcp.NewServer(&mcp.Implementation{Name: "ainfra", Version: options.Build.Version},
		&mcp.ServerOptions{Instructions: "Read-only ainfra tools are exposed by default.", Logger: logger})
	mcp.AddTool(server, &mcp.Tool{Name: "ainfra.version",
		Description: "Return the ainfra build and supported contract versions.",
		Annotations: &mcp.ToolAnnotations{Title: "ainfra version", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ VersionInput) (*mcp.CallToolResult,
			VersionResult, error) {
			return nil, VersionResult{APIVersion: output.APIVersion, Tool: "ainfra.version",
				OK: true, Result: app.Version(options.Build), Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "ainfra.project.inspect",
		Description: "Return the immutable project identity selected at server startup.",
		Annotations: &mcp.ToolAnnotations{Title: "inspect ainfra project", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ ProjectInput) (*mcp.CallToolResult,
			ProjectResult, error) {
			return nil, ProjectResult{APIVersion: output.APIVersion,
				Tool: "ainfra.project.inspect", OK: true, Result: session.Project,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "ainfra.status",
		Description: "Return sanitized retained lifecycle status for the startup project.",
		Annotations: &mcp.ToolAnnotations{Title: "ainfra retained status", ReadOnlyHint: true,
			IdempotentHint: true, OpenWorldHint: boolPointer(false)}},
		func(_ context.Context, _ *mcp.CallToolRequest, _ StatusInput) (*mcp.CallToolResult,
			StatusResult, error) {
			status, err := session.Status()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, StatusResult{
					APIVersion: output.APIVersion, Tool: "ainfra.status", OK: false,
					Diagnostics: []diagnostic.Diagnostic{{Code: "AINFRA-E4601",
						Severity: diagnostic.SeverityError, Message: err.Error(), Component: "status",
						NextAction: "Inspect retained run evidence and deployment configuration."}},
				}, nil
			}
			return nil, StatusResult{APIVersion: output.APIVersion,
				Tool: "ainfra.status", OK: true, Result: &status,
				Diagnostics: []diagnostic.Diagnostic{}}, nil
		})
	addContractResources(server)
	return server
}

func addContractResources(server *mcp.Server) {
	for _, name := range contracts.Schemas() {
		name := name
		contents, ok := contracts.Schema(name)
		if !ok {
			panic("registered schema is not embedded: " + name)
		}
		uri := "ainfra://schemas/v1/" + name
		server.AddResource(&mcp.Resource{Name: name, Title: "ainfra v1 schema: " + name,
			Description: "Published, immutable ainfra v1 JSON Schema.",
			MIMEType:    "application/schema+json", URI: uri, Size: int64(len(contents))},
			func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
					URI: uri, MIMEType: "application/schema+json", Text: string(contents),
				}}}, nil
			})
	}
	catalog := contractCatalog()
	const catalogURI = "ainfra://contracts/v1"
	server.AddResource(&mcp.Resource{Name: "ainfra-contracts-v1",
		Title:       "ainfra v1 contracts and documentation",
		Description: "Supported contract versions, schema resources, and documentation references.",
		MIMEType:    "application/json", URI: catalogURI, Size: int64(len(catalog))},
		func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
				URI: catalogURI, MIMEType: "application/json", Text: catalog,
			}}}, nil
		})
}

func contractCatalog() string {
	type reference struct {
		Name string `json:"name"`
		URI  string `json:"uri"`
	}
	schemaNames := contracts.Schemas()
	schemas := make([]reference, len(schemaNames))
	for index, name := range schemaNames {
		schemas[index] = reference{Name: name, URI: "ainfra://schemas/v1/" + name}
	}
	value := struct {
		APIVersion                string                           `json:"apiVersion"`
		SupportedContractVersions output.SupportedContractVersions `json:"supportedContractVersions"`
		Schemas                   []reference                      `json:"schemas"`
		Documentation             []reference                      `json:"documentation"`
	}{APIVersion: "ainfra.contracts/v1",
		SupportedContractVersions: app.Version(app.Build{}).SupportedContractVersions,
		Schemas:                   schemas,
		Documentation: []reference{
			{Name: "documentation", URI: "https://projectious-work.github.io/ainfra/docs/"},
			{Name: "roadmap", URI: "https://projectious-work.github.io/ainfra/docs/roadmap/"},
			{Name: "v1 specification", URI: "https://github.com/projectious-work/ainfra/tree/v1.x-dev/spec/doc/v1"},
		},
	}
	contents, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

func boolPointer(value bool) *bool { return &value }

// ProtocolVersion identifies the MCP specification selected for Phase 7.
func ProtocolVersion() string { return protocolVersion }
