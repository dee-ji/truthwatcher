package elsecall

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCompilerToJunosFixture(t *testing.T) {
	s := NewCompilerService()
	spec := map[string]any{
		"metadata":         map[string]any{"name": "campus-leaf-1"},
		"routing_intent":   map[string]any{"bgp": map[string]any{"asn": 65000}},
		"target_scope":     map[string]any{"sites": []any{"dc1"}},
		"desired_services": []any{map[string]any{"type": "l3-underlay"}},
		"interface_intent": map[string]any{"uplinks": map[string]any{"speed": "100g"}},
	}
	artifact, err := s.RenderForVendor(context.Background(), "junos", s.BuildIntermediate(spec))
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	fixturePath := filepath.Join("..", "..", "examples", "rendered-configs", "junos-leaf-1.set")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if artifact.Contents != string(fixture) {
		t.Fatalf("artifact did not match fixture\nwant:\n%s\n----\ngot:\n%s", string(fixture), artifact.Contents)
	}
}
