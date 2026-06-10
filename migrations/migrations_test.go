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

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "000001_init" {
		t.Fatalf("migration ID = %q, want 000001_init", got[0].ID)
	}
}
