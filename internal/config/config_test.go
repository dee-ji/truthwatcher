package config

import "testing"

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
	}
	if cfg.LogLevel != DefaultLogLevel {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, DefaultLogLevel)
	}
	if cfg.DevMode {
		t.Fatal("DevMode = true, want false")
	}
	if cfg.ConfigPath != "" {
		t.Fatalf("ConfigPath = %q, want empty", cfg.ConfigPath)
	}
}

func TestLoadFromLookupUsesEnvironment(t *testing.T) {
	env := map[string]string{
		EnvAddr:        "0.0.0.0:9090",
		EnvDatabaseURL: "postgres://truthwatcher@example.invalid/truthwatcher",
		EnvLogLevel:    "DEBUG",
		EnvDevMode:     "true",
	}

	cfg, err := LoadFromLookup(func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("LoadFromLookup returned error: %v", err)
	}

	if cfg.HTTPAddr != "0.0.0.0:9090" {
		t.Fatalf("HTTPAddr = %q, want 0.0.0.0:9090", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://truthwatcher@example.invalid/truthwatcher" {
		t.Fatalf("DatabaseURL = %q, want env value", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if !cfg.DevMode {
		t.Fatal("DevMode = false, want true")
	}
}

func TestValidateRejectsMissingHTTPAddr(t *testing.T) {
	cfg := Default()
	cfg.HTTPAddr = " "

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for missing HTTP address")
	}
}

func TestValidateRejectsInvalidHTTPAddr(t *testing.T) {
	cfg := Default()
	cfg.HTTPAddr = "127.0.0.1"

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for invalid HTTP address")
	}
}

func TestValidateRejectsInvalidLogLevel(t *testing.T) {
	cfg := Default()
	cfg.LogLevel = "trace"

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for invalid log level")
	}
}

func TestLoadFromLookupRejectsInvalidDevMode(t *testing.T) {
	_, err := LoadFromLookup(func(key string) (string, bool) {
		if key == EnvDevMode {
			return "sometimes", true
		}
		return "", false
	})
	if err == nil {
		t.Fatal("LoadFromLookup returned nil error for invalid dev mode")
	}
}
