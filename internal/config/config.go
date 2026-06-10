package config

import (
	"fmt"
	"strings"
)

const DefaultHTTPAddr = "127.0.0.1:8080"

type Config struct {
	HTTPAddr   string
	ConfigPath string
}

func Default() Config {
	return Config{
		HTTPAddr: DefaultHTTPAddr,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return fmt.Errorf("http address is required")
	}
	return nil
}
