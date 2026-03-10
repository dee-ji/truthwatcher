package authn

import "fmt"

// Config configures authentication behavior for Spanreed-compatible HTTP APIs.
type Config struct {
	Mode           Mode      `json:"mode" yaml:"mode"`
	JWT            JWTConfig `json:"jwt" yaml:"jwt"`
	LocalDevBypass bool      `json:"local_dev_bypass" yaml:"local_dev_bypass"`
	BypassSubject  string    `json:"bypass_subject" yaml:"bypass_subject"`
	BypassRoles    []string  `json:"bypass_roles" yaml:"bypass_roles"`
}

type Mode string

const (
	ModeDisabled Mode = "disabled"
	ModeJWT      Mode = "jwt"
)

type JWTConfig struct {
	Issuer   string `json:"issuer" yaml:"issuer"`
	Audience string `json:"audience" yaml:"audience"`
}

func (c Config) Validate() error {
	if c.Mode != ModeDisabled && c.Mode != ModeJWT {
		return fmt.Errorf("unsupported auth mode %q", c.Mode)
	}
	if c.Mode == ModeJWT && c.LocalDevBypass && len(c.BypassRoles) == 0 {
		return fmt.Errorf("at least one bypass role is required when local dev bypass is enabled")
	}
	if c.LocalDevBypass && c.BypassSubject == "" {
		return fmt.Errorf("bypass subject is required when local dev bypass is enabled")
	}
	return nil
}
