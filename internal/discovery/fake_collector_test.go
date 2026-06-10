package discovery

import (
	"context"
	"errors"
	"strings"
	"testing"

	"truthwatcher/internal/policy"
)

func TestInferFakeProfileName(t *testing.T) {
	tests := map[string]string{
		"fixture://junos-mx":  ProfileJuniperJunos,
		"fixture://iosxr-asr": ProfileCiscoIOSXR,
	}

	for target, want := range tests {
		t.Run(target, func(t *testing.T) {
			got, err := InferFakeProfileName(target)
			if err != nil {
				t.Fatalf("InferFakeProfileName returned error: %v", err)
			}
			if got != want {
				t.Fatalf("profile = %q, want %q", got, want)
			}
		})
	}
}

func TestFakeCollectorCollectsFixtureOutputs(t *testing.T) {
	profile, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected Junos profile")
	}

	outputs, err := NewFakeCollector("../../examples/fixtures", policy.NewEngine()).Collect(
		context.Background(),
		"fixture://junos-mx",
		profile,
		DefaultFakeTasks(),
	)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if got, want := len(outputs), 4; got != want {
		t.Fatalf("output count = %d, want %d", got, want)
	}

	commands := map[string]bool{}
	for _, output := range outputs {
		if output.Target != "fixture://junos-mx" {
			t.Fatalf("target = %q, want fixture target", output.Target)
		}
		if output.Method != FakeMethod {
			t.Fatalf("method = %q, want %q", output.Method, FakeMethod)
		}
		if output.ProfileName != ProfileJuniperJunos {
			t.Fatalf("profile = %q, want %q", output.ProfileName, ProfileJuniperJunos)
		}
		if strings.TrimSpace(output.RawOutput) == "" {
			t.Fatalf("command %q returned empty output", output.Command)
		}
		commands[output.Command] = true
	}

	for _, command := range []string{"show version", "show chassis hardware", "show lldp neighbors", "show bgp summary"} {
		if !commands[command] {
			t.Fatalf("missing command output for %q", command)
		}
	}
}

func TestFakeCollectorCollectsIOSXRFixtureOutputs(t *testing.T) {
	profile, ok := BuiltInProfile(ProfileCiscoIOSXR)
	if !ok {
		t.Fatal("expected IOS-XR profile")
	}

	outputs, err := NewFakeCollector("../../examples/fixtures", policy.NewEngine()).Collect(
		context.Background(),
		"fixture://iosxr-asr",
		profile,
		DefaultFakeTasks(),
	)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if got, want := len(outputs), 5; got != want {
		t.Fatalf("output count = %d, want %d", got, want)
	}
}

func TestFakeCollectorDeniesUnsafeProfileCommand(t *testing.T) {
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

	_, err := NewFakeCollector("../../examples/fixtures", policy.NewEngine()).Collect(
		context.Background(),
		"fixture://junos-mx",
		profile,
		[]policy.Task{policy.TaskGetRoutes},
	)
	if !errors.Is(err, policy.ErrCommandDenied) {
		t.Fatalf("expected ErrCommandDenied, got %v", err)
	}
}

func TestFakeCollectorRejectsNonFixtureTarget(t *testing.T) {
	profile, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected Junos profile")
	}

	_, err := NewFakeCollector("../../examples/fixtures", policy.NewEngine()).Collect(
		context.Background(),
		"192.0.2.10",
		profile,
		DefaultFakeTasks(),
	)
	if err == nil {
		t.Fatal("expected non-fixture target to fail")
	}
}

func TestCommandFixtureFilename(t *testing.T) {
	if got, want := commandFixtureFilename("show chassis hardware"), "show_chassis_hardware.txt"; got != want {
		t.Fatalf("filename = %q, want %q", got, want)
	}
}
