package intent

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

func TestValidateSpec(t *testing.T) {
	err := validateSpec(map[string]any{"metadata": map[string]any{"name": "leaf-1"}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCreateRejectsInvalidSpec(t *testing.T) {
	svc := NewInMemoryService()
	_, err := svc.Create(context.Background(), domain.Intent{Spec: map[string]any{}})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestCreateAndValidate(t *testing.T) {
	svc := NewInMemoryService()
	in, err := svc.Create(context.Background(), domain.Intent{Name: "leaf", Spec: map[string]any{"metadata": map[string]any{"name": "leaf"}}})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if err := svc.Validate(context.Background(), in.ID); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}
