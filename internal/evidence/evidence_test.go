package evidence

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeRepository struct {
	createParams CreateEvidenceParams
	createResult Evidence
}

func (f *fakeRepository) CreateEvidence(ctx context.Context, params CreateEvidenceParams) (Evidence, error) {
	f.createParams = params
	return f.createResult, nil
}

func (f *fakeRepository) GetEvidence(ctx context.Context, id string) (Evidence, error) {
	return Evidence{ID: id}, nil
}

func (f *fakeRepository) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]Evidence, error) {
	return []Evidence{{DiscoveryRunID: discoveryRunID}}, nil
}

func TestHashRawOutput(t *testing.T) {
	got := HashRawOutput("show version\n")
	want := "0c9f47ac500773a2715fdaa3a8c0d2ea757b2c783f7130f40ed83c22d842e888"
	if got != want {
		t.Fatalf("hash = %q, want %q", got, want)
	}
}

func TestCreateEvidenceDefaultsMetadata(t *testing.T) {
	repo := &fakeRepository{createResult: Evidence{ID: "evidence-1"}}

	_, err := NewService(repo).CreateEvidence(context.Background(), CreateEvidenceParams{
		DiscoveryRunID: "run-1",
		Target:         " router1 ",
		Method:         " ssh ",
		CommandOrAPI:   " show version ",
		RawOutput:      "raw output",
	})
	if err != nil {
		t.Fatalf("CreateEvidence returned error: %v", err)
	}

	if got, want := string(repo.createParams.Metadata), "{}"; got != want {
		t.Fatalf("Metadata = %q, want %q", got, want)
	}
	if got, want := repo.createParams.Target, "router1"; got != want {
		t.Fatalf("Target = %q, want %q", got, want)
	}
	if got, want := repo.createParams.Method, "ssh"; got != want {
		t.Fatalf("Method = %q, want %q", got, want)
	}
	if got, want := repo.createParams.CommandOrAPI, "show version"; got != want {
		t.Fatalf("CommandOrAPI = %q, want %q", got, want)
	}
}

func TestCreateEvidenceRejectsInvalidMetadata(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateEvidence(context.Background(), CreateEvidenceParams{
		DiscoveryRunID: "run-1",
		Target:         "router1",
		Method:         "ssh",
		CommandOrAPI:   "show version",
		RawOutput:      "raw output",
		Metadata:       json.RawMessage(`{`),
	})
	if err == nil {
		t.Fatal("CreateEvidence returned nil error for invalid metadata")
	}
}

func TestServiceRequiresRepository(t *testing.T) {
	_, err := NewService(nil).GetEvidence(context.Background(), "evidence-1")
	if err == nil {
		t.Fatal("GetEvidence returned nil error without repository")
	}
}
