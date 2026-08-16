// Package mcpserver adapts typed ainfra application results to MCP.
package mcpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/projectious-work/ainfra/internal/app"
	"github.com/projectious-work/ainfra/internal/diagnostic"
	"github.com/projectious-work/ainfra/internal/output"
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
	return server
}

func boolPointer(value bool) *bool { return &value }

// ProtocolVersion identifies the MCP specification selected for Phase 7.
func ProtocolVersion() string { return protocolVersion }
