package migrations

import (
	"testing"
	"testing/fstest"
)

func TestLoadOrdersMigrations(t *testing.T) {
	source := fstest.MapFS{
		"000002_second.up.sql":   {Data: []byte("select 2;")},
		"000002_second.down.sql": {Data: []byte("select -2;")},
		"000001_first.up.sql":    {Data: []byte("select 1;")},
		"000001_first.down.sql":  {Data: []byte("select -1;")},
	}

	got, err := Load(source)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "000001_first" {
		t.Fatalf("first migration = %q, want 000001_first", got[0].ID)
	}
	if got[1].ID != "000002_second" {
		t.Fatalf("second migration = %q, want 000002_second", got[1].ID)
	}
}

func TestLoadRequiresUpAndDown(t *testing.T) {
	source := fstest.MapFS{
		"000001_first.up.sql": {Data: []byte("select 1;")},
	}

	if _, err := Load(source); err == nil {
		t.Fatal("Load returned nil error for incomplete migration")
	}
}

func TestEmbeddedMigrationsLoad(t *testing.T) {
	got, err := Embedded()
	if err != nil {
		t.Fatalf("Embedded returned error: %v", err)
	}

	if len(got) != 11 {
		t.Fatalf("len = %d, want 11", len(got))
	}
	if got[0].ID != "000001_init" {
		t.Fatalf("migration ID = %q, want 000001_init", got[0].ID)
	}
	if got[1].ID != "000002_discovery_runs" {
		t.Fatalf("migration ID = %q, want 000002_discovery_runs", got[1].ID)
	}
	if got[2].ID != "000003_evidence" {
		t.Fatalf("migration ID = %q, want 000003_evidence", got[2].ID)
	}
	if got[3].ID != "000004_assets_facts_relationships" {
		t.Fatalf("migration ID = %q, want 000004_assets_facts_relationships", got[3].ID)
	}
	if got[4].ID != "000005_uncertainty" {
		t.Fatalf("migration ID = %q, want 000005_uncertainty", got[4].ID)
	}
	if got[5].ID != "000006_parser_results" {
		t.Fatalf("migration ID = %q, want 000006_parser_results", got[5].ID)
	}
	if got[6].ID != "000007_audit_records" {
		t.Fatalf("migration ID = %q, want 000007_audit_records", got[6].ID)
	}
	if got[7].ID != "000008_devices" {
		t.Fatalf("migration ID = %q, want 000008_devices", got[7].ID)
	}
	if got[8].ID != "000009_device_registry_v1_fields" {
		t.Fatalf("migration ID = %q, want 000009_device_registry_v1_fields", got[8].ID)
	}
	if got[9].ID != "000010_identity_candidates" {
		t.Fatalf("migration ID = %q, want 000010_identity_candidates", got[9].ID)
	}
	if got[10].ID != "000011_identity_candidate_reviews" {
		t.Fatalf("migration ID = %q, want 000011_identity_candidate_reviews", got[10].ID)
	}
}
