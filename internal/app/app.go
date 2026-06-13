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
	"truthwatcher/internal/assets"
	"truthwatcher/internal/config"
	"truthwatcher/internal/db"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/graph"
	"truthwatcher/internal/logging"
	"truthwatcher/internal/parser"
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
	case "parse":
		return a.runParse(ctx, args[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (a App) runVersion(args []string, stdout io.Writer) error {
	if wantsHelp(args) {
		printVersionHelp(stdout)
		return nil
	}
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
	if wantsHelp(args) {
		printServerHelp(stdout)
		return nil
	}

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
	if wantsHelp(args) {
		printMigrateHelp(stdout)
		return nil
	}
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
	if isHelpArg(args[0]) {
		printDiscoverHelp(stdout)
		return nil
	}

	switch args[0] {
	case "fake":
		return a.runDiscoverFake(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("usage: truthwatcher discover fake --target fixture://junos-mx")
	}
}

func (a App) runDiscoverFake(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if wantsHelp(args) {
		printDiscoverFakeHelp(stdout)
		return nil
	}

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

func (a App) runParse(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: truthwatcher parse discovery-run --id <id> --platform junos")
	}
	if isHelpArg(args[0]) {
		printParseHelp(stdout)
		return nil
	}

	switch args[0] {
	case "discovery-run":
		return a.runParseDiscoveryRun(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("usage: truthwatcher parse discovery-run --id <id> --platform junos")
	}
}

func (a App) runParseDiscoveryRun(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if wantsHelp(args) {
		printParseDiscoveryRunHelp(stdout)
		return nil
	}

	loadConfig := a.loadConfig
	if loadConfig == nil {
		loadConfig = config.Load
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var discoveryRunID string
	var platform string
	flags := flag.NewFlagSet("parse discovery-run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&discoveryRunID, "id", "", "discovery run id whose stored evidence should be parsed")
	flags.StringVar(&platform, "platform", "", "parser platform, for example junos or iosxr")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("parse discovery-run accepts flags only")
	}
	if strings.TrimSpace(discoveryRunID) == "" {
		return fmt.Errorf("--id is required")
	}
	if strings.TrimSpace(platform) == "" {
		return fmt.Errorf("--platform is required")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return fmt.Errorf("%s is required for parse discovery-run", config.EnvDatabaseURL)
	}

	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	evidenceService := evidence.NewService(db.NewEvidenceRepository(conn))
	assetService := assets.NewService(db.NewAssetRepository(conn))
	parseService := parser.NewPersistenceService(parser.PersistenceOptions{
		Evidence:     evidenceService,
		Assets:       assetService,
		ParseResults: db.NewParseResultRepository(conn),
		Registry:     parser.BuiltInRegistry(),
	})
	result, err := parseService.ParseDiscoveryRun(ctx, parser.ParseDiscoveryRunParams{
		DiscoveryRunID: discoveryRunID,
		Platform:       platform,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "parsed discovery run %s: %d evidence, %d assets, %d facts, %d relationships, %d warnings\n",
		result.DiscoveryRunID,
		result.EvidenceCount,
		len(result.Assets),
		len(result.Facts),
		len(result.Relationships),
		len(result.Warnings),
	)
	return nil
}

func serveHTTP(ctx context.Context, cfg config.Config, logger *slog.Logger, stdout io.Writer) error {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	var discoveryRuns *discovery.Service
	var evidenceStore *evidence.Service
	var assetStore *assets.Service
	var graphStore *graph.Service
	var parserStore *parser.PersistenceService
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
		assetService := assets.NewService(db.NewAssetRepository(conn))
		assetStore = &assetService
		graphService := graph.NewService(assetService)
		graphStore = &graphService
		parserService := parser.NewPersistenceService(parser.PersistenceOptions{
			Evidence:     evidenceService,
			Assets:       assetService,
			ParseResults: db.NewParseResultRepository(conn),
			Registry:     parser.BuiltInRegistry(),
		})
		parserStore = &parserService
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
			Assets:        assetStore,
			Graph:         graphStore,
			Parser:        parserStore,
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

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher version
  truthwatcher server [--addr 127.0.0.1:8080] [--config ./truthwatcher.yaml]
  truthwatcher migrate up
  truthwatcher migrate status
  truthwatcher discover fake --target fixture://junos-mx
  truthwatcher parse discovery-run --id <id> --platform junos

Commands:
  version   Print the Truthwatcher version.
  server    Start the HTTP server skeleton.
  migrate   Run or inspect database migrations.
  discover  Run a local fixture-backed discovery.
  parse     Persist parser output from stored evidence.

Run "truthwatcher <command> --help" for command details.
`)
}

func printVersionHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher version

Print the Truthwatcher binary version.
`)
}

func printServerHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher server [--addr 127.0.0.1:8080] [--config ./truthwatcher.yaml]

Start the single-binary HTTP server. When TRUTHWATCHER_DATABASE_URL is set,
the server connects to PostgreSQL for API-backed data; without it, the server
still serves health, version, and embedded UI routes.

Flags:
  --addr    HTTP listen address. Overrides TRUTHWATCHER_ADDR.
  --config  Reserved path for future local config-file support.
`)
}

func printMigrateHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher migrate up
  truthwatcher migrate status

Run embedded PostgreSQL migrations compiled into the Truthwatcher binary.
TRUTHWATCHER_DATABASE_URL is required.

Subcommands:
  up      Apply all pending embedded migrations.
  status  Print each embedded migration and whether it is applied.
`)
}

func printDiscoverHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher discover fake --target fixture://junos-mx

Run a discovery workflow. The current packaged workflow is fake fixture-backed
discovery, which stores evidence without touching a network.

Subcommands:
  fake  Collect fixture-backed evidence and store it in PostgreSQL.
`)
}

func printDiscoverFakeHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher discover fake --target fixture://junos-mx [--profile juniper_junos] [--tasks identify_device,get_inventory]

Collect deterministic fixture-backed command outputs, check them through the
read-only policy engine, and store raw evidence before any facts are created.
TRUTHWATCHER_DATABASE_URL is required.

Flags:
  --target    Fixture target, for example fixture://junos-mx or fixture://iosxr-pe.
  --profile   Built-in discovery profile. Inferred from target when omitted.
  --tasks     Comma-separated safe discovery tasks. Defaults to fixture-backed basics.
  --fixtures  Fixture root directory. Defaults to examples/fixtures.
`)
}

func printParseHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher parse discovery-run --id <id> --platform junos

Parse already-stored evidence into assets, facts, and relationships. This does
not run discovery or touch a network.

Subcommands:
  discovery-run  Parse all evidence for a discovery run.
`)
}

func printParseDiscoveryRunHelp(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher parse discovery-run --id <id> --platform junos

Convert stored raw evidence for one discovery run into persisted assets, facts,
and relationships. Raw evidence must already exist; parser warnings are recorded
without deleting or mutating evidence.
TRUTHWATCHER_DATABASE_URL is required.

Flags:
  --id        Discovery run id.
  --platform  Parser platform, for example junos or iosxr.
`)
}
