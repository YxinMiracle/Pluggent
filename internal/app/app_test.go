package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yxinmiracle/pluggent/internal/plugin"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "pluggent.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}

func TestRunWritesVersion(t *testing.T) {
	var stdout bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath: writeConfig(t, `plugins:
  - id: runtime-info-provider
  - id: version-reporter
`),
		Stdout:  &stdout,
		Version: "test-version",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	const want = "pluggent test-version\n"
	if got := stdout.String(); got != want {
		t.Fatalf("Run() output = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownPlugin(t *testing.T) {
	var stdout bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath: writeConfig(t, `plugins:
  - id: missing-plugin
`),
		Stdout:  &stdout,
		Version: "test-version",
	})
	if !errors.Is(err, ErrUnknownPluginID) {
		t.Fatalf("Run() error = %v, want ErrUnknownPluginID", err)
	}

	if got := stdout.String(); got != "" {
		t.Fatalf("Run() output = %q, want empty output", got)
	}
}

func TestRunRequiresProviderBeforeConsumer(t *testing.T) {
	var stdout bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath: writeConfig(t, `plugins:
  - id: version-reporter
  - id: runtime-info-provider
`),
		Stdout:  &stdout,
		Version: "test-version",
	})
	if !errors.Is(err, plugin.ErrServiceNotFound) {
		t.Fatalf("Run() error = %v, want plugin.ErrServiceNotFound", err)
	}

	if got := stdout.String(); got != "" {
		t.Fatalf("Run() output = %q, want empty output", got)
	}
}

func TestRunReturnsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stdout bytes.Buffer

	err := Run(ctx, Options{
		Stdout:  &stdout,
		Version: "test-version",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}

	if got := stdout.String(); got != "" {
		t.Fatalf("Run() output = %q, want empty output", got)
	}
}
