package radiant

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type recordingModule struct {
	initialized *[]string
	name        string
	err         error
}

func (m recordingModule) Initialize(context.Context) error {
	*m.initialized = append(*m.initialized, m.name)
	return m.err
}

func TestServiceStartInitializesConfiguredModulesInOrder(t *testing.T) {
	var calls []string

	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), Dependencies{
		Archive:   recordingModule{initialized: &calls, name: "archive"},
		Ideals:    recordingModule{initialized: &calls, name: "ideals"},
		Elsecall:  recordingModule{initialized: &calls, name: "elsecall"},
		Shadesmar: recordingModule{initialized: &calls, name: "shadesmar"},
	})

	if err := svc.Start(context.Background()); err != nil {
		t.Fatalf("Start() returned unexpected error: %v", err)
	}

	want := []string{"archive", "ideals", "elsecall", "shadesmar"}
	if len(calls) != len(want) {
		t.Fatalf("expected %d module initializations, got %d", len(want), len(calls))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("module call %d: want %q, got %q", i, want[i], calls[i])
		}
	}
}

func TestServiceStartReturnsModuleError(t *testing.T) {
	var calls []string
	initErr := errors.New("boom")

	svc := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), Dependencies{
		Archive: recordingModule{initialized: &calls, name: "archive"},
		Ideals:  recordingModule{initialized: &calls, name: "ideals", err: initErr},
	})

	err := svc.Start(context.Background())
	if err == nil {
		t.Fatal("Start() expected error, got nil")
	}
	if !errors.Is(err, initErr) {
		t.Fatalf("Start() error should wrap initErr")
	}

	want := []string{"archive", "ideals"}
	if len(calls) != len(want) {
		t.Fatalf("expected %d module initializations, got %d", len(want), len(calls))
	}
}
