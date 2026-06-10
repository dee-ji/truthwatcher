package discovery

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"truthwatcher/internal/policy"
)

func TestSSHCollectorConfigValidate(t *testing.T) {
	config := SSHCollectorConfig{
		TargetHost:   "192.0.2.10",
		Port:         22,
		Username:     "readonly",
		PlatformHint: "junos",
		Timeout:      time.Second,
	}

	if err := config.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestSSHCollectorPasswordRequiresEnvironmentValue(t *testing.T) {
	config := SSHCollectorConfig{
		TargetHost: "192.0.2.10",
		Port:       22,
		Username:   "readonly",
	}

	if _, err := config.passwordFromEnvironment(); !errors.Is(err, ErrCredentialRequired) {
		t.Fatalf("expected ErrCredentialRequired, got %v", err)
	}
}

func TestSSHCollectorUsesEnvironmentCredentialReference(t *testing.T) {
	t.Setenv("TRUTHWATCHER_TEST_SSH_PASSWORD", "secret")

	config := SSHCollectorConfig{
		TargetHost:    "192.0.2.10",
		Username:      "readonly",
		CredentialRef: "env://TRUTHWATCHER_TEST_SSH_PASSWORD",
	}

	password, err := config.passwordFromEnvironment()
	if err != nil {
		t.Fatalf("passwordFromEnvironment returned error: %v", err)
	}
	if password != "secret" {
		t.Fatal("passwordFromEnvironment returned unexpected password")
	}
}

func TestSSHCollectorRejectsUnsupportedCredentialReference(t *testing.T) {
	config := SSHCollectorConfig{
		TargetHost:    "192.0.2.10",
		Username:      "readonly",
		CredentialRef: "vault://network/readonly",
	}

	_, err := config.passwordFromEnvironment()
	if !errors.Is(err, ErrUnsupportedCredentialRef) {
		t.Fatalf("expected ErrUnsupportedCredentialRef, got %v", err)
	}
}

func TestSSHCollectorHostKeyCallbackAllowsExplicitInsecureLabMode(t *testing.T) {
	callback, err := hostKeyCallback(SSHCollectorConfig{InsecureIgnoreHostKey: true})
	if err != nil {
		t.Fatalf("hostKeyCallback returned error: %v", err)
	}
	if callback == nil {
		t.Fatal("expected host key callback")
	}
}

func TestSSHCollectorChecksPolicyBeforeDial(t *testing.T) {
	t.Setenv("TRUTHWATCHER_TEST_SSH_PASSWORD", "secret")

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
	collector := NewSSHCollector(SSHCollectorConfig{
		TargetHost:     "192.0.2.10",
		Username:       "readonly",
		PasswordEnvVar: "TRUTHWATCHER_TEST_SSH_PASSWORD",
	}, policy.NewEngine())
	collector.dial = func(ctx context.Context, config SSHCollectorConfig, password string) (sshClient, error) {
		t.Fatal("dial must not be called when policy rejects the profile")
		return nil, nil
	}

	_, err := collector.Collect(context.Background(), "192.0.2.10", profile, []policy.Task{policy.TaskGetRoutes})
	if !errors.Is(err, policy.ErrCommandDenied) {
		t.Fatalf("expected ErrCommandDenied, got %v", err)
	}
}

func TestSSHCollectorCollectsWithMockClient(t *testing.T) {
	t.Setenv("TRUTHWATCHER_TEST_SSH_PASSWORD", "secret")

	profile, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected Junos profile")
	}
	client := &fakeSSHClient{outputs: map[string]string{
		"show version": "Junos: 22.4R3-S2.4\n",
	}}
	collector := NewSSHCollector(SSHCollectorConfig{
		TargetHost:     "192.0.2.10",
		Username:       "readonly",
		PasswordEnvVar: "TRUTHWATCHER_TEST_SSH_PASSWORD",
	}, policy.NewEngine())
	collector.dial = func(ctx context.Context, config SSHCollectorConfig, password string) (sshClient, error) {
		if password != "secret" {
			t.Fatal("unexpected password")
		}
		return client, nil
	}

	outputs, err := collector.Collect(context.Background(), "192.0.2.10", profile, []policy.Task{policy.TaskIdentifyDevice})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if got, want := len(outputs), 1; got != want {
		t.Fatalf("output count = %d, want %d", got, want)
	}
	if outputs[0].Method != SSHMethod {
		t.Fatalf("method = %q, want %q", outputs[0].Method, SSHMethod)
	}
	if outputs[0].Command != "show version" {
		t.Fatalf("command = %q, want show version", outputs[0].Command)
	}
	if client.closed != true {
		t.Fatal("expected SSH client to be closed")
	}
}

func TestSSHCollectorIntegrationPlaceholder(t *testing.T) {
	if os.Getenv("TRUTHWATCHER_RUN_SSH_INTEGRATION") == "" {
		t.Skip("set TRUTHWATCHER_RUN_SSH_INTEGRATION=1 with a lab target to run live SSH collector tests")
	}
	t.Skip("live SSH integration test placeholder; add lab-device assertions when the operator workflow is wired")
}

type fakeSSHClient struct {
	outputs map[string]string
	closed  bool
}

func (f *fakeSSHClient) Run(ctx context.Context, command string) (string, error) {
	output, ok := f.outputs[command]
	if !ok {
		return "", errors.New("unexpected command")
	}
	return output, nil
}

func (f *fakeSSHClient) Close() error {
	f.closed = true
	return nil
}
