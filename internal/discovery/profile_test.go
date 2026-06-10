package discovery

import (
	"errors"
	"testing"

	"truthwatcher/internal/policy"
)

func TestBuiltInProfileExists(t *testing.T) {
	for _, name := range []string{ProfileJuniperJunos, ProfileCiscoIOSXR} {
		t.Run(name, func(t *testing.T) {
			profile, ok := BuiltInProfile(name)
			if !ok {
				t.Fatalf("expected profile %q to exist", name)
			}
			if profile.Name != name {
				t.Fatalf("profile name = %q, want %q", profile.Name, name)
			}
			if profile.Platform == "" {
				t.Fatal("profile platform is empty")
			}
			if profile.Vendor == "" {
				t.Fatal("profile vendor is empty")
			}
		})
	}
}

func TestBuiltInProfileTasksMapToAllowedCommands(t *testing.T) {
	engine := policy.NewEngine()

	for name, profile := range BuiltInProfiles() {
		t.Run(name, func(t *testing.T) {
			if err := profile.Validate(engine); err != nil {
				t.Fatalf("expected profile to validate: %v", err)
			}

			for _, task := range []policy.Task{
				policy.TaskIdentifyDevice,
				policy.TaskGetInventory,
				policy.TaskGetInterfaces,
				policy.TaskGetNeighbors,
				policy.TaskGetARP,
				policy.TaskGetIPv6Neighbors,
				policy.TaskGetBGPSummary,
				policy.TaskGetRoutes,
			} {
				commands, err := profile.CommandsForTask(task)
				if err != nil {
					t.Fatalf("expected task %q to have commands: %v", task, err)
				}
				for _, command := range commands {
					if command.Command == "" {
						t.Fatalf("task %q has empty command", task)
					}
					if len(command.ParserHints) == 0 {
						t.Fatalf("task %q command %q has no parser hints", task, command.Command)
					}
					if err := engine.CheckAction(policy.Action{Task: task, Command: command.Command}); err != nil {
						t.Fatalf("task %q command %q failed policy: %v", task, command.Command, err)
					}
				}
			}
		})
	}
}

func TestProfileValidationDeniesDangerousCommands(t *testing.T) {
	profile := Profile{
		Name:     "unsafe",
		Platform: "test",
		Vendor:   "test",
		Tasks: map[policy.Task][]CommandMapping{
			policy.TaskGetRoutes: {
				{Command: "clear route all", ParserHints: []string{"unsafe"}},
			},
		},
	}

	err := profile.Validate(policy.NewEngine())
	if !errors.Is(err, policy.ErrCommandDenied) {
		t.Fatalf("expected ErrCommandDenied, got %v", err)
	}
}

func TestBuiltInProfilesReturnsCopies(t *testing.T) {
	profiles := BuiltInProfiles()
	profile := profiles[ProfileJuniperJunos]
	profile.Tasks[policy.TaskIdentifyDevice][0].Command = "clear system"

	fresh, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected fresh Junos profile")
	}
	commands, err := fresh.CommandsForTask(policy.TaskIdentifyDevice)
	if err != nil {
		t.Fatalf("CommandsForTask returned error: %v", err)
	}
	if got, want := commands[0].Command, "show version"; got != want {
		t.Fatalf("fresh command = %q, want %q", got, want)
	}
}
