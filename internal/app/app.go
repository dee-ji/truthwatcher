package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"truthwatcher/internal/api"
	"truthwatcher/internal/config"
	"truthwatcher/internal/db"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/logging"
	"truthwatcher/internal/policy"
	"truthwatcher/migrations"
)

const (
	Name    = "truthwatcher"
	Version = "0.1.0-dev"
)

type App struct {
	Version    string
	loadConfig func() (config.Config, error)
	serveHTTP  func(context.Context, config.Config, *slog.Logger, io.Writer) error
}

func New() App {
	return App{
		Version:    Version,
		loadConfig: config.Load,
		serveHTTP:  serveHTTP,
	}
}

func (a App) Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	if len(args) == 0 {
		printUsage(stderr)
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return nil
	case "version":
		return a.runVersion(args[1:], stdout)
	case "server":
		return a.runServer(ctx, args[1:], stdout, stderr)
	case "migrate":
		return a.runMigrate(ctx, args[1:], stdout)
	case "discover":
		return a.runDiscover(ctx, args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (a App) runVersion(args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("version accepts no arguments")
	}

	version := strings.TrimSpace(a.Version)
	if version == "" {
		version = Version
	}

	fmt.Fprintf(stdout, "%s %s\n", Name, version)
	return nil
}

func (a App) runServer(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	loadConfig := a.loadConfig
	if loadConfig == nil {
		loadConfig = config.Load
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&cfg.HTTPAddr, "addr", cfg.HTTPAddr, "HTTP listen address")
	flags.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "path to config file; reserved for the config prompt")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("server accepts flags only")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	logger, err := logging.New(stderr, cfg.LogLevel, cfg.DevMode)
	if err != nil {
		return err
	}

	serve := a.serveHTTP
	if serve == nil {
		serve = serveHTTP
	}

	return serve(ctx, cfg, logger, stdout)
}

func (a App) runMigrate(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: truthwatcher migrate up|status")
	}

	loadConfig := a.loadConfig
	if loadConfig == nil {
		loadConfig = config.Load
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return fmt.Errorf("%s is required for migrate", config.EnvDatabaseURL)
	}

	source, err := migrations.Embedded()
	if err != nil {
		return err
	}

	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	migrator := db.NewMigrator(conn, source)

	switch args[0] {
	case "up":
		ran, err := migrator.Up(ctx)
		if err != nil {
			return err
		}
		if len(ran) == 0 {
			fmt.Fprintln(stdout, "database already up to date")
			return nil
		}
		for _, migration := range ran {
			fmt.Fprintf(stdout, "applied %s\n", migration.ID)
		}
		return nil
	case "status":
		status, err := migrator.Status(ctx)
		if err != nil {
			return err
		}
		for _, item := range status {
			state := "pending"
			if item.Applied {
				state = "applied"
			}
			fmt.Fprintf(stdout, "%s %s\n", item.Migration.ID, state)
		}
		return nil
	default:
		return fmt.Errorf("usage: truthwatcher migrate up|status")
	}
}

func (a App) runDiscover(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: truthwatcher discover fake --target fixture://junos-mx")
	}

	switch args[0] {
	case "fake":
		return a.runDiscoverFake(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("usage: truthwatcher discover fake --target fixture://junos-mx")
	}
}

func (a App) runDiscoverFake(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	loadConfig := a.loadConfig
	if loadConfig == nil {
		loadConfig = config.Load
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var target string
	var profileName string
	var taskList string
	var fixtureRoot string
	flags := flag.NewFlagSet("discover fake", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&target, "target", "", "fixture target, for example fixture://junos-mx")
	flags.StringVar(&profileName, "profile", "", "built-in discovery profile; inferred from fixture target when omitted")
	flags.StringVar(&taskList, "tasks", "", "comma-separated discovery tasks; defaults to fixture-backed identity, inventory, neighbors, and BGP summary")
	flags.StringVar(&fixtureRoot, "fixtures", discovery.DefaultFixtureRoot, "fixture root directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("discover fake accepts flags only")
	}
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("--target is required")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return fmt.Errorf("%s is required for discover fake", config.EnvDatabaseURL)
	}
	if strings.TrimSpace(profileName) == "" {
		profileName, err = discovery.InferFakeProfileName(target)
		if err != nil {
			return err
		}
	}

	profile, ok := discovery.BuiltInProfile(profileName)
	if !ok {
		return fmt.Errorf("unknown discovery profile %q", profileName)
	}

	tasks, err := parseDiscoveryTasks(taskList)
	if err != nil {
		return err
	}

	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	runService := discovery.NewService(db.NewDiscoveryRunRepository(conn))
	evidenceService := evidence.NewService(db.NewEvidenceRepository(conn))

	collector := discovery.NewFakeCollector(fixtureRoot, policy.NewEngine())
	result, err := runService.StartDiscoveryRun(ctx, discovery.StartDiscoveryRunParams{
		Seed: discovery.DiscoverySeed{
			Target: target,
			Method: discovery.FakeMethod,
		},
		Profile:   profile,
		Tasks:     tasks,
		Collector: collector,
		Evidence:  evidenceService,
		Policy:    policy.NewEngine(),
	})
	if err != nil {
		return err
	}

	for _, item := range result.Evidence {
		fmt.Fprintf(stdout, "stored evidence %s %q\n", item.ID, item.CommandOrAPI)
	}

	fmt.Fprintf(stdout, "completed discovery run %s with %d evidence records\n", result.DiscoveryRun.ID, len(result.Evidence))
	return nil
}

func serveHTTP(ctx context.Context, cfg config.Config, logger *slog.Logger, stdout io.Writer) error {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	var discoveryRuns *discovery.Service
	var evidenceStore *evidence.Service
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		conn, err := db.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer conn.Close()

		service := discovery.NewService(db.NewDiscoveryRunRepository(conn))
		discoveryRuns = &service
		evidenceService := evidence.NewService(db.NewEvidenceRepository(conn))
		evidenceStore = &evidenceService
	}

	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler: api.NewHandler(api.Options{
			Version:       Version,
			Logger:        logger,
			DiscoveryRuns: discoveryRuns,
			Evidence:      evidenceStore,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	fmt.Fprintf(stdout, "%s server starting on http://%s\n", Name, listener.Addr().String())
	logger.Info("server starting",
		"addr", listener.Addr().String(),
		"dev_mode", cfg.DevMode,
		"database_configured", cfg.DatabaseURL != "",
	)

	select {
	case <-ctx.Done():
		logger.Info("server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		logger.Info("server stopped")
		return nil
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
}

func parseDiscoveryTasks(taskList string) ([]policy.Task, error) {
	if strings.TrimSpace(taskList) == "" {
		return discovery.DefaultFakeTasks(), nil
	}

	engine := policy.NewEngine()
	parts := strings.Split(taskList, ",")
	tasks := make([]policy.Task, 0, len(parts))
	for _, part := range parts {
		task := policy.Task(strings.TrimSpace(part))
		if err := engine.CheckTask(task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("at least one discovery task is required")
	}
	return tasks, nil
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher version
  truthwatcher server [--addr 127.0.0.1:8080] [--config ./truthwatcher.yaml]
  truthwatcher migrate up
  truthwatcher migrate status
  truthwatcher discover fake --target fixture://junos-mx

Commands:
  version   Print the Truthwatcher version.
  server    Start the HTTP server skeleton.
  migrate   Run or inspect database migrations.
  discover  Run a local fixture-backed discovery.
`)
}
