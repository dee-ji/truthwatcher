package knowledge

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseDevelopmentConfig(t *testing.T) {
	cfg, err := Parse(strings.NewReader(`project:
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
        - adr

    - name: mistspren-github
      type: github
      enabled: false
      repo: github.com/dee-ji/mistspren
      branch: main
`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Project.Name != "truthwatcher" {
		t.Fatalf("Project.Name = %q, want truthwatcher", cfg.Project.Name)
	}
	if got, want := len(cfg.Providers), 2; got != want {
		t.Fatalf("provider count = %d, want %d", got, want)
	}
	if cfg.Providers[0].Name != "mistspren-local" || cfg.Providers[0].Type != "filesystem" || !cfg.Providers[0].Enabled {
		t.Fatalf("local provider parsed incorrectly: %+v", cfg.Providers[0])
	}
	if got, want := len(cfg.Providers[0].Purpose), 2; got != want {
		t.Fatalf("purpose count = %d, want %d", got, want)
	}
	if cfg.Providers[1].Enabled {
		t.Fatalf("github provider enabled = true, want false")
	}
}

func TestExpandEnvReportsMissingVariables(t *testing.T) {
	expanded, missing := ExpandEnv("${TRUTHWATCHER_HOME}/../${MISTSPREN_HOME}", func(key string) (string, bool) {
		if key == "TRUTHWATCHER_HOME" {
			return "/repo/truthwatcher", true
		}
		return "", false
	})

	if expanded != "/repo/truthwatcher/../" {
		t.Fatalf("expanded = %q, want /repo/truthwatcher/../", expanded)
	}
	if got, want := strings.Join(missing, ","), "MISTSPREN_HOME"; got != want {
		t.Fatalf("missing = %q, want %q", got, want)
	}
}

func TestCheckProviderReportsFilesystemAvailable(t *testing.T) {
	provider := Provider{Name: "mistspren-local", Type: "filesystem", Enabled: true, Root: "${MISTSPREN_HOME}"}
	dir := t.TempDir()

	result := CheckProvider(provider, func(key string) (string, bool) {
		return dir, key == "MISTSPREN_HOME"
	}, os.Stat)

	if result.Status != StatusAvailable {
		t.Fatalf("status = %q, want %q: %s", result.Status, StatusAvailable, result.Detail)
	}
	if result.Target != dir {
		t.Fatalf("target = %q, want %q", result.Target, dir)
	}
}

func TestCheckProviderReportsMissingEnvAsMissing(t *testing.T) {
	provider := Provider{Name: "mistspren-local", Type: "filesystem", Enabled: true, Root: "${MISTSPREN_HOME}"}

	result := CheckProvider(provider, func(string) (string, bool) {
		return "", false
	}, os.Stat)

	if result.Status != StatusMissing {
		t.Fatalf("status = %q, want %q", result.Status, StatusMissing)
	}
	if !strings.Contains(result.Detail, "MISTSPREN_HOME") {
		t.Fatalf("detail = %q, want missing env name", result.Detail)
	}
}

func TestCheckProviderReportsDisabledGitHub(t *testing.T) {
	provider := Provider{Name: "mistspren-github", Type: "github", Enabled: false, Repo: "github.com/dee-ji/mistspren"}

	result := CheckProvider(provider, nil, nil)

	if result.Status != StatusDisabled {
		t.Fatalf("status = %q, want %q", result.Status, StatusDisabled)
	}
	if result.Target != "github.com/dee-ji/mistspren" {
		t.Fatalf("target = %q, want github repo", result.Target)
	}
}

func TestCheckProviderReportsEnabledGitHubAsMisconfigured(t *testing.T) {
	provider := Provider{Name: "mistspren-github", Type: "github", Enabled: true, Repo: "github.com/dee-ji/mistspren"}

	result := CheckProvider(provider, nil, nil)

	if result.Status != StatusMisconfigured {
		t.Fatalf("status = %q, want %q", result.Status, StatusMisconfigured)
	}
	if !strings.Contains(result.Detail, "future remote workflows") {
		t.Fatalf("detail = %q, want future workflow boundary", result.Detail)
	}
}

func TestCheckProviderReportsNonDirectoryAsMisconfigured(t *testing.T) {
	provider := Provider{Name: "mistspren-local", Type: "filesystem", Enabled: true, Root: "/tmp/mistspren"}

	result := CheckProvider(provider, nil, func(string) (os.FileInfo, error) {
		return nil, errors.New("not used")
	})

	if result.Status != StatusMissing {
		t.Fatalf("status = %q, want %q", result.Status, StatusMissing)
	}
}
