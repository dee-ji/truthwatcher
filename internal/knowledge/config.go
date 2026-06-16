package knowledge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const DefaultConfigPath = "truthwatcher.yaml"

type Config struct {
	Project   Project
	Providers []Provider
}

type Project struct {
	Name      string
	Repo      string
	LocalPath string
}

type Provider struct {
	Name    string
	Type    string
	Enabled bool
	Root    string
	Repo    string
	Branch  string
	Purpose []string
}

func LoadFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	return Parse(file)
}

func Parse(r io.Reader) (Config, error) {
	var cfg Config
	scanner := bufio.NewScanner(r)

	section := ""
	inProviders := false
	inPurpose := false
	var current *Provider

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		indent := leadingSpaces(raw)
		if indent == 0 {
			section = strings.TrimSuffix(line, ":")
			inProviders = false
			inPurpose = false
			current = nil
			continue
		}

		if section == "knowledge" && indent == 2 && line == "providers:" {
			inProviders = true
			inPurpose = false
			continue
		}

		if inProviders && indent == 4 && strings.HasPrefix(line, "- ") {
			provider := Provider{}
			cfg.Providers = append(cfg.Providers, provider)
			current = &cfg.Providers[len(cfg.Providers)-1]
			inPurpose = false
			key, value, ok := parseKeyValue(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			if ok {
				assignProvider(current, key, value)
			}
			continue
		}

		if inProviders && current != nil {
			if indent == 6 {
				key, value, ok := parseKeyValue(line)
				if !ok {
					return Config{}, fmt.Errorf("invalid provider line %q", line)
				}
				if key == "purpose" && value == "" {
					inPurpose = true
					continue
				}
				inPurpose = false
				assignProvider(current, key, value)
				continue
			}
			if inPurpose && indent == 8 && strings.HasPrefix(line, "- ") {
				current.Purpose = append(current.Purpose, cleanScalar(strings.TrimSpace(strings.TrimPrefix(line, "- "))))
				continue
			}
		}

		if section == "project" && indent == 2 {
			key, value, ok := parseKeyValue(line)
			if !ok {
				return Config{}, fmt.Errorf("invalid project line %q", line)
			}
			assignProject(&cfg.Project, key, value)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func assignProject(project *Project, key, value string) {
	switch key {
	case "name":
		project.Name = cleanScalar(value)
	case "repo":
		project.Repo = cleanScalar(value)
	case "local_path":
		project.LocalPath = cleanScalar(value)
	}
}

func assignProvider(provider *Provider, key, value string) {
	switch key {
	case "name":
		provider.Name = cleanScalar(value)
	case "type":
		provider.Type = cleanScalar(value)
	case "enabled":
		enabled, err := strconv.ParseBool(cleanScalar(value))
		provider.Enabled = err == nil && enabled
	case "root":
		provider.Root = cleanScalar(value)
	case "repo":
		provider.Repo = cleanScalar(value)
	case "branch":
		provider.Branch = cleanScalar(value)
	}
}

func parseKeyValue(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), strings.TrimSpace(value), true
}

func cleanScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}

func leadingSpaces(value string) int {
	count := 0
	for _, r := range value {
		if r != ' ' {
			return count
		}
		count++
	}
	return count
}

var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func ExpandEnv(value string, lookup func(string) (string, bool)) (string, []string) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	var missing []string
	expanded := envPattern.ReplaceAllStringFunc(value, func(match string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if replacement, ok := lookup(name); ok {
			return replacement
		}
		missing = append(missing, name)
		return ""
	})
	return expanded, missing
}
