package source_test

import (
	"testing"

	"github.com/projectious-work/ainfra/internal/source"
)

func TestParseCanonicalSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, value, ref string
		want             source.Reference
	}{
		{
			name: "local parent repository", value: "local:../template",
			want: source.Reference{
				Kind: source.KindLocal, Canonical: "local:../template",
				Display: "local:../template", LocalPath: "../template",
			},
		},
		{
			name:  "https repository subdirectory",
			value: "git::https://github.com/example/templates.git//providers/aws",
			ref:   "v1.2.3",
			want: source.Reference{
				Kind:         source.KindGit,
				Canonical:    "git::https://github.com/example/templates.git//providers/aws",
				Display:      "git::https://github.com/example/templates.git//providers/aws",
				Repository:   "https://github.com/example/templates.git",
				Subdirectory: "providers/aws", RequestedRef: "v1.2.3",
			},
		},
		{
			name: "ssh repository root", value: "git::ssh://git@example.com/templates",
			ref: "main",
			want: source.Reference{
				Kind:       source.KindGit,
				Canonical:  "git::ssh://git@example.com/templates",
				Display:    "git::ssh://redacted@example.com/templates",
				Repository: "ssh://git@example.com/templates", RequestedRef: "main",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := source.Parse(test.value, test.ref)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("reference = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseRedactsSensitiveGitURLParts(t *testing.T) {
	t.Parallel()
	got, err := source.Parse(
		"git::https://user:password@example.com/templates?access_token=secret&depth=1//aws",
		"main",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != "https://user:password@example.com/templates?access_token=secret&depth=1" {
		t.Fatalf("repository = %q", got.Repository)
	}
	if got.Display != "git::https://redacted@example.com/templates?access_token=redacted&depth=1//aws" {
		t.Fatalf("display = %q", got.Display)
	}
}

func TestParseRejectsInvalidSources(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, value, ref string }{
		{name: "empty"},
		{name: "surrounding whitespace", value: " local:template"},
		{name: "unknown scheme", value: "file:template"},
		{name: "local missing path", value: "local:"},
		{name: "local absolute", value: "local:/tmp/template"},
		{name: "local with ref", value: "local:template", ref: "main"},
		{name: "git missing ref", value: "git::https://example.com/repo.git"},
		{name: "git relative URL", value: "git::example.com/repo.git", ref: "main"},
		{name: "git unsupported scheme", value: "git::http://example.com/repo.git", ref: "main"},
		{name: "git fragment", value: "git::https://example.com/repo.git#main", ref: "main"},
		{name: "subdirectory absolute", value: "git::https://example.com/repo.git///etc", ref: "main"},
		{name: "subdirectory empty segment", value: "git::https://example.com/repo.git//a//b", ref: "main"},
		{name: "subdirectory backslash", value: "git::https://example.com/repo.git//a\\b", ref: "main"},
		{name: "subdirectory traversal", value: "git::https://example.com/repo.git//../a", ref: "main"},
		{name: "newline", value: "local:template\nother"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got, err := source.Parse(test.value, test.ref); err == nil {
				t.Fatalf("Parse() = %#v, want error", got)
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"local:../template",
		"git::https://example.com/repo.git//templates/aws",
		"git::ssh://git@example.com/repo.git",
		"git::https://user:secret@example.com/repo?token=secret//x",
	} {
		f.Add(seed, "main")
	}
	f.Fuzz(func(t *testing.T, value, ref string) {
		parsed, err := source.Parse(value, ref)
		if err != nil {
			return
		}
		if parsed.Canonical == "" || parsed.Display == "" {
			t.Fatalf("successful parse returned incomplete identity: %#v", parsed)
		}
		if parsed.Kind == source.KindGit && parsed.RequestedRef == "" {
			t.Fatal("successful Git parse returned no requested ref")
		}
	})
}
