package rbac

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/truthwatcher/truthwatcher/internal/authn"
)

const (
	PermissionIntentRead      = "intent:read"
	PermissionIntentWrite     = "intent:write"
	PermissionDeploymentRead  = "deployment:read"
	PermissionDeploymentWrite = "deployment:write"
	PermissionTopologyRead    = "topology:read"
	PermissionTopologyWrite   = "topology:write"
	PermissionReconcileRead   = "reconcile:read"
	PermissionReconcileWrite  = "reconcile:write"
	PermissionAuditRead       = "audit:read"
)

var ErrForbidden = errors.New("forbidden")

type Evaluator interface {
	Evaluate(ctx context.Context, permission string) error
}

type RoleCatalog interface {
	PermissionsForRoles(roles []string) map[string]struct{}
}

type StaticRoleCatalog struct {
	rolePermissions map[string][]string
}

func DefaultRoleCatalog() StaticRoleCatalog {
	return StaticRoleCatalog{rolePermissions: map[string][]string{
		"admin": {
			PermissionIntentRead, PermissionIntentWrite,
			PermissionDeploymentRead, PermissionDeploymentWrite,
			PermissionTopologyRead, PermissionTopologyWrite,
			PermissionReconcileRead, PermissionReconcileWrite,
			PermissionAuditRead,
		},
		"operator": {
			PermissionIntentRead, PermissionIntentWrite,
			PermissionDeploymentRead, PermissionDeploymentWrite,
			PermissionTopologyRead,
			PermissionReconcileRead, PermissionReconcileWrite,
		},
		"viewer": {
			PermissionIntentRead,
			PermissionDeploymentRead,
			PermissionTopologyRead,
			PermissionReconcileRead,
			PermissionAuditRead,
		},
	}}
}

func (c StaticRoleCatalog) PermissionsForRoles(roles []string) map[string]struct{} {
	resolved := make(map[string]struct{})
	for _, role := range roles {
		for _, permission := range c.rolePermissions[role] {
			resolved[permission] = struct{}{}
		}
	}
	return resolved
}

type SimpleEvaluator struct {
	catalog RoleCatalog
}

func NewSimpleEvaluator(catalog RoleCatalog) *SimpleEvaluator {
	if catalog == nil {
		catalog = DefaultRoleCatalog()
	}
	return &SimpleEvaluator{catalog: catalog}
}

func (e *SimpleEvaluator) Evaluate(ctx context.Context, permission string) error {
	claims, ok := authn.ClaimsFromContext(ctx)
	if !ok {
		return fmt.Errorf("%w: missing auth context", ErrForbidden)
	}
	granted := e.catalog.PermissionsForRoles(claims.Roles)
	for _, direct := range claims.Permissions {
		granted[direct] = struct{}{}
	}
	if _, ok := granted[permission]; ok {
		return nil
	}
	return fmt.Errorf("%w: missing permission %q in roles=%v", ErrForbidden, permission, sorted(claims.Roles))
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
