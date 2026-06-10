package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"truthwatcher/internal/policy"
)

const (
	SSHMethod                = "ssh"
	DefaultSSHPort           = 22
	DefaultSSHTimeout        = 10 * time.Second
	DefaultSSHPasswordEnvVar = "TRUTHWATCHER_SSH_PASSWORD"
	DefaultKnownHostsFile    = ".ssh/known_hosts"
	SSHReadOnlyWarning       = "SSH collector is read-only and executes only profile allowlisted commands after policy approval."
)

var (
	ErrCredentialRequired       = errors.New("ssh credential reference or password environment variable is required")
	ErrUnsupportedCredentialRef = errors.New("unsupported ssh credential reference")
	ErrHostKeyVerification      = errors.New("ssh host key verification failed")
)

// SSHCollectorConfig is the boundary between discovery planning and live SSH execution.
type SSHCollectorConfig struct {
	TargetHost            string
	Port                  int
	Username              string
	CredentialRef         string
	PasswordEnvVar        string
	PlatformHint          string
	Timeout               time.Duration
	KnownHostsPath        string
	InsecureIgnoreHostKey bool
}

// Validate rejects incomplete SSH configuration before any connection attempt.
func (c SSHCollectorConfig) Validate() error {
	c = c.normalized()
	if strings.TrimSpace(c.TargetHost) == "" {
		return fmt.Errorf("ssh target host/IP is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("ssh port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.Username) == "" {
		return fmt.Errorf("ssh username is required")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("ssh timeout must be non-negative")
	}
	return nil
}

func (c SSHCollectorConfig) normalized() SSHCollectorConfig {
	c.TargetHost = strings.TrimSpace(c.TargetHost)
	c.Username = strings.TrimSpace(c.Username)
	c.CredentialRef = strings.TrimSpace(c.CredentialRef)
	c.PasswordEnvVar = strings.TrimSpace(c.PasswordEnvVar)
	c.PlatformHint = strings.TrimSpace(c.PlatformHint)
	c.KnownHostsPath = strings.TrimSpace(c.KnownHostsPath)
	if c.Port == 0 {
		c.Port = DefaultSSHPort
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultSSHTimeout
	}
	if c.PasswordEnvVar == "" {
		c.PasswordEnvVar = DefaultSSHPasswordEnvVar
	}
	return c
}

// SSHCollector executes profile allowlisted commands over SSH.
//
// Warning: this collector is read-only by contract. Callers must pass only
// discovery profiles and tasks that have been approved by the policy engine.
type SSHCollector struct {
	Config SSHCollectorConfig
	Policy policy.Engine
	dial   func(ctx context.Context, config SSHCollectorConfig, password string) (sshClient, error)
}

type sshClient interface {
	Run(ctx context.Context, command string) (string, error)
	Close() error
}

// NewSSHCollector creates a live SSH collector. It does not execute anything until Collect is called.
func NewSSHCollector(config SSHCollectorConfig, engine policy.Engine) SSHCollector {
	return SSHCollector{
		Config: config.normalized(),
		Policy: engine,
		dial:   dialSSH,
	}
}

func (c SSHCollector) Collect(ctx context.Context, target string, profile Profile, tasks []policy.Task) ([]CollectedOutput, error) {
	config := c.Config.normalized()
	if strings.TrimSpace(target) != "" && strings.TrimSpace(target) != config.TargetHost {
		return nil, fmt.Errorf("ssh collect target %q does not match configured target %q", target, config.TargetHost)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if err := profile.Validate(c.Policy); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("at least one discovery task is required")
	}

	password, err := config.passwordFromEnvironment()
	if err != nil {
		return nil, err
	}

	dial := c.dial
	if dial == nil {
		dial = dialSSH
	}
	client, err := dial(ctx, config, password)
	if err != nil {
		return nil, err
	}
	defer client.Close()

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
			rawOutput, err := client.Run(ctx, command.Command)
			if err != nil {
				return nil, fmt.Errorf("ssh command %q: %w", command.Command, err)
			}
			outputs = append(outputs, CollectedOutput{
				Target:      config.TargetHost,
				Method:      SSHMethod,
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

func (c SSHCollectorConfig) passwordFromEnvironment() (string, error) {
	envVar := c.PasswordEnvVar
	if envVar == "" {
		envVar = DefaultSSHPasswordEnvVar
	}
	if c.CredentialRef != "" {
		name, ok := envVarFromCredentialRef(c.CredentialRef)
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrUnsupportedCredentialRef, c.CredentialRef)
		}
		envVar = name
	}

	password, ok := os.LookupEnv(envVar)
	if !ok || password == "" {
		return "", fmt.Errorf("%w: %s", ErrCredentialRequired, envVar)
	}
	return password, nil
}

func envVarFromCredentialRef(ref string) (string, bool) {
	ref = strings.TrimSpace(ref)
	for _, prefix := range []string{"env://", "env:"} {
		if strings.HasPrefix(ref, prefix) {
			name := strings.TrimSpace(strings.TrimPrefix(ref, prefix))
			return name, name != ""
		}
	}
	return "", false
}

func dialSSH(ctx context.Context, config SSHCollectorConfig, password string) (sshClient, error) {
	addr := net.JoinHostPort(config.TargetHost, strconv.Itoa(config.Port))
	clientConfig := &ssh.ClientConfig{
		User:    config.Username,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: config.Timeout,
	}
	hostKeyCallback, err := hostKeyCallback(config)
	if err != nil {
		return nil, err
	}
	clientConfig.HostKeyCallback = hostKeyCallback

	dialer := net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial ssh: %w", err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, clientConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake ssh: %w", err)
	}

	return realSSHClient{client: ssh.NewClient(sshConn, chans, reqs)}, nil
}

func hostKeyCallback(config SSHCollectorConfig) (ssh.HostKeyCallback, error) {
	if config.InsecureIgnoreHostKey {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	path := config.KnownHostsPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("%w: resolve home directory: %w", ErrHostKeyVerification, err)
		}
		path = filepath.Join(home, DefaultKnownHostsFile)
	}

	callback, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("%w: load known_hosts %s: %w", ErrHostKeyVerification, path, err)
	}
	return callback, nil
}

type realSSHClient struct {
	client *ssh.Client
}

func (c realSSHClient) Run(ctx context.Context, command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	type result struct {
		output []byte
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		output, err := session.CombinedOutput(command)
		ch <- result{output: output, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-ch:
		return string(result.output), result.err
	}
}

func (c realSSHClient) Close() error {
	return c.client.Close()
}
