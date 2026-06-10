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
	"truthwatcher/internal/logging"
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

func serveHTTP(ctx context.Context, cfg config.Config, logger *slog.Logger, stdout io.Writer) error {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	var discoveryRuns *discovery.Service
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		conn, err := db.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer conn.Close()

		service := discovery.NewService(db.NewDiscoveryRunRepository(conn))
		discoveryRuns = &service
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

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  truthwatcher version
  truthwatcher server [--addr 127.0.0.1:8080] [--config ./truthwatcher.yaml]
  truthwatcher migrate up
  truthwatcher migrate status

Commands:
  version   Print the Truthwatcher version.
  server    Start the HTTP server skeleton.
  migrate   Run or inspect database migrations.
`)
}
