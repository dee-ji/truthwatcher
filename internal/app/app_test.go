package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"truthwatcher/internal/config"
	"truthwatcher/internal/policy"
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

func TestCommandHelpDoesNotRequireRuntimeDependencies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "root",
			args: []string{"--help"},
			want: []string{"Usage:", "truthwatcher server", "truthwatcher migrate", "truthwatcher discover fake"},
		},
		{
			name: "version",
			args: []string{"version", "--help"},
			want: []string{"Usage:", "truthwatcher version"},
		},
		{
			name: "server",
			args: []string{"server", "--help"},
			want: []string{"Usage:", "truthwatcher server", "embedded UI"},
		},
		{
			name: "migrate",
			args: []string{"migrate", "--help"},
			want: []string{"Usage:", "truthwatcher migrate up", "embedded PostgreSQL migrations"},
		},
		{
			name: "discover",
			args: []string{"discover", "--help"},
			want: []string{"Usage:", "truthwatcher discover fake", "without touching a network"},
		},
		{
			name: "discover fake",
			args: []string{"discover", "fake", "--help"},
			want: []string{"Usage:", "truthwatcher discover fake", "read-only policy engine"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			err := New().Run(context.Background(), tt.args, &stdout, &stderr)
			if err != nil {
				t.Fatalf("help command returned error: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
				}
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
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

func TestMigrateRequiresSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := New().Run(context.Background(), []string{"migrate"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("migrate without subcommand returned nil error")
	}
}

func TestMigrateRequiresDatabaseURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := App{
		Version: "test-version",
		loadConfig: func() (config.Config, error) {
			return config.Default(), nil
		},
	}

	err := app.Run(context.Background(), []string{"migrate", "status"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("migrate without database url returned nil error")
	}
	if !strings.Contains(err.Error(), config.EnvDatabaseURL) {
		t.Fatalf("error = %q, want database url env name", err.Error())
	}
}

func TestParseDiscoveryTasksDefaults(t *testing.T) {
	tasks, err := parseDiscoveryTasks("")
	if err != nil {
		t.Fatalf("parseDiscoveryTasks returned error: %v", err)
	}
	if got, want := len(tasks), 4; got != want {
		t.Fatalf("task count = %d, want %d", got, want)
	}
}

func TestParseDiscoveryTasksRejectsUnknownTask(t *testing.T) {
	_, err := parseDiscoveryTasks("identify_device,format_disk")
	if !errors.Is(err, policy.ErrTaskNotAllowed) {
		t.Fatalf("expected ErrTaskNotAllowed, got %v", err)
	}
}

func TestDiscoverFakeRequiresDatabaseURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := App{
		Version: "test-version",
		loadConfig: func() (config.Config, error) {
			return config.Default(), nil
		},
	}

	err := app.Run(context.Background(), []string{"discover", "fake", "--target", "fixture://junos-mx"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("discover fake without database url returned nil error")
	}
	if !strings.Contains(err.Error(), config.EnvDatabaseURL) {
		t.Fatalf("error = %q, want database url env name", err.Error())
	}
}
