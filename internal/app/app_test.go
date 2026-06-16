package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
		{
			name: "export",
			args: []string{"export", "--help"},
			want: []string{"Usage:", "truthwatcher export json", "assets, facts, relationships"},
		},
		{
			name: "export json",
			args: []string{"export", "json", "--help"},
			want: []string{"Usage:", "truthwatcher export json", "Raw evidence output is not exported"},
		},
		{
			name: "import",
			args: []string{"import", "--help"},
			want: []string{"Usage:", "truthwatcher import json", "import candidates"},
		},
		{
			name: "import json",
			args: []string{"import", "json", "--help"},
			want: []string{"Usage:", "truthwatcher import json", "does not persist records"},
		},
		{
			name: "dev",
			args: []string{"dev", "--help"},
			want: []string{"Usage:", "truthwatcher dev check-knowledge", "development helper"},
		},
		{
			name: "dev check-knowledge",
			args: []string{"dev", "check-knowledge", "--help"},
			want: []string{"Usage:", "truthwatcher dev check-knowledge", "MISTSPREN_HOME"},
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

func TestExportJSONRequiresDatabaseURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := App{
		Version: "test-version",
		loadConfig: func() (config.Config, error) {
			return config.Default(), nil
		},
	}

	err := app.Run(context.Background(), []string{"export", "json", "--output", filepath.Join(t.TempDir(), "snapshot.json")}, &stdout, &stderr)
	if err == nil {
		t.Fatal("export json without database url returned nil error")
	}
	if !strings.Contains(err.Error(), config.EnvDatabaseURL) {
		t.Fatalf("error = %q, want database url env name", err.Error())
	}
}

func TestImportJSONValidatesCandidatesWithoutDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	data := `{
  "schema_version": "truthwatcher.file_snapshot.v1",
  "generated_at": "2026-06-13T12:00:00Z",
  "assets": [{
    "id": "asset-a",
    "type": "device",
    "identity_key": "device:serial:aaa",
    "confidence": 0.95,
    "confidence_reason": "directly observed from evidence",
    "state": "observed",
    "metadata": {}
  }],
  "facts": [{
    "id": "fact-a",
    "asset_id": "asset-a",
    "name": "hostname",
    "value": "router-a",
    "source": "parser",
    "confidence": 0.95,
    "confidence_reason": "directly observed from evidence",
    "state": "observed",
    "evidence_id": "evidence-a"
  }],
  "relationships": [{
    "id": "relationship-a",
    "source_asset_id": "asset-a",
    "target_asset_id": "asset-b",
    "relationship_type": "lldp_neighbor_of",
    "confidence": 0.8,
    "confidence_reason": "directly observed from evidence",
    "state": "observed",
    "evidence_id": "evidence-a",
    "metadata": {}
  }],
  "evidence_metadata": [{
    "id": "evidence-a",
    "discovery_run_id": "run-a",
    "target": "router-a",
    "method": "ssh",
    "command_or_api": "show version",
    "raw_output_hash": "hash-a"
  }]
}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := New().Run(context.Background(), []string{"import", "json", "--input", path}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("import json returned error: %v", err)
	}
	output := stdout.String()
	for _, want := range []string{
		"validated import candidates",
		"assets: 1",
		"facts: 1",
		"relationships: 1",
		"evidence metadata: 1",
		"does not persist records or treat imported data as observed proof",
		"warning: imported observed facts are candidates only",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want substring %q", output, want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDevCheckKnowledgeReportsProviders(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "truthwatcher.yaml")
	mistsprenPath := t.TempDir()
	data := `project:
  name: truthwatcher
  repo: github.com/dee-ji/truthwatcher
  local_path: ${TRUTHWATCHER_HOME}

knowledge:
  providers:
    - name: mistspren-local
      type: filesystem
      enabled: true
      root: ${MISTSPREN_HOME}
      purpose:
        - memory

    - name: mistspren-github
      type: github
      enabled: false
      repo: github.com/dee-ji/mistspren
      branch: main
`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("MISTSPREN_HOME", mistsprenPath)
	t.Setenv("TRUTHWATCHER_HOME", t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := New().Run(context.Background(), []string{"dev", "check-knowledge", "--config", configPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dev check-knowledge returned error: %v", err)
	}
	output := stdout.String()
	for _, want := range []string{
		"provider\ttype\tenabled\ttarget\tstatus",
		"mistspren-local\tfilesystem\ttrue\t" + mistsprenPath + "\tavailable",
		"mistspren-github\tgithub\tfalse\tgithub.com/dee-ji/mistspren\tdisabled",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want substring %q", output, want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDevCheckKnowledgeReportsMissingMistsprenHome(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "truthwatcher.yaml")
	data := `knowledge:
  providers:
    - name: mistspren-local
      type: filesystem
      enabled: true
      root: ${MISTSPREN_HOME}
`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	original, hadOriginal := os.LookupEnv("MISTSPREN_HOME")
	if err := os.Unsetenv("MISTSPREN_HOME"); err != nil {
		t.Fatalf("unset MISTSPREN_HOME: %v", err)
	}
	t.Cleanup(func() {
		if hadOriginal {
			_ = os.Setenv("MISTSPREN_HOME", original)
			return
		}
		_ = os.Unsetenv("MISTSPREN_HOME")
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := New().Run(context.Background(), []string{"dev", "check-knowledge", "--config", configPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dev check-knowledge returned error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "mistspren-local\tfilesystem\ttrue\t-\tmissing") {
		t.Fatalf("stdout = %q, want missing provider", output)
	}
	if !strings.Contains(output, "missing environment variable: MISTSPREN_HOME") {
		t.Fatalf("stdout = %q, want missing env detail", output)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestDevCheckKnowledgeMissingConfigDoesNotFail(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	missingPath := filepath.Join(t.TempDir(), "missing.yaml")
	err := New().Run(context.Background(), []string{"dev", "check-knowledge", "--config", missingPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dev check-knowledge returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no knowledge providers configured") {
		t.Fatalf("stdout = %q, want no providers message", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
