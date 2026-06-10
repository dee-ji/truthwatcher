package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	EnvAddr        = "TRUTHWATCHER_ADDR"
	EnvDatabaseURL = "TRUTHWATCHER_DATABASE_URL"
	EnvLogLevel    = "TRUTHWATCHER_LOG_LEVEL"
	EnvDevMode     = "TRUTHWATCHER_DEV_MODE"

	DefaultHTTPAddr = "127.0.0.1:8080"
	DefaultLogLevel = "info"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    string
	DevMode     bool
	ConfigPath  string
}

func Default() Config {
	return Config{
		HTTPAddr: DefaultHTTPAddr,
		LogLevel: DefaultLogLevel,
	}
}

func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

func LoadFromLookup(lookup func(string) (string, bool)) (Config, error) {
	cfg := Default()
	if lookup == nil {
		lookup = func(string) (string, bool) {
			return "", false
		}
	}

	if value, ok := lookup(EnvAddr); ok {
		cfg.HTTPAddr = strings.TrimSpace(value)
	}
	if value, ok := lookup(EnvDatabaseURL); ok {
		cfg.DatabaseURL = strings.TrimSpace(value)
	}
	if value, ok := lookup(EnvLogLevel); ok {
		cfg.LogLevel = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := lookup(EnvDevMode); ok {
		devMode, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return Config{}, fmt.Errorf("%s must be a boolean: %w", EnvDevMode, err)
		}
		cfg.DevMode = devMode
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return fmt.Errorf("http address is required")
	}

	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return fmt.Errorf("http address must be host:port: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(c.LogLevel)) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log level must be one of debug, info, warn, or error")
	}

	return nil
}
