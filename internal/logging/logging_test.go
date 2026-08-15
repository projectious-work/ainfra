package logging_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	operational "github.com/projectious-work/ainfra/internal/logging"
)

func TestLoggerSharesOnePreRedactedEventAcrossSinks(t *testing.T) {
	t.Parallel()
	var text, json bytes.Buffer
	event := operational.NewEvent(time.Unix(0, 0), "warn", "apply",
		"credential secret-value refused", "apply", "run-secret-value",
		[]string{"secret-value"})
	logger := operational.Logger{Sinks: []operational.Sink{
		&operational.WriterSink{Writer: &text, Format: "text"},
		&operational.WriterSink{Writer: &json, Format: "json"},
	}}
	if err := logger.Write(event); err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{"text": text.String(), "json": json.String()} {
		if strings.Contains(contents, "secret-value") || !strings.Contains(contents, "<redacted>") {
			t.Fatalf("%s sink did not receive pre-redacted event: %q", name, contents)
		}
	}
}

func TestLoggerSurfacesSinkFailureOnHealthySink(t *testing.T) {
	t.Parallel()
	var healthy bytes.Buffer
	logger := operational.Logger{Sinks: []operational.Sink{
		&operational.WriterSink{Writer: failingWriter{}, Format: "json"},
		&operational.WriterSink{Writer: &healthy, Format: "json"},
	}}
	if err := logger.Write(operational.NewEvent(time.Now(), "warn", "test",
		"message", "status", "", nil)); err == nil {
		t.Fatal("sink failure was not returned")
	}
	if !strings.Contains(healthy.String(), "operational log sink(s) failed") {
		t.Fatalf("healthy sink did not receive failure notice: %q", healthy.String())
	}
}

func TestWriterSinkPreservesConcurrentCorrelatedRecords(t *testing.T) {
	t.Parallel()
	var destination bytes.Buffer
	logger := operational.Logger{Sinks: []operational.Sink{
		&operational.WriterSink{Writer: &destination, Format: "json"},
	}}
	const count = 100
	var writers sync.WaitGroup
	writers.Add(count)
	for index := range count {
		go func() {
			defer writers.Done()
			runID := fmt.Sprintf("run-%03d", index)
			if err := logger.Write(operational.NewEvent(time.Unix(0, 0), "info",
				"child", "stream event", "apply", runID, nil)); err != nil {
				t.Errorf("write correlated event: %v", err)
			}
		}()
	}
	writers.Wait()
	lines := bytes.Split(bytes.TrimSpace(destination.Bytes()), []byte{'\n'})
	if len(lines) != count {
		t.Fatalf("record count=%d want=%d", len(lines), count)
	}
	seen := make(map[string]bool, count)
	for _, line := range lines {
		var event operational.Event
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("interleaved event record %q: %v", line, err)
		}
		if event.RunID == "" || seen[event.RunID] {
			t.Fatalf("invalid correlated event: %#v", event)
		}
		seen[event.RunID] = true
	}
}

func TestFileSinkRotatesCompressedPrivateFilesAndRejectsSymlink(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "ainfra.log")
	sink, err := operational.NewFileSink(operational.FileOptions{Path: path,
		Format: "text", MaxBytes: 100, MaxBackups: 2, Compress: true})
	if err != nil {
		t.Fatal(err)
	}
	event := operational.NewEvent(time.Unix(0, 0), "warn", "component",
		strings.Repeat("x", 80), "apply", "run", nil)
	if err := sink.WriteEvent(event); err != nil {
		t.Fatal(err)
	}
	if err := sink.WriteEvent(event); err != nil {
		t.Fatal(err)
	}
	backup := path + ".1.gz"
	info, err := os.Stat(backup)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("backup info=%v err=%v", info, err)
	}
	file, err := os.Open(backup)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(reader)
	if err != nil || !strings.Contains(string(contents), strings.Repeat("x", 80)) {
		t.Fatalf("rotated contents=%q err=%v", contents, err)
	}
	_ = reader.Close()
	_ = file.Close()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(backup, path); err != nil {
		t.Fatal(err)
	}
	if err := sink.WriteEvent(event); err == nil {
		t.Fatal("symlink replacement unexpectedly accepted")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("fixture failure") }
