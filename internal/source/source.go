// Package source parses and canonicalizes immutable template source references.
package source

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"
)

// Kind identifies one supported v1 template source scheme.
type Kind string

const (
	// KindLocal is a contained filesystem source resolved by policy later.
	KindLocal Kind = "local"
	// KindGit is an HTTPS or SSH Git repository with an optional subdirectory.
	KindGit Kind = "git"
)

// Reference is the canonical structural form of a requested template source.
type Reference struct {
	Kind         Kind
	Canonical    string
	Display      string
	LocalPath    string
	Repository   string
	Subdirectory string
	RequestedRef string
}

// Parse validates and canonicalizes one v1 source plus its separately declared
// Git ref. It performs no filesystem access, network access, or acquisition.
func Parse(value, requestedRef string) (Reference, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return Reference{}, errors.New("template source must be non-empty and contain no surrounding whitespace")
	}
	if !utf8.ValidString(value) || strings.ContainsAny(value, "\x00\r\n") {
		return Reference{}, errors.New("template source must be valid single-line UTF-8")
	}
	switch {
	case strings.HasPrefix(value, "local:"):
		return parseLocal(value, requestedRef)
	case strings.HasPrefix(value, "git::"):
		return parseGit(value, requestedRef)
	default:
		return Reference{}, fmt.Errorf("unsupported template source scheme in %q", value)
	}
}

func parseLocal(value, requestedRef string) (Reference, error) {
	if requestedRef != "" {
		return Reference{}, errors.New("local template source must not declare a Git ref")
	}
	rawPath := strings.TrimPrefix(value, "local:")
	cleaned, err := cleanRelativePath(rawPath, true)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid local template source: %w", err)
	}
	canonical := "local:" + cleaned
	return Reference{
		Kind: KindLocal, Canonical: canonical, Display: canonical,
		LocalPath: cleaned,
	}, nil
}

func parseGit(value, requestedRef string) (Reference, error) {
	if requestedRef == "" || strings.TrimSpace(requestedRef) != requestedRef ||
		!utf8.ValidString(requestedRef) || strings.ContainsAny(requestedRef, "\x00\r\n") {
		return Reference{}, errors.New("git template source requires a non-blank explicit ref")
	}
	remainder := strings.TrimPrefix(value, "git::")
	repository, subdirectory := splitGitSubdirectory(remainder)
	parsed, err := url.Parse(repository)
	if err != nil {
		return Reference{}, fmt.Errorf("parse Git repository URL: %w", err)
	}
	if (parsed.Scheme != "https" && parsed.Scheme != "ssh") ||
		parsed.Host == "" || parsed.Path == "" || parsed.Fragment != "" {
		return Reference{}, errors.New("git repository must be an absolute HTTPS or SSH URL without a fragment")
	}
	cleanedSubdirectory, err := cleanRelativePath(subdirectory, false)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid Git template subdirectory: %w", err)
	}
	if cleanedSubdirectory == ".." || strings.HasPrefix(cleanedSubdirectory, "../") {
		return Reference{}, errors.New("git template subdirectory must not traverse its checkout")
	}
	repository = parsed.String()
	canonical := "git::" + repository
	display := "git::" + redactURL(parsed)
	if cleanedSubdirectory != "" {
		canonical += "//" + cleanedSubdirectory
		display += "//" + cleanedSubdirectory
	}
	return Reference{
		Kind: KindGit, Canonical: canonical, Display: display,
		Repository: repository, Subdirectory: cleanedSubdirectory,
		RequestedRef: requestedRef,
	}, nil
}

func splitGitSubdirectory(value string) (string, string) {
	scheme := strings.Index(value, "://")
	if scheme < 0 {
		return value, ""
	}
	searchFrom := scheme + len("://")
	if delimiter := strings.Index(value[searchFrom:], "//"); delimiter >= 0 {
		delimiter += searchFrom
		return value[:delimiter], value[delimiter+2:]
	}
	return value, ""
}

func cleanRelativePath(value string, required bool) (string, error) {
	if value == "" {
		if required {
			return "", errors.New("path is required")
		}
		return "", nil
	}
	if !utf8.ValidString(value) || strings.ContainsAny(value, "\\\x00\r\n") {
		return "", errors.New("path must be valid single-line UTF-8 using slash separators")
	}
	if strings.HasPrefix(value, "/") {
		return "", errors.New("path must be relative")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" {
			return "", errors.New("path must not contain an empty segment")
		}
	}
	cleaned := path.Clean(value)
	if cleaned == "." && !required {
		return "", nil
	}
	if cleaned == "." && value != "." {
		return "", errors.New("path is ambiguous after cleaning")
	}
	return cleaned, nil
}

func redactURL(value *url.URL) string {
	copy := *value
	if copy.User != nil {
		copy.User = url.User("redacted")
	}
	query := copy.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "credential") ||
			lower == "key" || lower == "sig" || strings.Contains(lower, "signature") {
			query.Set(key, "redacted")
		}
	}
	copy.RawQuery = query.Encode()
	return copy.String()
}
