package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	operational "github.com/projectious-work/ainfra/internal/logging"
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
	audit *requestAudit
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
	mcp.AddTool(server, tool, boundedToolHandler(limiter, tool.Name, handler))
}

func boundedToolHandler[In, Out any](limiter *requestLimiter,
	tool string, handler mcp.ToolHandlerFor[In, Out],
) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, request *mcp.CallToolRequest,
		input In,
	) (*mcp.CallToolResult, Out, error) {
		var zero Out
		if err := limiter.acquire(ctx); err != nil {
			return nil, zero, err
		}
		defer limiter.release()
		requestID, auditErr := limiter.audit.start(tool, mcpRunID(input))
		if auditErr != nil {
			return nil, zero, auditErr
		}
		result, output, handlerErr := handler(ctx, request, input)
		failed := handlerErr != nil || result != nil && result.IsError
		runID := mcpRunID(input)
		if runID == "" {
			runID = mcpResultRunID(output)
		}
		auditErr = limiter.audit.finish(tool, runID, requestID, failed)
		return result, output, errors.Join(handlerErr, auditErr)
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
		requestID, auditErr := limiter.audit.start(resource.URI, "")
		if auditErr != nil {
			return nil, auditErr
		}
		result, handlerErr := handler(ctx, request)
		auditErr = limiter.audit.finish(resource.URI, "", requestID, handlerErr != nil)
		return result, errors.Join(handlerErr, auditErr)
	})
}

type requestAudit struct {
	logger     *operational.Logger
	now        func() time.Time
	deployment string
	sequence   atomic.Uint64
}

func newRequestAudit(logger *operational.Logger, now func() time.Time,
	deployment string,
) *requestAudit {
	if now == nil {
		now = time.Now
	}
	return &requestAudit{logger: logger, now: now, deployment: deployment}
}

func (audit *requestAudit) start(tool, runID string) (string, error) {
	if audit == nil || audit.logger == nil {
		return "", nil
	}
	requestID := fmt.Sprintf("mcp-%016x", audit.sequence.Add(1))
	err := audit.logger.Write(operational.NewMCPEvent(audit.now(), "info",
		"request started", tool, audit.deployment, runID, requestID))
	return requestID, err
}

func (audit *requestAudit) finish(tool, runID, requestID string, failed bool) error {
	if audit == nil || audit.logger == nil {
		return nil
	}
	level, message := "info", "request finished"
	if failed {
		level, message = "error", "request failed"
	}
	return audit.logger.Write(operational.NewMCPEvent(audit.now(), level,
		message, tool, audit.deployment, runID, requestID))
}

func mcpRunID(input any) string {
	if retained, ok := input.(RetainedArtifactInput); ok {
		if safeCorrelationID(retained.RunID) {
			return retained.RunID
		}
	}
	return ""
}

func mcpResultRunID(result any) string {
	if plan, ok := result.(CreatePlanResult); ok && plan.Result != nil &&
		safeCorrelationID(plan.Result.RunID) {
		return plan.Result.RunID
	}
	return ""
}

func safeCorrelationID(value string) bool {
	if len(value) < 16 || len(value) > 128 || value == "." || value == ".." {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '.' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
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
