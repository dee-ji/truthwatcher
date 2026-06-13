package extensibility

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/policy"
)

func TestScriptRunnerDisabledByDefault(t *testing.T) {
	runner := NewScriptRunner(ScriptRunnerOptions{})

	_, err := runner.Import(context.Background(), ImportRequest{Source: "/tmp/not-run"})
	if !errors.Is(err, ErrScriptRunnerDisabled) {
		t.Fatalf("error = %v, want disabled", err)
	}
}

func TestScriptRunnerRequiresAllowlistedScript(t *testing.T) {
	script := writeTestScript(t, `#!/bin/sh
echo '{}'
`)
	runner := NewScriptRunner(ScriptRunnerOptions{Enabled: true})

	_, err := runner.Import(context.Background(), ImportRequest{Source: script})
	if !errors.Is(err, ErrScriptNotAllowed) {
		t.Fatalf("error = %v, want not allowlisted", err)
	}
}

func TestScriptRunnerImportsEvidenceAndFacts(t *testing.T) {
	script := writeTestScript(t, `#!/bin/sh
cat >/dev/null
cat <<'JSON'
{
  "evidence": [
    {
      "target": "router-a",
      "method": "script",
      "command_or_api": "show version",
      "raw_output": "Hostname: router-a",
      "metadata": {"source": "test-script"}
    }
  ],
  "candidates": {
    "facts": [
      {
        "asset_id": "asset-a",
        "name": "hostname",
        "value": "router-a",
        "source": "byo_script",
        "confidence": 0.4,
        "confidence_reason": "returned by local BYO script",
        "state": "user_seeded"
      }
    ]
  }
}
JSON
`)
	runner := NewScriptRunner(ScriptRunnerOptions{Enabled: true, AllowedScripts: []string{script}})
	scope := json.RawMessage(`{"target":"router-a","tasks":["identify_device"]}`)

	result, err := runner.Import(context.Background(), ImportRequest{Source: script, Scope: scope})
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if got, want := len(result.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if result.Evidence[0].CommandOrAPI != "show version" {
		t.Fatalf("command_or_api = %q, want show version", result.Evidence[0].CommandOrAPI)
	}
	if got, want := len(result.Candidates.Facts), 1; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	fact := result.Candidates.Facts[0]
	if fact.Source != "byo_script" {
		t.Fatalf("fact source = %q, want byo_script", fact.Source)
	}
	if fact.State != assets.StateUserSeeded {
		t.Fatalf("fact state = %q, want user_seeded", fact.State)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("warnings are empty")
	}
}

func TestScriptRunnerRejectsDeniedInputTask(t *testing.T) {
	script := writeTestScript(t, `#!/bin/sh
echo '{}'
`)
	runner := NewScriptRunner(ScriptRunnerOptions{Enabled: true, AllowedScripts: []string{script}})

	_, err := runner.Import(context.Background(), ImportRequest{
		Source: script,
		Scope:  json.RawMessage(`{"tasks":["configure"]}`),
	})
	if err == nil {
		t.Fatal("Import returned nil error for denied task")
	}
}

func TestScriptRunnerRejectsDeniedEvidenceCommand(t *testing.T) {
	script := writeTestScript(t, `#!/bin/sh
cat >/dev/null
cat <<'JSON'
{
  "evidence": [
    {
      "target": "router-a",
      "method": "script",
      "command_or_api": "reload",
      "raw_output": "dangerous"
    }
  ]
}
JSON
`)
	runner := NewScriptRunner(ScriptRunnerOptions{Enabled: true, AllowedScripts: []string{script}, Policy: policy.NewEngine()})

	_, err := runner.Import(context.Background(), ImportRequest{Source: script})
	if err == nil {
		t.Fatal("Import returned nil error for denied evidence command")
	}
}

func TestScriptRunnerTimeout(t *testing.T) {
	script := writeTestScript(t, `#!/bin/sh
sleep 1
`)
	runner := NewScriptRunner(ScriptRunnerOptions{
		Enabled:        true,
		AllowedScripts: []string{script},
		Timeout:        10 * time.Millisecond,
	})

	_, err := runner.Import(context.Background(), ImportRequest{Source: script})
	if err == nil {
		t.Fatal("Import returned nil error for timeout")
	}
}

func writeTestScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}
