package app

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestRunWritesVersion(t *testing.T) {
	var stdout bytes.Buffer

	err := Run(context.Background(), Options{
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
