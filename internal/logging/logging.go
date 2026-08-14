// Package logging delivers already-redacted structured operational events.
package logging

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/projectious-work/ainfra/internal/security"
)

// Event is the sole value accepted by operational sinks.
type Event struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Command   string `json:"command"`
	RunID     string `json:"runId,omitempty"`
}

// NewEvent constructs a UTC, pre-redacted event shared by every sink.
func NewEvent(at time.Time, level, component, message, command, runID string,
	sensitiveValues []string) Event {
	return Event{Timestamp: at.UTC().Format(time.RFC3339Nano), Level: level,
		Component: security.RedactString(component, sensitiveValues),
		Message:   security.RedactString(message, sensitiveValues),
		Command:   security.RedactString(command, sensitiveValues),
		RunID:     security.RedactString(runID, sensitiveValues)}
}

// Sink receives only the already-redacted Event value.
type Sink interface{ WriteEvent(Event) error }

// Logger fans one immutable event out to configured sinks.
type Logger struct{ Sinks []Sink }

func (logger Logger) Write(event Event) error {
	failed := make([]int, 0)
	var failures []error
	for index, sink := range logger.Sinks {
		if err := sink.WriteEvent(event); err != nil {
			failed = append(failed, index)
			failures = append(failures, fmt.Errorf("operational log sink %d: %w", index, err))
		}
	}
	if len(failures) == 0 {
		return nil
	}
	notice := event
	notice.Level = "error"
	notice.Component = "logging"
	notice.Message = fmt.Sprintf("%d operational log sink(s) failed", len(failures))
	for index, sink := range logger.Sinks {
		if contains(failed, index) {
			continue
		}
		_ = sink.WriteEvent(notice)
	}
	return errors.Join(failures...)
}

func contains(values []int, value int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// WriterSink writes text or JSON events to a process-owned writer.
type WriterSink struct {
	Writer io.Writer
	Format string
}

func (sink WriterSink) WriteEvent(event Event) error {
	if sink.Writer == nil {
		return errors.New("log writer is unavailable")
	}
	contents, err := render(event, sink.Format)
	if err != nil {
		return err
	}
	_, err = sink.Writer.Write(contents)
	return err
}

// FileOptions controls deterministic owner-only rotation.
type FileOptions struct {
	Path       string
	Format     string
	MaxBytes   int64
	MaxBackups int
	MaxAge     time.Duration
	Compress   bool
	Now        func() time.Time
}

// FileSink is a synchronized, symlink-refusing rotating destination.
type FileSink struct {
	mu      sync.Mutex
	options FileOptions
}

func NewFileSink(options FileOptions) (*FileSink, error) {
	if !filepath.IsAbs(options.Path) {
		return nil, errors.New("absolute operational log path required")
	}
	if options.MaxBytes <= 0 || options.MaxBackups < 1 {
		return nil, errors.New("positive rotation size and backup count required")
	}
	parent := filepath.Dir(options.Path)
	if info, err := os.Lstat(parent); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("safe operational log directory required")
	}
	if info, err := os.Lstat(options.Path); err == nil &&
		(!info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0) {
		return nil, errors.New("operational log must be an owner-only regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return &FileSink{options: options}, nil
}

func (sink *FileSink) WriteEvent(event Event) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	contents, err := render(event, sink.options.Format)
	if err != nil {
		return err
	}
	if err := sink.purgeAged(); err != nil {
		return err
	}
	if info, err := os.Lstat(sink.options.Path); err == nil {
		if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
			return errors.New("operational log path changed or became unsafe")
		}
		if info.Size()+int64(len(contents)) > sink.options.MaxBytes {
			if err := sink.rotate(); err != nil {
				return err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	parent, err := os.OpenRoot(filepath.Dir(sink.options.Path))
	if err != nil {
		return err
	}
	file, err := parent.OpenFile(filepath.Base(sink.options.Path),
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	parentCloseErr := parent.Close()
	if err != nil {
		return errors.Join(err, parentCloseErr)
	}
	if parentCloseErr != nil {
		_ = file.Close()
		return parentCloseErr
	}
	info, statErr := file.Stat()
	if statErr != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		_ = file.Close()
		return errors.New("opened operational log is unsafe")
	}
	_, writeErr := file.Write(contents)
	return errors.Join(writeErr, file.Sync(), file.Close())
}

func (sink *FileSink) rotate() error {
	for index := sink.options.MaxBackups; index >= 1; index-- {
		current := backupName(sink.options.Path, index, sink.options.Compress)
		if index == sink.options.MaxBackups {
			if err := os.Remove(current); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}
		next := backupName(sink.options.Path, index+1, sink.options.Compress)
		if err := os.Rename(current, next); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if sink.options.Compress {
		return compressFile(sink.options.Path, backupName(sink.options.Path, 1, true))
	}
	return os.Rename(sink.options.Path, backupName(sink.options.Path, 1, false))
}

func (sink *FileSink) purgeAged() error {
	if sink.options.MaxAge <= 0 {
		return nil
	}
	now := time.Now
	if sink.options.Now != nil {
		now = sink.options.Now
	}
	for index := 1; index <= sink.options.MaxBackups; index++ {
		path := backupName(sink.options.Path, index, sink.options.Compress)
		if info, err := os.Lstat(path); err == nil && now().Sub(info.ModTime()) > sink.options.MaxAge {
			if err := os.Remove(path); err != nil {
				return err
			}
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func render(event Event, format string) ([]byte, error) {
	if format == "json" {
		var contents bytes.Buffer
		encoder := json.NewEncoder(&contents)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(event)
		return contents.Bytes(), err
	}
	if format != "text" {
		return nil, errors.New("unsupported operational log format")
	}
	return []byte(fmt.Sprintf("%s %s %s %s command=%s run=%s\n", event.Timestamp,
		event.Level, event.Component, event.Message, event.Command, event.RunID)), nil
}

func backupName(path string, index int, compressed bool) string {
	suffix := fmt.Sprintf(".%d", index)
	if compressed {
		suffix += ".gz"
	}
	return path + suffix
}

func compressFile(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	compressor := gzip.NewWriter(output)
	compressor.ModTime = time.Unix(0, 0).UTC()
	_, copyErr := io.Copy(compressor, input)
	err = errors.Join(copyErr, compressor.Close(), output.Sync(), output.Close(), input.Close())
	if err != nil {
		return err
	}
	return os.Remove(source)
}
