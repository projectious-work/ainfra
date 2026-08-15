package security

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sync"
)

// RedactString applies the same chunk-safe exact-value boundary used for child
// streams to one complete structured field.
func RedactString(value string, sensitiveValues []string) string {
	var output bytes.Buffer
	writer := NewRedactingWriter(&output, sensitiveValues)
	_, _ = writer.Write([]byte(value))
	_ = writer.Close()
	redacted := output.String()
	redacted = bearerPattern.ReplaceAllString(redacted, "Bearer "+redaction)
	redacted = credentialPattern.ReplaceAllString(redacted, "${1}"+redaction)
	redacted = urlUserinfoPattern.ReplaceAllString(redacted, "${1}"+redaction+"@")
	redacted = tokenShapePattern.ReplaceAllString(redacted, redaction)
	return redacted
}

var (
	bearerPattern      = regexp.MustCompile(`(?i)Bearer[ \t]+[A-Za-z0-9._~+/=-]{8,}`)
	credentialPattern  = regexp.MustCompile(`(?i)((?:access_token|api[_-]?key|password|secret|token)=)[^&\s]+`)
	urlUserinfoPattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9+.-]*://)[^/@\s]+@`)
	tokenShapePattern  = regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16})\b`)
)

const redaction = "<redacted>"

// RedactingWriter removes registered exact sensitive values across writes.
type RedactingWriter struct {
	mu      sync.Mutex
	writer  io.Writer
	secrets [][]byte
	pending []byte
	closed  bool
}

// NewRedactingWriter creates a bounded streaming exact-value redactor.
func NewRedactingWriter(writer io.Writer, values []string) *RedactingWriter {
	secrets := make([][]byte, 0, len(values))
	for _, value := range values {
		if value != "" {
			secrets = append(secrets, []byte(value))
		}
	}
	return &RedactingWriter{writer: writer, secrets: secrets}
}

// Write consumes bytes while retaining only a possible secret prefix.
func (writer *RedactingWriter) Write(input []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.closed {
		return 0, fmt.Errorf("write closed redactor")
	}
	for _, value := range input {
		writer.pending = append(writer.pending, value)
		for len(writer.pending) > 0 {
			matched, prefix := writer.classifyPending()
			if matched {
				if _, err := io.WriteString(writer.writer, redaction); err != nil {
					return 0, err
				}
				writer.pending = writer.pending[:0]
				break
			}
			if prefix {
				break
			}
			if _, err := writer.writer.Write(writer.pending[:1]); err != nil {
				return 0, err
			}
			writer.pending = writer.pending[1:]
		}
	}
	return len(input), nil
}

// Close flushes non-secret pending bytes. It does not close the destination.
func (writer *RedactingWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.closed {
		return nil
	}
	writer.closed = true
	if len(writer.pending) == 0 {
		return nil
	}
	_, err := writer.writer.Write(writer.pending)
	writer.pending = nil
	return err
}

func (writer *RedactingWriter) classifyPending() (bool, bool) {
	prefix := false
	for _, secret := range writer.secrets {
		if bytes.Equal(writer.pending, secret) {
			return true, true
		}
		if len(writer.pending) < len(secret) && bytes.Equal(writer.pending, secret[:len(writer.pending)]) {
			prefix = true
		}
	}
	return false, prefix
}
