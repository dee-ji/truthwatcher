package policy

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Task is an abstract, vendor-neutral discovery action. Collectors should map
// these tasks to platform-specific read-only commands before execution.
type Task string

const (
	TaskIdentifyDevice   Task = "identify_device"
	TaskGetInventory     Task = "get_inventory"
	TaskGetInterfaces    Task = "get_interfaces"
	TaskGetNeighbors     Task = "get_neighbors"
	TaskGetARP           Task = "get_arp"
	TaskGetIPv6Neighbors Task = "get_ipv6_neighbors"
	TaskGetBGPSummary    Task = "get_bgp_summary"
	TaskGetRoutes        Task = "get_routes"
)

var (
	ErrTaskNotAllowed  = errors.New("policy: task is not allowed")
	ErrCommandDenied   = errors.New("policy: command is denied")
	ErrShellNotAllowed = errors.New("policy: shell access is not allowed")
	ErrCommandRequired = errors.New("policy: command is required")
)

// Action is the minimum policy input future collectors need before execution.
type Action struct {
	Task    Task
	Command string
}

// Engine evaluates read-only discovery policy.
type Engine struct {
	allowedTasks     map[Task]struct{}
	deniedPatterns   []string
	deniedShellParts []string
}

// NewEngine returns the default read-only discovery policy.
func NewEngine() Engine {
	return Engine{
		allowedTasks: map[Task]struct{}{
			TaskIdentifyDevice:   {},
			TaskGetInventory:     {},
			TaskGetInterfaces:    {},
			TaskGetNeighbors:     {},
			TaskGetARP:           {},
			TaskGetIPv6Neighbors: {},
			TaskGetBGPSummary:    {},
			TaskGetRoutes:        {},
		},
		deniedPatterns: []string{
			"configure",
			"commit",
			"delete",
			"reload",
			"clear",
			"write memory",
			"copy",
			"request system reboot",
		},
		deniedShellParts: []string{
			";",
			"&&",
			"||",
			"|",
			"`",
			"$(",
			">",
			"<",
		},
	}
}

// AllowedTasks returns the vendor-neutral tasks this policy permits.
func (e Engine) AllowedTasks() []Task {
	tasks := make([]Task, 0, len(e.allowedTasks))
	for task := range e.allowedTasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// CheckAction validates both the abstract task and concrete command.
func (e Engine) CheckAction(action Action) error {
	if err := e.CheckTask(action.Task); err != nil {
		return err
	}
	return e.CheckCommand(action.Command)
}

// CheckTask rejects unknown abstract discovery tasks.
func (e Engine) CheckTask(task Task) error {
	if task == "" {
		return fmt.Errorf("%w: empty task", ErrTaskNotAllowed)
	}
	if _, ok := e.allowedTasks[task]; !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotAllowed, task)
	}
	return nil
}

// CheckCommand rejects commands known to be unsafe or shell-like.
func (e Engine) CheckCommand(command string) error {
	normalized := normalizeCommand(command)
	if normalized == "" {
		return ErrCommandRequired
	}
	for _, part := range e.deniedShellParts {
		if strings.Contains(command, part) {
			return fmt.Errorf("%w: %q", ErrShellNotAllowed, part)
		}
	}
	for _, pattern := range e.deniedPatterns {
		if commandMatchesPattern(normalized, pattern) {
			return fmt.Errorf("%w: %q", ErrCommandDenied, pattern)
		}
	}
	return nil
}

func normalizeCommand(command string) string {
	return strings.Join(strings.Fields(strings.ToLower(command)), " ")
}

func commandMatchesPattern(command string, pattern string) bool {
	pattern = normalizeCommand(pattern)
	if strings.Contains(pattern, " ") {
		return containsPhrase(command, pattern)
	}
	return containsToken(command, pattern)
}

func containsPhrase(command string, phrase string) bool {
	if command == phrase {
		return true
	}
	return strings.Contains(command, phrase)
}

func containsToken(command string, token string) bool {
	for _, field := range strings.FieldsFunc(command, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
	}) {
		if field == token {
			return true
		}
	}
	return false
}
