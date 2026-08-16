package mcpserver

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestBoundedToolHandlerLimitsConcurrentRequests(t *testing.T) {
	t.Parallel()
	limiter := newRequestLimiter(3)
	started := make(chan struct{}, 12)
	release := make(chan struct{})
	var active, maximum atomic.Int32
	handler := boundedToolHandler(limiter,
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
	handler := boundedToolHandler(limiter,
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
