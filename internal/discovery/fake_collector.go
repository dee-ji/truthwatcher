package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"truthwatcher/internal/policy"
)

const (
	FakeMethod         = "fake"
	DefaultFixtureRoot = "examples/fixtures"
)

// DefaultFakeTasks keeps local development focused on fixture-backed evidence.
func DefaultFakeTasks() []policy.Task {
	return []policy.Task{
		policy.TaskIdentifyDevice,
		policy.TaskGetInventory,
		policy.TaskGetNeighbors,
		policy.TaskGetBGPSummary,
	}
}

// InferFakeProfileName maps known fixture targets to built-in profiles.
func InferFakeProfileName(target string) (string, error) {
	fixtureName, err := fixtureNameFromTarget(target)
	if err != nil {
		return "", err
	}

	switch {
	case strings.Contains(fixtureName, "junos") || strings.Contains(fixtureName, "juniper"):
		return ProfileJuniperJunos, nil
	case strings.Contains(fixtureName, "iosxr") || strings.Contains(fixtureName, "cisco"):
		return ProfileCiscoIOSXR, nil
	default:
		return "", fmt.Errorf("could not infer discovery profile from fixture target %q", target)
	}
}

// FakeCollector reads deterministic command outputs from local fixture files.
type FakeCollector struct {
	FixtureRoot string
	Policy      policy.Engine
}

// NewFakeCollector creates a local-only collector for development and tests.
func NewFakeCollector(fixtureRoot string, engine policy.Engine) FakeCollector {
	if strings.TrimSpace(fixtureRoot) == "" {
		fixtureRoot = DefaultFixtureRoot
	}
	return FakeCollector{
		FixtureRoot: fixtureRoot,
		Policy:      engine,
	}
}

func (c FakeCollector) Collect(ctx context.Context, target string, profile Profile, tasks []policy.Task) ([]CollectedOutput, error) {
	fixtureName, err := fixtureNameFromTarget(target)
	if err != nil {
		return nil, err
	}
	if err := profile.Validate(c.Policy); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("at least one discovery task is required")
	}

	var outputs []CollectedOutput
	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		commands, err := profile.CommandsForTask(task)
		if err != nil {
			return nil, err
		}
		for _, command := range commands {
			if err := c.Policy.CheckAction(policy.Action{Task: task, Command: command.Command}); err != nil {
				return nil, err
			}

			rawOutput, err := c.readFixture(fixtureName, command.Command)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, CollectedOutput{
				Target:      target,
				Method:      FakeMethod,
				Task:        task,
				Command:     command.Command,
				RawOutput:   rawOutput,
				ParserHints: append([]string(nil), command.ParserHints...),
				ProfileName: profile.Name,
				Platform:    profile.Platform,
				Vendor:      profile.Vendor,
			})
		}
	}

	return outputs, nil
}

func (c FakeCollector) readFixture(fixtureName string, command string) (string, error) {
	path := filepath.Join(c.FixtureRoot, fixtureName, commandFixtureFilename(command))
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read fake fixture %s: %w", path, err)
	}
	return string(data), nil
}

func fixtureNameFromTarget(target string) (string, error) {
	const prefix = "fixture://"
	if !strings.HasPrefix(target, prefix) {
		return "", fmt.Errorf("fake collector target must use fixture:// scheme")
	}
	name := strings.TrimSpace(strings.TrimPrefix(target, prefix))
	if name == "" {
		return "", fmt.Errorf("fake collector fixture target is required")
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\`) || strings.Contains(name, "..") {
		return "", fmt.Errorf("fake collector fixture target %q is invalid", name)
	}
	return name, nil
}

func commandFixtureFilename(command string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(command) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_") + ".txt"
}
