// Package logging delivers already-redacted structured operational events.
package logging

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/projectious-work/ainfra/internal/config"
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
type Logger struct {
	Sinks []Sink
	Level string
}

// Build constructs every explicitly configured sink or fails atomically.
func Build(settings config.Logging, stderr io.Writer) (Logger, func() error, error) {
	sinks := make([]Sink, 0, len(settings.Destinations))
	closers := make([]io.Closer, 0)
	closeAll := func() error {
		var failures []error
		for _, closer := range closers {
			failures = append(failures, closer.Close())
		}
		return errors.Join(failures...)
	}
	for _, destination := range settings.Destinations {
		switch destination.Type {
		case "stderr":
			sinks = append(sinks, &WriterSink{Writer: stderr, Format: destination.Format})
		case "file":
			file, err := NewFileSink(FileOptions{Path: destination.Path,
				Format: destination.Format, MaxBytes: int64(destination.Rotation.MaxSizeMiB) << 20,
				MaxBackups: destination.Rotation.MaxBackups,
				MaxAge:     time.Duration(destination.Rotation.MaxAgeDays) * 24 * time.Hour,
				Compress:   destination.Rotation.Compress})
			if err != nil {
				_ = closeAll()
				return Logger{}, nil, err
			}
			sinks = append(sinks, file)
		case "syslog":
			writer, err := syslog.Dial("", "", syslogFacility(destination.Facility)|syslog.LOG_INFO,
				destination.Tag)
			if err != nil {
				_ = closeAll()
				return Logger{}, nil, fmt.Errorf("initialize local syslog: %w", err)
			}
			closers = append(closers, writer)
			sinks = append(sinks, syslogSink{writer: writer})
		default:
			_ = closeAll()
			return Logger{}, nil, fmt.Errorf("unsupported operational log sink %q", destination.Type)
		}
	}
	return Logger{Sinks: sinks, Level: settings.Level}, closeAll, nil
}

type syslogSink struct{ writer *syslog.Writer }

func (sink syslogSink) WriteEvent(event Event) error {
	contents, err := render(event, "json")
	if err != nil {
		return err
	}
	return sink.writer.Info(strings.TrimSpace(string(contents)))
}

func syslogFacility(value string) syslog.Priority {
	switch value {
	case "local0":
		return syslog.LOG_LOCAL0
	case "local1":
		return syslog.LOG_LOCAL1
	case "local2":
		return syslog.LOG_LOCAL2
	case "local3":
		return syslog.LOG_LOCAL3
	case "local4":
		return syslog.LOG_LOCAL4
	case "local5":
		return syslog.LOG_LOCAL5
	case "local6":
		return syslog.LOG_LOCAL6
	case "local7":
		return syslog.LOG_LOCAL7
	default:
		return syslog.LOG_USER
	}
}

func (logger Logger) Write(event Event) error {
	if !enabled(logger.Level, event.Level) {
		return nil
	}
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

func enabled(minimum, level string) bool {
	if minimum == "" {
		return true
	}
	order := map[string]int{"error": 0, "warn": 1, "info": 2, "debug": 3, "trace": 4}
	minimumValue, minimumOK := order[minimum]
	levelValue, levelOK := order[level]
	return minimumOK && levelOK && levelValue <= minimumValue
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
	mu     sync.Mutex
	Writer io.Writer
	Format string
}

func (sink *WriterSink) WriteEvent(event Event) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
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
	directory := filepath.Dir(source)
	if filepath.Dir(target) != directory {
		return errors.New("compressed log target must share the source directory")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	input, err := root.Open(filepath.Base(source))
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	output, err := root.OpenFile(filepath.Base(target),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
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
	return root.Remove(filepath.Base(source))
}
