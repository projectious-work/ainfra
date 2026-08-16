package app

// MCPServeRequest contains server-start policy selected by the operator.
// Protocol-specific types belong to the MCP adapter and must not enter app.
type MCPServeRequest struct {
	ProjectPath  string
	Capabilities []string
}
