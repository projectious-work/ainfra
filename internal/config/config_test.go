package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/projectious-work/ainfra/internal/config"
)

func TestResolveUsesNormativePrecedenceAndSortedProvenance(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	system := writeConfig(t, root, "system.yaml", `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
ui:
  format: json
logging:
  level: info
paths:
  cache: system-cache
`)
	user := writeConfig(t, root, "user.yaml", `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
ui:
  format: text
logging:
  level: debug
`)
	flagFormat := "json"
	effective, err := config.Resolve(config.ResolveOptions{
		Defaults: config.Defaults("/default/cache", "/default/runs"),
		Files: []config.File{
			{Layer: config.LayerSystem, Path: system},
			{Layer: config.LayerUser, Path: user},
		},
		Environment: map[string]string{"AINFRA_LOG_LEVEL": "trace"},
		Flags:       config.Patch{UI: config.UIPatch{Format: &flagFormat}},
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if effective.Settings.UI.Format != "json" ||
		effective.Settings.Logging.Level != "trace" {
		t.Fatalf("wrong precedence: %#v", effective.Settings)
	}
	if effective.Settings.Paths.Cache != filepath.Join(root, "system-cache") {
		t.Fatalf("relative file path not resolved: %q", effective.Settings.Paths.Cache)
	}
	for index := 1; index < len(effective.Entries); index++ {
		if effective.Entries[index-1].Key >= effective.Entries[index].Key {
			t.Fatalf("entries not unique and sorted: %#v", effective.Entries)
		}
	}
	entry := findEntry(t, effective.Entries, "ui.format")
	if entry.Source != config.LayerFlag || !reflect.DeepEqual(
		entry.Overridden,
		[]config.Layer{config.LayerDefault, config.LayerSystem, config.LayerUser},
	) {
		t.Fatalf("wrong provenance: %#v", entry)
	}
}

func TestResolveRejectsProhibitedProjectSettings(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project := writeConfig(t, root, "ainfra.config.yaml", `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
ui:
  color: never
paths:
  cache: unsafe-redirection
`)
	options := config.ResolveOptions{
		Defaults: config.Defaults("/default/cache", "/default/runs"),
		Files:    []config.File{{Layer: config.LayerProject, Path: project}},
	}
	if _, err := config.Resolve(options); err == nil {
		t.Fatal("prohibited project setting unexpectedly accepted")
	}

	options.DiagnosticProject = true
	effective, err := config.Resolve(options)
	if err != nil {
		t.Fatalf("diagnostic resolution: %v", err)
	}
	if effective.Settings.UI.Color != "never" ||
		effective.Settings.Paths.Cache != "/default/cache" ||
		!reflect.DeepEqual(effective.Rejected, []string{"paths.cache"}) {
		t.Fatalf("unsafe diagnostic result: %#v", effective)
	}
}

func TestResolveRejectsUnknownFieldsAndBadEnvironment(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	unknown := writeConfig(t, root, "unknown.yaml", `apiVersion: ainfra.projectious.work/v1
kind: CLIConfig
secret: forbidden
`)
	if _, err := config.Resolve(config.ResolveOptions{
		Defaults: config.Defaults("/cache", "/runs"),
		Files:    []config.File{{Layer: config.LayerExplicit, Path: unknown, Required: true}},
	}); err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
	for name, environment := range map[string]map[string]string{
		"empty boolean": {"AINFRA_NON_INTERACTIVE": ""},
		"empty path":    {"AINFRA_CACHE_DIR": ""},
		"loose boolean": {"AINFRA_NON_INTERACTIVE": "yes"},
		"relative path": {"AINFRA_TOFU_PATH": "bin/tofu"},
		"bad enum":      {"AINFRA_FORMAT": "yaml"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := config.EnvironmentPatch(environment); err == nil {
				t.Fatal("invalid environment unexpectedly accepted")
			}
		})
	}
}

func TestEnvironmentLoggingDestinationsAreClosedAndDeterministic(t *testing.T) {
	t.Parallel()
	patch, err := config.EnvironmentPatch(map[string]string{
		"AINFRA_LOG_FORMAT": "json",
		"AINFRA_LOG_FILE":   "/logs/ainfra.jsonl",
		"AINFRA_LOG_SYSLOG": "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	if patch.Logging.Destinations == nil || len(*patch.Logging.Destinations) != 3 {
		t.Fatalf("destinations = %#v", patch.Logging.Destinations)
	}
	destinations := *patch.Logging.Destinations
	if destinations[0].Type != "stderr" || destinations[1].Type != "file" ||
		destinations[2].Type != "syslog" || destinations[1].Rotation.MaxSizeMiB != 10 {
		t.Fatalf("destinations = %#v", destinations)
	}
}

func TestResolveReportsAbsentOptionalLayer(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	effective, err := config.Resolve(config.ResolveOptions{
		Defaults: config.Defaults("/cache", "/runs"),
		Files:    []config.File{{Layer: config.LayerUser, Path: missing}},
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(effective.Files) != 1 || effective.Files[0].Status != "absent" {
		t.Fatalf("missing layer report: %#v", effective.Files)
	}
}

func TestFilesUsesOnlyNormativeHostLocations(t *testing.T) {
	t.Parallel()
	tests := map[string][]string{
		"linux": {
			"/etc/ainfra/config.yaml",
			"/home/test/.config/ainfra/config.yaml",
			"/deployment/ainfra.config.yaml",
			"/explicit/config.yaml",
		},
		"darwin": {
			"/Library/Application Support/ainfra/config.yaml",
			"/home/test/Library/Application Support/ainfra/config.yaml",
			"/deployment/ainfra.config.yaml",
			"/explicit/config.yaml",
		},
	}
	for goos, want := range tests {
		files, err := config.Files(config.LocationOptions{
			GOOS: goos, HomeDirectory: "/home/test",
			DeploymentRoot: "/deployment", ExplicitPath: "/explicit/config.yaml",
		})
		if err != nil {
			t.Fatalf("%s locations: %v", goos, err)
		}
		got := make([]string, len(files))
		for index := range files {
			got[index] = files[index].Path
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s paths = %#v, want %#v", goos, got, want)
		}
	}
}

func findEntry(t *testing.T, entries []config.Entry, key string) config.Entry {
	t.Helper()
	for _, entry := range entries {
		if entry.Key == key {
			return entry
		}
	}
	t.Fatalf("missing entry %s", key)
	return config.Entry{}
}

func writeConfig(t *testing.T, root, name, contents string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
