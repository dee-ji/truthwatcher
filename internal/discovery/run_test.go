package discovery

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeRepository struct {
	createParams CreateDiscoveryRunParams
	createRun    DiscoveryRun
}

func (f *fakeRepository) CreateDiscoveryRun(ctx context.Context, params CreateDiscoveryRunParams) (DiscoveryRun, error) {
	f.createParams = params
	return f.createRun, nil
}

func (f *fakeRepository) GetDiscoveryRun(ctx context.Context, id string) (DiscoveryRun, error) {
	return DiscoveryRun{ID: id, Status: StatusPending}, nil
}

func (f *fakeRepository) ListDiscoveryRuns(ctx context.Context) ([]DiscoveryRun, error) {
	return []DiscoveryRun{{ID: "run-1", Status: StatusPending}}, nil
}

func (f *fakeRepository) UpdateDiscoveryRunStatus(ctx context.Context, params UpdateDiscoveryRunStatusParams) (DiscoveryRun, error) {
	return DiscoveryRun{ID: params.ID, Status: params.Status}, nil
}

func TestCreateDiscoveryRunDefaultsSeedInput(t *testing.T) {
	repo := &fakeRepository{
		createRun: DiscoveryRun{ID: "run-1", Status: StatusPending},
	}

	_, err := NewService(repo).CreateDiscoveryRun(context.Background(), CreateDiscoveryRunParams{})
	if err != nil {
		t.Fatalf("CreateDiscoveryRun returned error: %v", err)
	}

	if got, want := string(repo.createParams.SeedInput), "{}"; got != want {
		t.Fatalf("SeedInput = %q, want %q", got, want)
	}
}

func TestCreateDiscoveryRunRejectsInvalidSeedInput(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateDiscoveryRun(context.Background(), CreateDiscoveryRunParams{
		SeedInput: json.RawMessage(`{`),
	})
	if err == nil {
		t.Fatal("CreateDiscoveryRun returned nil error for invalid JSON")
	}
}

func TestUpdateDiscoveryRunStatusRejectsInvalidStatus(t *testing.T) {
	_, err := NewService(&fakeRepository{}).UpdateDiscoveryRunStatus(context.Background(), UpdateDiscoveryRunStatusParams{
		ID:     "run-1",
		Status: RunStatus("bad"),
	})
	if err == nil {
		t.Fatal("UpdateDiscoveryRunStatus returned nil error for invalid status")
	}
}

func TestServiceRequiresRepository(t *testing.T) {
	_, err := NewService(nil).ListDiscoveryRuns(context.Background())
	if err == nil {
		t.Fatal("ListDiscoveryRuns returned nil error without repository")
	}
}

func TestNewID(t *testing.T) {
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID returned error: %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("id length = %d, want 36", len(id))
	}
	if id[14] != '4' {
		t.Fatalf("uuid version = %q, want 4", id[14])
	}
}
