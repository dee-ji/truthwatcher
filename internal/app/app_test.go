package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"truthwatcher/internal/config"
)

func TestVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := App{Version: "test-version"}.Run(context.Background(), []string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("version command returned error: %v", err)
	}

	if got, want := stdout.String(), "truthwatcher test-version\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestServerCommandStartsAndStops(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotConfig config.Config

	app := App{
		Version: "test-version",
		loadConfig: func() (config.Config, error) {
			return config.Default(), nil
		},
		serveHTTP: func(ctx context.Context, cfg config.Config, logger *slog.Logger, stdout io.Writer) error {
			if logger == nil {
				t.Fatal("logger is nil")
			}
			gotConfig = cfg
			fmt.Fprintln(stdout, "fake server started")
			return nil
		},
	}

	err := app.Run(context.Background(), []string{"server", "--addr", "127.0.0.1:0", "--config", "./truthwatcher.yaml"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("server command returned error: %v", err)
	}

	if gotConfig.HTTPAddr != "127.0.0.1:0" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:0", gotConfig.HTTPAddr)
	}
	if gotConfig.ConfigPath != "./truthwatcher.yaml" {
		t.Fatalf("ConfigPath = %q, want ./truthwatcher.yaml", gotConfig.ConfigPath)
	}
	if !strings.Contains(stdout.String(), "fake server started") {
		t.Fatalf("server output = %q, want fake startup message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := New().Run(context.Background(), []string{"bogus"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("unknown command returned nil error")
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
}
