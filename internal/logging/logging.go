package logging

import (
	"io"
	"log/slog"
)

func New(w io.Writer) *slog.Logger {
	if w == nil {
		w = io.Discard
	}
	return slog.New(slog.NewTextHandler(w, nil))
}
