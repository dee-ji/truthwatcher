package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func New(w io.Writer, level string, devMode bool) (*slog.Logger, error) {
	if w == nil {
		w = io.Discard
	}

	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "", "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("unsupported log level %q", level)
	}

	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource: devMode,
		Level:     slogLevel,
	})), nil
}
