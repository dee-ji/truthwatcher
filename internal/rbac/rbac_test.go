package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/authn"
)

func TestSimpleEvaluatorAllowsOperatorIntentWrite(t *testing.T) {
	eval := NewSimpleEvaluator(DefaultRoleCatalog())
	ctx := authn.ContextWithClaims(context.Background(), authn.Claims{Subject: "user-1", Roles: []string{"operator"}})
	if err := eval.Evaluate(ctx, PermissionIntentWrite); err != nil {
		t.Fatalf("expected allowed, got %v", err)
	}
}

func TestSimpleEvaluatorDeniesViewerDeploymentWrite(t *testing.T) {
	eval := NewSimpleEvaluator(DefaultRoleCatalog())
	ctx := authn.ContextWithClaims(context.Background(), authn.Claims{Subject: "user-1", Roles: []string{"viewer"}})
	err := eval.Evaluate(ctx, PermissionDeploymentWrite)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
