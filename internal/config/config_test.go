package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}

func TestLoadReadsPluginOrderAndDisabledState(t *testing.T) {
	path := writeTestFile(t, "pluggent.yaml", `plugins:
  - id: runtime-info-provider
  - id: version-reporter
    disabled: true
`)

	configuration, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(configuration.Plugins) != 2 {
		t.Fatalf("len(Config.Plugins) = %d, want 2", len(configuration.Plugins))
	}
	if got := configuration.Plugins[0].ID; got != "runtime-info-provider" {
		t.Fatalf("first plugin ID = %q, want runtime-info-provider", got)
	}
	if !configuration.Plugins[1].Disabled {
		t.Fatal("second plugin Disabled = false, want true")
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	path := writeTestFile(t, "pluggent.yaml", `plugins:
  - id: runtime-info-provider
    unexpected: true
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
}

func TestLoadRejectsDuplicatePluginID(t *testing.T) {
	path := writeTestFile(t, "pluggent.yaml", `plugins:
  - id: runtime-info-provider
  - id: runtime-info-provider
`)

	_, err := Load(path)
	if !errors.Is(err, ErrDuplicatePluginID) {
		t.Fatalf("Load() error = %v, want ErrDuplicatePluginID", err)
	}
}

func TestLoadRejectsMultipleDocuments(t *testing.T) {
	path := writeTestFile(t, "pluggent.yaml", `plugins:
  - id: runtime-info-provider
---
plugins:
  - id: version-reporter
`)

	_, err := Load(path)
	if !errors.Is(err, ErrMultipleDocuments) {
		t.Fatalf("Load() error = %v, want ErrMultipleDocuments", err)
	}
}

func TestResolveConfigPathUsesEnvironmentValue(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key != ConfigPathEnvironmentVariable {
			return "", false
		}

		return " custom/pluggent.yaml ", true
	}

	if got := ResolveConfigPath(lookup); got != "custom/pluggent.yaml" {
		t.Fatalf("ResolveConfigPath() = %q, want custom/pluggent.yaml", got)
	}
}

func TestResolveConfigPathUsesDefault(t *testing.T) {
	if got := ResolveConfigPath(nil); got != DefaultConfigPath {
		t.Fatalf("ResolveConfigPath() = %q, want %q", got, DefaultConfigPath)
	}
}

func TestLoadEnvironmentFileAllowsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.env")

	if err := LoadEnvironmentFile(path); err != nil {
		t.Fatalf("LoadEnvironmentFile() error = %v", err)
	}
}

func TestProcessEnvironmentOverridesEnvironmentFile(t *testing.T) {
	t.Setenv(ConfigPathEnvironmentVariable, "process/pluggent.yaml")

	path := writeTestFile(
		t,
		".env",
		ConfigPathEnvironmentVariable+"=dotenv/pluggent.yaml\n",
	)

	if err := LoadEnvironmentFile(path); err != nil {
		t.Fatalf("LoadEnvironmentFile() error = %v", err)
	}

	if got := ResolveConfigPath(os.LookupEnv); got != "process/pluggent.yaml" {
		t.Fatalf(
			"ResolveConfigPath() = %q, want process/pluggent.yaml",
			got,
		)
	}
}
