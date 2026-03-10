package authn

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Claims struct {
	Subject     string   `json:"sub"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Issuer      string   `json:"iss"`
	Audience    string   `json:"aud"`
}

type ClaimsParser interface {
	Parse(rawToken string) (Claims, error)
}

type Parser struct{}

func NewParser() *Parser { return &Parser{} }

func (p *Parser) Parse(rawToken string) (Claims, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("invalid jwt format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, fmt.Errorf("decode jwt payload: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, fmt.Errorf("parse jwt payload: %w", err)
	}
	if claims.Subject == "" {
		return Claims{}, errors.New("missing subject claim")
	}
	return claims, nil
}

type contextKey string

const ClaimsContextKey contextKey = "authn_claims"

func ContextWithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(Claims)
	return claims, ok
}
