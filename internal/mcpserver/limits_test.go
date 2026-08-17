package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	operational "github.com/projectious-work/ainfra/internal/logging"
	"github.com/projectious-work/ainfra/internal/output"
)

func TestBoundedToolHandlerLimitsConcurrentRequests(t *testing.T) {
	t.Parallel()
	limiter := newRequestLimiter(3)
	started := make(chan struct{}, 12)
	release := make(chan struct{})
	var active, maximum atomic.Int32
	handler := boundedToolHandler(limiter, "test.block",
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult,
			struct{}, error) {
			current := active.Add(1)
			for previous := maximum.Load(); current > previous &&
				!maximum.CompareAndSwap(previous, current); previous = maximum.Load() {
			}
			started <- struct{}{}
			select {
			case <-release:
			case <-ctx.Done():
			}
			active.Add(-1)
			return nil, struct{}{}, ctx.Err()
		})
	var wait sync.WaitGroup
	for range 12 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _, _ = handler(context.Background(), nil, struct{}{})
		}()
	}
	for range 3 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("bounded handler did not start permitted requests")
		}
	}
	select {
	case <-started:
		t.Fatal("bounded handler exceeded concurrency limit")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	wait.Wait()
	if maximum.Load() != 3 {
		t.Fatalf("maximum concurrency = %d", maximum.Load())
	}
}

func TestBoundedToolHandlerHonorsCancellationWhileWaiting(t *testing.T) {
	t.Parallel()
	limiter := newRequestLimiter(1)
	if err := limiter.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer limiter.release()
	called := false
	handler := boundedToolHandler(limiter, "test.cancel",
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult,
			struct{}, error) {
			called = true
			return nil, struct{}{}, nil
		})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := handler(ctx, nil, struct{}{})
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("waiting cancellation: called=%t err=%v", called, err)
	}
}

func TestBoundedToolHandlerEmitsCorrelatedAuditEvents(t *testing.T) {
	t.Parallel()
	var destination bytes.Buffer
	logger := operational.Logger{Level: "info", Sinks: []operational.Sink{
		&operational.WriterSink{Writer: &destination, Format: "json"},
	}}
	limiter := newRequestLimiter(1)
	limiter.audit = newRequestAudit(&logger,
		func() time.Time { return time.Unix(0, 0) }, "example")
	handler := boundedToolHandler(limiter, "ainfra.output.read",
		func(context.Context, *mcp.CallToolRequest, RetainedArtifactInput) (*mcp.CallToolResult,
			struct{}, error) {
			return &mcp.CallToolResult{IsError: true}, struct{}{}, nil
		})
	const runID = "20260816T120000Z-0123456789abcdef"
	result, _, err := handler(context.Background(), nil, RetainedArtifactInput{RunID: runID})
	if err != nil || result == nil || !result.IsError {
		t.Fatalf("handler result=%+v err=%v", result, err)
	}
	decoder := json.NewDecoder(&destination)
	var started, finished operational.Event
	if err := decoder.Decode(&started); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&finished); err != nil {
		t.Fatal(err)
	}
	if started.Message != "request started" || finished.Message != "request failed" ||
		started.RequestID == "" || finished.RequestID != started.RequestID ||
		started.Command != "ainfra.output.read" || started.Deployment != "example" ||
		started.RunID != runID || finished.Level != "error" {
		t.Fatalf("unexpected audit events: start=%+v finish=%+v", started, finished)
	}
}

func TestMCPRunIDOmitsUnsafeAuditInput(t *testing.T) {
	t.Parallel()
	if value := mcpRunID(RetainedArtifactInput{RunID: "password=do-not-log"}); value != "" {
		t.Fatalf("unsafe run correlation was retained: %q", value)
	}
}

func TestBoundedToolHandlerCorrelatesCreatedPlanOnFinish(t *testing.T) {
	t.Parallel()
	var destination bytes.Buffer
	logger := operational.Logger{Level: "info", Sinks: []operational.Sink{
		&operational.WriterSink{Writer: &destination, Format: "json"},
	}}
	limiter := newRequestLimiter(1)
	limiter.audit = newRequestAudit(&logger,
		func() time.Time { return time.Unix(0, 0) }, "example")
	const runID = "20260816T120000Z-0123456789abcdef"
	handler := boundedToolHandler(limiter, "ainfra.plan.create",
		func(context.Context, *mcp.CallToolRequest, CreatePlanInput) (*mcp.CallToolResult,
			CreatePlanResult, error) {
			return nil, CreatePlanResult{Result: &output.Plan{RunID: runID}}, nil
		})
	if _, _, err := handler(context.Background(), nil, CreatePlanInput{Intent: "apply"}); err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(&destination)
	var started, finished operational.Event
	if err := decoder.Decode(&started); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&finished); err != nil {
		t.Fatal(err)
	}
	if started.RunID != "" || finished.RunID != runID ||
		finished.RequestID != started.RequestID {
		t.Fatalf("unexpected plan audit events: start=%+v finish=%+v", started, finished)
	}
}

func TestFrameLimitReadCloser(t *testing.T) {
	t.Parallel()
	reader := newFrameLimitReadCloser(io.NopCloser(strings.NewReader("1234\n12")), 4)
	contents, err := io.ReadAll(reader)
	if err != nil || string(contents) != "1234\n12" {
		t.Fatalf("bounded frames: %q, %v", contents, err)
	}
	reader = newFrameLimitReadCloser(io.NopCloser(strings.NewReader("12345\n")), 4)
	contents, err = io.ReadAll(reader)
	if !errors.Is(err, errStdioFrameTooLarge) || len(contents) != 0 {
		t.Fatalf("oversized frame: %q, %v", contents, err)
	}
	reader = newFrameLimitReadCloser(io.NopCloser(strings.NewReader("{\n}\n")), 4)
	contents, err = io.ReadAll(reader)
	if !errors.Is(err, errInvalidStdioFrame) || len(contents) != 0 {
		t.Fatalf("multi-line frame: %q, %v", contents, err)
	}
}
