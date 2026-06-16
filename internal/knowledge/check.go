package knowledge

import (
	"os"
	"strings"
)

const (
	StatusAvailable     = "available"
	StatusMissing       = "missing"
	StatusDisabled      = "disabled"
	StatusMisconfigured = "misconfigured"
)

type CheckResult struct {
	Name     string
	Type     string
	Enabled  bool
	Target   string
	Status   string
	Detail   string
	Provider Provider
}

func CheckProviders(cfg Config, lookup func(string) (string, bool), stat func(string) (os.FileInfo, error)) []CheckResult {
	if lookup == nil {
		lookup = os.LookupEnv
	}
	if stat == nil {
		stat = os.Stat
	}

	results := make([]CheckResult, 0, len(cfg.Providers))
	for _, provider := range cfg.Providers {
		results = append(results, CheckProvider(provider, lookup, stat))
	}
	return results
}

func CheckProvider(provider Provider, lookup func(string) (string, bool), stat func(string) (os.FileInfo, error)) CheckResult {
	result := CheckResult{
		Name:     provider.Name,
		Type:     provider.Type,
		Enabled:  provider.Enabled,
		Provider: provider,
	}

	if strings.TrimSpace(result.Name) == "" || strings.TrimSpace(result.Type) == "" {
		result.Status = StatusMisconfigured
		result.Detail = "provider name and type are required"
		return result
	}

	switch provider.Type {
	case "filesystem":
		resolved, missing := ExpandEnv(provider.Root, lookup)
		result.Target = resolved
		if !provider.Enabled {
			result.Status = StatusDisabled
			return result
		}
		if strings.TrimSpace(provider.Root) == "" {
			result.Status = StatusMisconfigured
			result.Detail = "filesystem provider root is required"
			return result
		}
		if len(missing) > 0 {
			result.Status = StatusMissing
			result.Detail = "missing environment variable: " + strings.Join(missing, ",")
			return result
		}
		info, err := stat(resolved)
		if err != nil {
			result.Status = StatusMissing
			result.Detail = err.Error()
			return result
		}
		if !info.IsDir() {
			result.Status = StatusMisconfigured
			result.Detail = "filesystem provider root is not a directory"
			return result
		}
		result.Status = StatusAvailable
		return result
	case "github":
		result.Target = provider.Repo
		if !provider.Enabled {
			result.Status = StatusDisabled
			return result
		}
		if strings.TrimSpace(provider.Repo) == "" {
			result.Status = StatusMisconfigured
			result.Detail = "github provider repo is required"
			return result
		}
		result.Status = StatusMisconfigured
		result.Detail = "github providers are reserved for future remote workflows and must stay disabled"
		return result
	default:
		result.Status = StatusMisconfigured
		result.Detail = "unknown provider type"
		return result
	}
}
