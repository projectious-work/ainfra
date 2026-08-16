package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	maxConcurrentRequests = 8
	maxStdioFrameBytes    = 4 << 20
)

var (
	errInvalidStdioFrame  = errors.New("MCP stdio frame is not valid JSON")
	errStdioFrameTooLarge = errors.New("MCP stdio frame exceeds size limit")
)

type requestLimiter struct {
	slots chan struct{}
}

func newRequestLimiter(limit int) *requestLimiter {
	return &requestLimiter{slots: make(chan struct{}, limit)}
}

func (limiter *requestLimiter) acquire(ctx context.Context) error {
	select {
	case limiter.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (limiter *requestLimiter) release() { <-limiter.slots }

func addBoundedTool[In, Out any](server *mcp.Server, limiter *requestLimiter,
	tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out],
) {
	mcp.AddTool(server, tool, boundedToolHandler(limiter, handler))
}

func boundedToolHandler[In, Out any](limiter *requestLimiter,
	handler mcp.ToolHandlerFor[In, Out],
) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, request *mcp.CallToolRequest,
		input In,
	) (*mcp.CallToolResult, Out, error) {
		var zero Out
		if err := limiter.acquire(ctx); err != nil {
			return nil, zero, err
		}
		defer limiter.release()
		return handler(ctx, request, input)
	}
}

func addBoundedResource(server *mcp.Server, limiter *requestLimiter,
	resource *mcp.Resource, handler mcp.ResourceHandler,
) {
	server.AddResource(resource, func(ctx context.Context,
		request *mcp.ReadResourceRequest,
	) (*mcp.ReadResourceResult, error) {
		if err := limiter.acquire(ctx); err != nil {
			return nil, err
		}
		defer limiter.release()
		return handler(ctx, request)
	})
}

type frameLimitReadCloser struct {
	source  io.ReadCloser
	reader  *bufio.Reader
	limit   int
	pending []byte
}

func newFrameLimitReadCloser(reader io.ReadCloser, limit int) io.ReadCloser {
	return &frameLimitReadCloser{source: reader,
		reader: bufio.NewReaderSize(reader, limit+1), limit: limit}
}

func (reader *frameLimitReadCloser) Read(buffer []byte) (int, error) {
	if len(reader.pending) == 0 {
		frame, err := reader.reader.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			return 0, errStdioFrameTooLarge
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}
		if len(frame) == 0 {
			return 0, io.EOF
		}
		payload := bytes.TrimSuffix(frame, []byte{'\n'})
		payload = bytes.TrimSuffix(payload, []byte{'\r'})
		if len(payload) > reader.limit {
			return 0, errStdioFrameTooLarge
		}
		if !json.Valid(payload) {
			return 0, errInvalidStdioFrame
		}
		reader.pending = frame
	}
	n := copy(buffer, reader.pending)
	reader.pending = reader.pending[n:]
	return n, nil
}

func (reader *frameLimitReadCloser) Close() error { return reader.source.Close() }
