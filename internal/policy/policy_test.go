package policy

import (
	"errors"
	"testing"
)

func TestCheckTaskAllowsKnownTasks(t *testing.T) {
	engine := NewEngine()

	for _, task := range []Task{
		TaskIdentifyDevice,
		TaskGetInventory,
		TaskGetInterfaces,
		TaskGetNeighbors,
		TaskGetARP,
		TaskGetIPv6Neighbors,
		TaskGetBGPSummary,
		TaskGetRoutes,
	} {
		t.Run(string(task), func(t *testing.T) {
			if err := engine.CheckTask(task); err != nil {
				t.Fatalf("expected task to be allowed: %v", err)
			}
		})
	}
}

func TestCheckTaskDeniesUnknownTasks(t *testing.T) {
	engine := NewEngine()

	err := engine.CheckTask(Task("format_disk"))
	if !errors.Is(err, ErrTaskNotAllowed) {
		t.Fatalf("expected ErrTaskNotAllowed, got %v", err)
	}
}

func TestCheckCommandAllowsReadOnlyShowCommands(t *testing.T) {
	engine := NewEngine()

	for _, command := range []string{
		"show version",
		"show inventory",
		"show interfaces brief",
		"show lldp neighbors",
		"show bgp summary",
		"show route summary",
		"show configuration",
	} {
		t.Run(command, func(t *testing.T) {
			if err := engine.CheckCommand(command); err != nil {
				t.Fatalf("expected command to be allowed: %v", err)
			}
		})
	}
}

func TestCheckCommandDeniesDangerousPatterns(t *testing.T) {
	engine := NewEngine()

	tests := map[string]string{
		"configure terminal":                 "configure",
		"conf t":                             "conf t",
		"edit interfaces ge-0/0/0":           "edit",
		"commit confirmed":                   "commit",
		"write terminal":                     "write",
		"delete flash:old.bin":               "delete",
		"remove file old.log":                "remove",
		"reload in 5":                        "reload",
		"reboot":                             "reboot",
		"clear counters":                     "clear",
		"write   memory":                     "write memory",
		"copy running-config startup-config": "copy",
		"scp config user@host:/tmp":          "scp",
		"ftp get image.bin":                  "ftp",
		"tftp get image.bin":                 "tftp",
		"erase startup-config":               "erase",
		"format flash:":                      "format",
		"install add file image.bin":         "install",
		"upgrade system image.bin":           "upgrade",
		"set system host-name router-a":      "set",
		"no shutdown":                        "no",
		"request system reboot":              "request system reboot",
		"REQUEST SYSTEM REBOOT":              "request system reboot",
	}

	for command, pattern := range tests {
		t.Run(command, func(t *testing.T) {
			err := engine.CheckCommand(command)
			if !errors.Is(err, ErrCommandDenied) {
				t.Fatalf("expected ErrCommandDenied for pattern %q, got %v", pattern, err)
			}
		})
	}
}

func TestCheckCommandDeniesShellAccess(t *testing.T) {
	engine := NewEngine()

	for _, command := range []string{
		"show version; reload",
		"show version && reload",
		"show version || reload",
		"show version | include IOS",
		"show version > /tmp/out",
		"show version $(reload)",
		"show version `reload`",
	} {
		t.Run(command, func(t *testing.T) {
			err := engine.CheckCommand(command)
			if !errors.Is(err, ErrShellNotAllowed) {
				t.Fatalf("expected ErrShellNotAllowed, got %v", err)
			}
		})
	}
}

func TestCheckCommandUsesTokenBoundaries(t *testing.T) {
	engine := NewEngine()

	if err := engine.CheckCommand("show commitment history"); err != nil {
		t.Fatalf("expected substring inside another token to be allowed, got %v", err)
	}
}

func TestCheckActionRequiresAllowedTaskAndSafeCommand(t *testing.T) {
	engine := NewEngine()

	if err := engine.CheckAction(Action{Task: TaskGetRoutes, Command: "show route summary"}); err != nil {
		t.Fatalf("expected action to be allowed: %v", err)
	}

	err := engine.CheckAction(Action{Task: TaskGetRoutes, Command: "clear route all"})
	if !errors.Is(err, ErrCommandDenied) {
		t.Fatalf("expected dangerous command to be denied, got %v", err)
	}
}
