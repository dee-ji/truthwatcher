package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNewHonorsLogLevel(t *testing.T) {
	var out bytes.Buffer

	logger, err := New(&out, "warn", false)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	logger.InfoContext(context.Background(), "hidden")
	logger.WarnContext(context.Background(), "visible")

	got := out.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("log output = %q, want info message filtered", got)
	}
	if !strings.Contains(got, "visible") {
		t.Fatalf("log output = %q, want warn message", got)
	}
}

func TestNewRejectsInvalidLogLevel(t *testing.T) {
	if _, err := New(&bytes.Buffer{}, "trace", false); err == nil {
		t.Fatal("New returned nil error for invalid log level")
	}
}
