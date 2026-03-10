package elsecall

import (
	"context"
	"errors"
	"testing"
)

func TestBuildIntermediateAndRenderJunos(t *testing.T) {
	s := NewCompilerService()
	spec := map[string]any{
		"metadata":         map[string]any{"name": "campus-leaf-1"},
		"routing_intent":   map[string]any{"bgp": map[string]any{"asn": 65000}},
		"target_scope":     map[string]any{"sites": []any{"dc1"}},
		"desired_services": []any{map[string]any{"type": "l3-underlay"}},
		"interface_intent": map[string]any{"uplinks": map[string]any{"speed": "100g"}},
	}
	ir := s.BuildIntermediate(spec)
	artifact, err := s.RenderForVendor(context.Background(), "junos", ir)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if artifact.Vendor != "junos" || artifact.Format != "set" {
		t.Fatalf("unexpected artifact metadata: %#v", artifact)
	}
	if artifact.Contents == "" {
		t.Fatalf("expected rendered contents")
	}
}

func TestRenderUnsupportedVendor(t *testing.T) {
	s := NewCompilerService()
	_, err := s.RenderForVendor(context.Background(), "eos", DeviceConfigIR{Hostname: "leaf-1", BGPASN: 65000})
	if !errors.Is(err, ErrUnsupportedVendor) {
		t.Fatalf("expected unsupported vendor error, got %v", err)
	}
}
