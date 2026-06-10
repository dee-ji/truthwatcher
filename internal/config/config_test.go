package config

import "testing"

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
	if cfg.ConfigPath != "" {
		t.Fatalf("ConfigPath = %q, want empty", cfg.ConfigPath)
	}
}

func TestValidateRejectsMissingHTTPAddr(t *testing.T) {
	cfg := Default()
	cfg.HTTPAddr = " "

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for missing HTTP address")
	}
}
